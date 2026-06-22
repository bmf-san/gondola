.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

.PHONY: create-certs
create-certs: ## Generate self-signed certificates for the proxy.
	mkdir -p ./proxy/certificates
	cd ./proxy/certificates && go run $$(go env GOROOT)/src/crypto/tls/generate_cert.go -rsa-bits 2048 -host localhost,backend1.local,backend2.local

.PHONY: up
up: ## Build the gondola binary and start the example stack.
	cd .. && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o _examples/proxy/gondola ./cmd
	docker compose up --build

.PHONY: down
down: ## Stop the example stack.
	docker compose down