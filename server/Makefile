# Kubin Server Makefile

.PHONY: help
help: ## Show this help message
	@echo "Kubin Server Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev
dev: ## Start server in development mode
	@echo "🚀 Starting server..."
	go run cmd/server/main.go

.PHONY: build
build: ## Build server binary
	@echo "🔨 Building server..."
	go build -o kubin-server cmd/server/main.go

.PHONY: test
test: ## Run tests
	@echo "🧪 Running tests..."
	go test ./...

.PHONY: run
run: build ## Build and run server
	@echo "🚀 Running server..."
	./kubin-server

.PHONY: format
format: ## Format code
	@echo "🧹 Formatting code..."
	go fmt ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -f kubin-server
