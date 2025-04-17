package main

import (
    "bufio"
    "encoding/json"
    "log"
    "net"
    "sync"
)

type Message struct {
    DestID  string      `json:"dest_id"`
    Payload interface{} `json:"payload"`
}

type Client struct {
    ID   string
    Conn net.Conn
    Send chan []byte
}

var (
    clients   = make(map[string]*Client)
    clientsMu sync.Mutex
)

func handleClient(conn net.Conn) {
    defer conn.Close()
    reader := bufio.NewReader(conn)

    // Read first message: client ID
    idLine, err := reader.ReadBytes('\n')
    if err != nil {
        log.Printf("Failed to read ID: %v", err)
        return
    }
    id := string(idLine[:len(idLine)-1])

    client := &Client{
        ID:   id,
        Conn: conn,
        Send: make(chan []byte, 10),
    }

    clientsMu.Lock()
    clients[id] = client
    clientsMu.Unlock()

    log.Printf("Client registered: %s", id)

    go func() {
        for msg := range client.Send {
            conn.Write(msg)
            conn.Write([]byte("\n"))
        }
    }()

    // Listen for messages
    for {
        line, err := reader.ReadBytes('\n')
        if err != nil {
            log.Printf("Client %s disconnected: %v", id, err)
            break
        }

        var msg Message
        if err := json.Unmarshal(line, &msg); err != nil {
            log.Printf("Invalid JSON from %s: %v", id, err)
            continue
        }

        clientsMu.Lock()
        dest, ok := clients[msg.DestID]
        clientsMu.Unlock()

        if ok {
            encoded, _ := json.Marshal(msg)
            dest.Send <- encoded
        } else {
            log.Printf("No client with ID %s found", msg.DestID)
        }
    }

    // Clean up
    clientsMu.Lock()
    delete(clients, id)
    clientsMu.Unlock()
}

func main() {
    ln, err := net.Listen("tcp", ":9000")
    if err != nil {
        log.Fatal(err)
    }
    defer ln.Close()

    log.Println("Server listening on :9000")
    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Printf("Accept failed: %v", err)
            continue
        }
        go handleClient(conn)
    }
}
