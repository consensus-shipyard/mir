import sys
import json

results = []


def load_data(data):
    result = {"experiment-id": data["experiment-id"]}
    result.update(data["params"])
    result.update(data["aggregates"])
    results.append(result)


def format_value(val):
    if isinstance(val, float):
        return f"{val:>15.4f}"
    else:
        return "{0:>15}".format(val)


def print_results(file):
    if len(results) == 0:
        print("No results found.")
        return

    keys = results[0].keys()
    print(", ".join([format_value(k) for k in keys]), file=file)

    for result in results:
        data = [result[k] for k in keys]
        print(", ".join([format_value(v) for v in data]), file=file)


out_file = sys.argv[1]

for input_file in sys.argv[2:]:
    try:
        print(f"processing: {input_file}")
        with open(input_file, 'r') as file:
            load_data(json.load(file))
    except FileNotFoundError:
        print(f"File '{input_file}' not found.")
        sys.exit(1)
    except json.JSONDecodeError as e:
        print(f"Error decoding JSON file {input_file}: {e}")
        sys.exit(1)

with open(out_file, "w") as file:
    print_results(file)

print("Output written to: "+out_file)
