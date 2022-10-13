import yaml
import sys

port = 10000

for idx, a in enumerate(yaml.safe_load(sys.stdin)):
    print("{0} /ip4/{1}/tcp/{2}".format(idx, a, port))
