#!/bin/bash

own_id=$1
bench_config_file_name="$2"
ready_sync_file_name="$3"
deliver_sync_file_name="$4"
bench_output_file_name="$5"
log_output_file_name="$6"

tmux new-session -d -s bench-node
tmux send-keys "./mir/bin/bench node -i $own_id -c ${bench_config_file_name} --ready-sync-file ${ready_sync_file_name} --deliver-sync-file ${deliver_sync_file_name} --client-stat-file ${bench_output_file_name} > ${log_output_file_name} 2>&1" C-m
