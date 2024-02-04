#!/bin/bash

CONFIG="-numberOfNodes 4 -dropRate 0.05 -minDelay 0.01 -maxDelay 1"

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
  send-keys -t "demo:0.0" "go run . -nodeID 0 $CONFIG 2>&1 | tee \"$NODE_0_LOG\"" Enter \; \
  send-keys -t "demo:0.1" "go run . -nodeID 1 $CONFIG 2>&1 | tee \"$NODE_1_LOG\"" Enter \; \
  send-keys -t "demo:0.2" "go run . -nodeID 2 $CONFIG 2>&1 | tee \"$NODE_2_LOG\"" Enter \; \
  send-keys -t "demo:0.3" "go run . -nodeID 3 $CONFIG 2>&1 | tee \"$NODE_3_LOG\"" Enter \;

sleep 5 # wait for it to start up
for i in {1..15}; do sleep 1;
  tmux send-keys -t "demo:0.0" "0-$i" Enter \;;
  tmux send-keys -t "demo:0.1" "1-$i" Enter \;;
  tmux send-keys -t "demo:0.2" "2-$i" Enter \;;
  tmux send-keys -t "demo:0.3" "3-$i" Enter \;;
done

tmux attach-session -t "demo:0.0"
#!/usr/bin/env bash
set -e