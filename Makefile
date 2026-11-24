# Kubin Development Makefile

# Default target
.PHONY: help image check
help: ## Show this help message
	@echo "Kubin Development Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

image: ## Build the kubin image (UI and server) as kubin:dev
	docker build -t kubin:dev .

check: ## Lint and test everything: cli, server, ui (run `make -C ui install` first)
	$(MAKE) -C cli lint
	$(MAKE) -C cli test
	$(MAKE) -C server lint
	$(MAKE) -C server test
	$(MAKE) -C ui lint
	$(MAKE) -C ui typecheck
	$(MAKE) -C ui test
	$(MAKE) -C ui build
