tmux new-session -d -s "demo" \; \
  new-window   -t "demo" \; \
  \
  split-window -t "demo:0" -v \; \
  split-window -t "demo:0.0" -h \; \
  split-window -t "demo:0.2" -h \; \
  \
  send-keys -t "demo:0.0" "go run ./samples/chat-demo 0 " Enter \; \
  send-keys -t "demo:0.1" "go run ./samples/chat-demo 1 " Enter \; \
  send-keys -t "demo:0.2" "go run ./samples/chat-demo 2 " Enter \; \
  send-keys -t "demo:0.3" "go run ./samples/chat-demo 3 " Enter \; \
  attach-session -t "demo:0.0"