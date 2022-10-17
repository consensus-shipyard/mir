#!/bin/bash

own_id=$1

tmux kill-server
tmux new-session -d -s bench-node
tmux send-keys "./mir/bin/bench node -i $own_id -m membership --statPeriod 1s" C-m
