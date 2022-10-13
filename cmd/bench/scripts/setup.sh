#!/bin/bash

user=$(whoami)

echo "$1" > firstarg

# Install Go dependencies for code generation
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/golang/mock/mockgen@v1.6.0

# Load .profile (for the PATH variable)
source /home/$user/.profile

# Add go binaries to PATH if not already there
if [[ ":$PATH:" == *":/home/$user/go/bin:"* ]]; then
  echo "PATH already contains /home/$user/go/bin"
else
  echo "PATH=\$PATH:/home/$user/go/bin" >> /home/$user/.profile
fi
