# Copyright Protocol Labs. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

targetdir := bin

all: test samples

.PHONY: help
help:
	@echo 'Usage: make [target]...'
	@echo ''
	@echo 'Generic targets:'
	@echo '  all (default)    - Build and test all'
	@echo '  clean            - Remove all build and test artifacts'
	@echo '  test             - Run all tests'
	@echo '  lint             - Run code quality checks'
	@echo '  generate         - Generate dependent files'

.PHONY: clean
clean:
	rm -rf mir-deployment-test failed-test-data $(targetdir)

.PHONY: test
test:
	go test -race ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: generate
generate:
	go generate ./...

samples: $(targetdir)/chat-demo

$(targetdir)/chat-demo:
	go build -o $(targetdir)/chat-demo ./samples/chat-demo
