#!/bin/bash

dir="$1"
settings="$2"
hosts="$3"

scripts/set-up-experiments.sh "$dir" "$settings" "$hosts" || exit 1
scripts/run-experiments.sh "$dir"
