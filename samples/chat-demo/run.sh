NODE_0_LOG="./node_0.log"
NODE_1_LOG="./node_1.log"
NODE_2_LOG="./node_2.log"
NODE_3_LOG="./node_3.log"

rm -rf ./node_*.log

NET="${1:-"libp2p"}"

if [[ "$NET" != "grpc" && "$NET" != "libp2p" ]]; then
  echo "unknown transport $NET"
  exit 1
fi

tmux new-session -d -s "demo" \; \
  new-window   -t "demo" \; \
  \
  split-window -t "demo:0" -v \; \
  split-window -t "demo:0.0" -h \; \
  split-window -t "demo:0.2" -h \; \
  \
  send-keys -t "demo:0.0" "go run ./samples/chat-demo -n $NET 0 2>&1 | tee $NODE_0_LOG" Enter \; \
  send-keys -t "demo:0.1" "go run ./samples/chat-demo -n $NET 1 2>&1 | tee $NODE_1_LOG" Enter \; \
  send-keys -t "demo:0.2" "go run ./samples/chat-demo -n $NET 2 2>&1 | tee $NODE_2_LOG" Enter \; \
  send-keys -t "demo:0.3" "go run ./samples/chat-demo -n $NET 3 2>&1 | tee $NODE_3_LOG" Enter \; \
  attach-session -t "demo:0.0"
