# Kubin server Makefile

.PHONY: help run build test lint clean

help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the server on :8080
	LOG_LEVEL=debug go run ./cmd

build: ## Build the server binary into dist/
	go build -o dist/kubin-server ./cmd

test: ## Run tests
	go test ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

clean: ## Remove build output
	rm -rf dist
