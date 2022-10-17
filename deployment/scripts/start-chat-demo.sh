#!/bin/bash

own_id=$1

tmux kill-server
tmux new-session -d -s chat-demo-node
tmux send-keys "./mir/bin/chat-demo -i membership $own_id 2>&1 | tee chat-demo-$own_id.log" C-m
