#!/bin/bash

source scripts/global-vars.sh

dir="$1"
settings="$2"

# If no hosts file has been given, use the default global variable.
if [ -n "$3" ]; then
    hosts="$3"
else
    hosts="$ansible_hosts_file"
fi

num_hosts=$(wc -l < "$hosts")

mkdir "$dir" || exit 1
cp "$settings" "$dir/parameter-set.yaml" || exit 1
cp "$hosts" "$dir/$ansible_hosts_file" || exit 1

python3 scripts/hosts-to-membership.py < "$dir/$ansible_hosts_file" > "$dir/membership.json" || exit 1
go run ../cmd/bench params -m "$dir/membership.json" -o "$node_config_file_name" -s "$settings" -d "$dir" -w 4 || exit 1

ansible-playbook -i "$dir/$ansible_hosts_file" --forks "$num_hosts" setup.yaml || exit 1
