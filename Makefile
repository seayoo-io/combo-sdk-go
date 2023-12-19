## This is a self-documented Makefile. For usage information, run `make help`:
##
## For more information, refer to https://suva.sh/posts/well-documented-makefiles/

SHELL = bash

GOPATH := $(shell go env GOPATH)

# Respect $GOBIN if set in environment or via $GOENV file.
BIN := $(shell go env GOBIN)
ifndef BIN
BIN := $(GOPATH)/bin
endif

default: help

##@ Environment

.PHONY: deps
deps:  ## Install build and development dependencies
	@echo "==> Updating build dependencies..."
	go install gotest.tools/gotestsum@v1.11.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	@echo "<== Finished updating build dependencies"

##@ Development

# All the .go files, excluding vendor/ and generated files
GO_FILES=$(shell find . -iname '*.go' -type f | grep -v /vendor/ |grep -v ".gen.go"| grep -v ".pb.go")

tidy: ## Tidy up the go mod files
	@echo "==> running go mod tidy"
	@rm -f go.sum
	@go mod tidy -v
	@echo "<== go mod tidy finished"

.PHONY: fmt
fmt: ## Format go source code with gofmt
	@gofmt -l -w $(GO_FILES)

.PHONY: lint
lint: ## Run go linters with golangci-lint
	@$(BIN)/golangci-lint run

# go: -race requires cgo; enable cgo by setting CGO_ENABLED=1
.PHONY: test
test: ## Run unit tests with gotestsum
	@echo "==> Running unit tests with gotestsum..."
	CGO_ENABLED=1 $(BIN)/gotestsum --format=testname -- -cover -race  ./...

.PHONY: build
build: ## Build go source code
	@go build .

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# Print variables helper
# `make print-VARIABLE`
# https://stackoverflow.com/a/25817631
# print-%:;@echo -e $* = $($*) \\n$* origin is $(origin $*)
print-%:
	@echo $* = $($*)
	@echo $* origin is $(origin $*)
