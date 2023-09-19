#!/bin/bash

killall bench
tmux kill-server

sleep 3

killall -9 bench
tmux kill-server

# Some of the above commands inevitably fail, as not all processes are running on all machines.
# This will prevent Ansible (through which this script is expected to be run) from complaining about it.
true
