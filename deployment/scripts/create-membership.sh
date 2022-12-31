#!/bin/bash

hosts=$1

echo "$hosts" | python3 mir/deployment/scripts/hosts-to-membership.py > membership