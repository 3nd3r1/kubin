# Kubin CLI Makefile

.PHONY: help dev build test lint clean

help: ## Show this help message
	@echo "Kubin CLI Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Run CLI in development mode
	@echo "🔧 Running CLI..."
	go run main.go

build: ## Build CLI binary
	@echo "🔨 Building CLI..."
	go build -ldflags "-X github.com/3nd3r1/kubin/cli/cmd.version=$$(git describe --tags --always --dirty) -X github.com/3nd3r1/kubin/cli/cmd.commit=$$(git rev-parse --short=12 HEAD)" -o kubin-cli main.go

test: ## Run tests
	@echo "🧪 Running tests..."
	go test ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -f kubin-cli
