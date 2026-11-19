# Kubin Development Makefile

# Default target
.PHONY: help image board next task plan usage
help: ## Show this help message
	@echo "Kubin Development Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

image: ## Build the kubin image (UI and server) as kubin:dev
	docker build -t kubin:dev .

board: ## Open the Backlog board; each card you drag to Run gets done, one at a time
	@trap 'kill 0' INT TERM; backlog browser --non-interactive & .agents/run --watch

next: ## Do the next ready task, then stop
	@.agents/run --next

task: ## Do one task, then stop: make task T=15
	@.agents/run $(T)

plan: ## Plan a milestone and tier its tasks: make plan M=m-2
	@.agents/run --plan $(M)

usage: ## Show what each agent session used
	@column -t -s "$$(printf '\t')" .agents/logs/usage.tsv 2>/dev/null || echo "no sessions yet"
