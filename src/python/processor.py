# processor.py
import sys
import numpy as np
import json

def process_array(arr):
    np_arr = np.array(arr)
    # You can modify np_arr here if you want to process it
    return np_arr.tolist()

if __name__ == "__main__":
    input_data = json.load(sys.stdin)
    result = process_array(input_data)
    json.dump(result, sys.stdout)
