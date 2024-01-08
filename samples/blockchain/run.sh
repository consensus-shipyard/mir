#!/bin/bash

NODE_0_LOG="./node_0.log"
NODE_1_LOG="./node_1.log"
NODE_2_LOG="./node_2.log"
NODE_3_LOG="./node_3.log"

rm -rf ./node_*.log

tmux kill-session -t demo
tmux new-session -d -s demo \; \
  \
  split-window -t demo:0 -v \; \
  split-window -t demo:0.0 -h \; \
  split-window -t "demo:0.2" -h \; \
  \
  send-keys -t "demo:0.0" "go run .  4 0 2>&1 | tee \"$NODE_0_LOG\"" Enter \; \
  send-keys -t "demo:0.1" "go run .  4 1 2>&1 | tee \"$NODE_1_LOG\"" Enter \; \
  send-keys -t "demo:0.2" "go run .  4 2 2>&1 | tee \"$NODE_2_LOG\"" Enter \; \
  send-keys -t "demo:0.3" "go run .  4 3 2>&1 | tee \"$NODE_3_LOG\"" Enter \; \
  attach-session -t "demo:0.0"
#!/usr/bin/env bash
set -e
