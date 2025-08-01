# Kubin Development Makefile

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Kubin Development Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development Environment
.PHONY: dev
dev: dev-up ## Start all development services

.PHONY: dev-up
dev-up: ## Start Docker Compose services
	@echo "🚀 Starting Kubin development environment..."
	docker compose up -d
	@echo "✅ Services started!"
	@echo "📊 PostgreSQL: http://localhost:5432"
	@echo "🔴 Redis: http://localhost:6379"
	@echo "📦 MinIO API: http://localhost:9000"
	@echo "🌐 MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"

.PHONY: dev-down
dev-down: ## Stop Docker Compose services
	@echo "🛑 Stopping development environment..."
	docker compose down

.PHONY: dev-clean
dev-clean: ## Remove Docker Compose containers and volumes
	@echo "🧹 Cleaning up development environment..."
	docker compose down -v --remove-orphans

.PHONY: logs
logs: dev-logs

.PHONY: dev-logs
dev-logs: ## View logs
	@echo "📈 View logs..."
	docker compose logs -f

# Cleanup
.PHONY: clean
clean: dev-down ## Clean up everything
	@echo "🧹 Cleaning up everything..."
	docker compose down -v --remove-orphans
	@echo "✅ Cleanup complete!" 
