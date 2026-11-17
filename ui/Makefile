# Kubin UI Makefile

.PHONY: help install dev build lint typecheck format

help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	npm ci

dev: ## Dev server on :5173, /api proxied to localhost:8080
	npm run dev

build: ## Build into server/internal/web/dist for go:embed
	npm run build

lint: ## eslint and prettier
	npm run lint

typecheck: ## tsc
	npm run typecheck

format: ## prettier --write
	npm run format
