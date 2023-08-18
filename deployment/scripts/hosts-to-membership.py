# This script generates a Trantor membership file from a list IP addresses read from standard input.
#
# Configuration number of the membership is always set to 0.

import sys
import json
import yaml

membership = {
    "configuration_number": 0,
    "validators": [],
}


def parse_validator(line: str):
    tokens = line.split("@")
    membership["validators"].append({
        "addr": tokens[0],
        "net_addr": tokens[1],
        "weight": "0",
    })


port = 10000
node_id = 0
# for node_id, ip_address in enumerate(yaml.safe_load(sys.stdin)): # Used if input is provided as a yaml list
for ip_address in sys.stdin:
    ip_address = ip_address.strip()
    if ip_address == "":  # Skip empty lines
        continue

    membership["validators"].append({
        "addr": "{0}".format(node_id),
        "net_addr": "/ip4/{0}/tcp/{1}".format(ip_address, port),
        "weight": "1",
    })
    node_id += 1

# Printing the output of json.dumps instead of using directly json.dump to stdout,
# since the latter seems to append extra characters to the output.
print(json.dumps(membership, indent=2))
