#!/usr/bin/env bash
set -eu

NODE_0_LOG="./pingpong-0.log"
NODE_1_LOG="./pingpong-1.log"

rm -f ./pingpong-*.log

function quoted {
  SPACE=""
  for arg in "$@"; do
    printf "%s" "$SPACE\"$arg\""
    SPACE=" "
  done
  printf "\n"
}

go build -o pingpong ./samples/pingpong

tmux new-session -d -s "demo" \; \
  new-window   -t "demo" \; \
  \
  split-window -t "demo:0" -v \; \
  \
  send-keys -t "demo:0.0" "./pingpong 0 $(quoted "$@") 2>&1 | tee \"$NODE_0_LOG\"" Enter \; \
  send-keys -t "demo:0.1" "./pingpong 1 $(quoted "$@") 2>&1 | tee \"$NODE_1_LOG\"" Enter \; \
  attach-session -t "demo:0.0"
