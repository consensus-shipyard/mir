import json
import sys


def sort_keys_as_numbers(obj):
    if isinstance(obj, dict):
        return {try_int(k): sort_keys_as_numbers(v) for k, v in obj.items()}
    if isinstance(obj, list):
        return [sort_keys_as_numbers(item) for item in obj]
    return obj


def try_int(s):
    try:
        return int(s)
    except ValueError:
        return s


with open(sys.argv[1], 'r') as file:
    data = json.load(file)

sorted_data = sort_keys_as_numbers(data)

print(json.dumps(sorted_data, indent=4, sort_keys=True))
