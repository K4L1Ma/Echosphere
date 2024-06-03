SHELL := /bin/bash

# Tools and versions
BUF := github.com/bufbuild/buf/cmd/buf@latest
GOLANGCI_LINT := github.com/golangci/golangci-lint/cmd/golangci-lint@latest
GOTESTFMT := github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest

# Project settings
PKG := ./...
IGNORE := /mocks
COVERAGE_FILE := coverage.out

# Default goal
.DEFAULT_GOAL := all

# Install dependencies
.PHONY: deps
deps:
	go mod tidy
	go mod download

# Generate gRPC code using buf
.PHONY: buf
buf: deps
	go run $(BUF) generate

# Lint code
.PHONY: lint
lint: deps
	go run $(GOLANGCI_LINT) run -c .golangci.yaml --allow-parallel-runners --path-prefix $(PWD)
	go run $(BUF) lint

# Run tests with coverage and format output
.PHONY: test
test: deps
	set -euo pipefail; \
	PACKAGES=$$(go list $(PKG) | grep -v $(IGNORE)); \
	COVERPKG=$$(echo $$PACKAGES | tr ' ' ','); \
	go test -json -v -coverpkg=$$COVERPKG -coverprofile=$(COVERAGE_FILE) $(PKG) -parallel=10 | go run $(GOTESTFMT) -hide empty-packages,successful-downloads

# Generate coverage report
.PHONY: cover
cover: test
	go tool cover -func=$(COVERAGE_FILE)

# Clean up build artifacts and coverage files
.PHONY: clean
clean:
	go clean -cache -testcache -modcache
	rm -f $(COVERAGE_FILE) coverage.html
	rm -rf bin

# Display help message
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  deps     - Install dependencies"
	@echo "  buf      - Generate gRPC code using buf"
	@echo "  lint     - Run linters"
	@echo "  test     - Run tests with coverage"
	@echo "  cover    - Generate coverage report"
	@echo "  clean    - Clean up build artifacts and coverage files"
	@echo "  all      - Install dependencies, generate code, lint, test, and build"

# Default target to install dependencies, generate code, lint, test, and build
.PHONY: all
all: deps buf lint test

