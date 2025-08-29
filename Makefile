# Kubin Development Makefile

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Kubin Development Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
