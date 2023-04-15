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
	go test -v -shuffle=on -count=1 -race -timeout 20m ./...

.PHONY: test_cov
test_cov:
	go test -coverprofile coverage.out -v -count=1 -timeout 20m ./...

.PHONY: format
format:
	gofmt -w -s .
	goimports -w -local "github.com/filecoin-project/mir" .

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fuzz
fuzz:
	./fuzz.sh

.PHONY: vulncheck
vulncheck:
	govulncheck -v ./...

.PHONY: generate
generate:
	rm -rf pkg/pb/*
	go generate ./...

cmd: $(targetdir)/bench $(targetdir)/mircat

.PHONY: $(targetdir)/bench
$(targetdir)/bench:
	go build -o $(targetdir)/bench ./cmd/bench

.PHONY: $(targetdir)/mircat
$(targetdir)/mircat:
	go build -o $(targetdir)/mircat ./cmd/mircat


samples: $(targetdir)/chat-demo $(targetdir)/availability-layer-demo $(targetdir)/bcb-demo $(targetdir)/pingpong

.PHONY: $(targetdir)/chat-demo
$(targetdir)/chat-demo:
	go build -o $(targetdir)/chat-demo ./samples/chat-demo

.PHONY: $(targetdir)/availability-layer-demo
$(targetdir)/availability-layer-demo:
	go build -o $(targetdir)/availability-layer-demo ./samples/availability-layer-demo

.PHONY: $(targetdir)/bcb-demo
$(targetdir)/bcb-demo:
	go build -o $(targetdir)/bcb-demo ./samples/bcb-demo

.PHONY: $(targetdir)/pingpong
$(targetdir)/pingpong:
	go build -o $(targetdir)/pingpong ./samples/pingpong

