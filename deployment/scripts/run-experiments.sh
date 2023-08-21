#!/bin/bash

source scripts/global-vars.sh

dir="$1"

num_hosts=$(wc -l < "$dir/$ansible_hosts_file" || exit 1)
if [[ -z "$num_hosts" || "$num_hosts" -eq 0 ]]; then
  echo "No hosts specified" >&2
  exit 1
fi

for exp_dir in "$dir"/[0-9][0-9][0-9][0-9]; do
  exp_id=$(basename "$exp_dir")

  if [[ -f "$exp_dir/results.json" ]]; then
    echo "Skipping experiment $exp_id. File already exists: $exp_dir/results.json"
  else
    # Run the experiment and download its raw results.
    ansible-playbook -i "$dir/hosts" --forks "$num_hosts" --extra-vars "node_config_file=$exp_dir/$node_config_file_name exp_id=$exp_id output_dir='$exp_dir' bench_output_file=$bench_output_file" run-benchmark.yaml || touch "${exp_dir}/FAILED"

    # Process the results of the experiment.
    python3 scripts/analyze-bench-output.py "$exp_id" "$exp_dir/$node_config_file_name" "$exp_dir/results.json" "$exp_dir"/data/*$node_output_file.json

    # Aggregate data from this and all previous experiments in a single table.
    python3 scripts/aggregate-data.py "$dir"/partial-results.csv "$dir"/[0-9][0-9][0-9][0-9]/results.json > /dev/null
  fi
done

# Aggregate data from all experiments in a single final table.
python3 scripts/aggregate-data.py "$dir"/all-results.csv "$dir"/[0-9][0-9][0-9][0-9]/results.json || exit 1
