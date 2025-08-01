# Kubin CLI Makefile

.PHONY: help dev build test clean

help: ## Show this help message
	@echo "Kubin CLI Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Run CLI in development mode
	@echo "🔧 Running CLI..."
	go run main.go

build: ## Build CLI binary
	@echo "🔨 Building CLI..."
	go build -o kubin-cli main.go

test: ## Run tests
	@echo "🧪 Running tests..."
	go test ./...

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -f kubin-cli
