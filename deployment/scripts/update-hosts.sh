#!/bin/bash

source scripts/global-vars.sh

dir="$1"
# If no hosts file has been given, use the default global variable.
if [ -n "$2" ]; then
    hosts="$2"
else
    hosts="$ansible_hosts_file"
fi

# Update the hosts file itself and membership file
cp "$hosts" "$dir/$ansible_hosts_file" || exit 1
python3 scripts/hosts-to-membership.py < "$dir/$ansible_hosts_file" > "$dir/membership.json" || exit 1

# Compile bench binary, as it may be called many times

go build -o ../bench ../cmd/bench

# Update the configuration file of each experiments to use the new membership
for exp_dir in "$dir"/[0-9][0-9][0-9][0-9]; do
  exp_id=$(basename "$exp_dir")

  if [[ -f "$exp_dir/results.json" ]]; then
    echo "Skipping experiment $exp_id. File already exists: $exp_dir/results.json"
  else
    echo "Updating experiment membership: $exp_dir."
    ../bench params -i "$exp_dir/$node_config_file_name" -m "$dir/membership.json" -o "$exp_dir/$node_config_file_name" || exit 1
  fi
done
