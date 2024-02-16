.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

CURRENT_BRANCH_GIT_COMMIT_HASH := $(shell git rev-parse HEAD)

.PHONY: up
up: ## Run docker compose up
	cd ../cmd && pwd && GOOS=linux GOARCH=amd64 go build -o ../_examples/proxy/main
	docker compose up --build

.PHONY: down
down: ## Run docker compose down
	docker compose dow

.PHONY: create-certs
create-certs: ## Create certificates
	cd ./proxy/certificates && go run $$(go env GOROOT)/src/crypto/tls/generate_cert.go -rsa-bits 2048 -host localhost