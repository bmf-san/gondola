.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

CURRENT_BRANCH_GIT_COMMIT_HASH := $(shell git rev-parse HEAD)

# TODO:
# This command was created to allow access to the latest commits during development.
# Once development has stabilized, it may be better to specify a release tag.
.PHONY: update
update: ## Update dependencies
	cd ./proxy && go mod edit -require=github.com/bmf-san/gondola@$(CURRENT_BRANCH_GIT_COMMIT_HASH) && go mod tidy

.PHONY: up
up: ## Run docker compose up
	docker compose up

.PHONY: down
down: ## Run docker compose down
	docker compose down