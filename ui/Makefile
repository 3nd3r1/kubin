# Kubin UI Makefile

.PHONY: help
help: ## Show this help message
	@echo "Kubin UI Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev
dev: ## Start development server
	@echo "🎨 Starting UI development server..."
	npm run dev

.PHONY: build
build: ## Build for production
	@echo "🔨 Building UI..."
	npm run build

.PHONY: test
test: ## Run tests
	@echo "🧪 Running tests..."
	npm test

.PHONY: install
install: ## Install dependencies
	@echo "📦 Installing dependencies..."
	npm install

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf .next
	rm -rf dist 
