.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

CURRENT_BRANCH_GIT_COMMIT_HASH := $(shell git rev-parse HEAD)

.PHONY: up
up: ## Run docker compose up
	cd ../cmd && pwd && GOOS=linux GOARCH=amd64 go build -o ../_examples/proxy/main
	docker compose up

.PHONY: down
down: ## Run docker compose down
	docker compose down