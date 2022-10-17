#!/bin/bash

hosts=$1

echo "$hosts" | python3 mir/cmd/bench/scripts/hosts-to-membership.py > membership