.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

.PHONY: install-tools
install-tools: ## Install staticcheck.
ifeq ($(shell command -v staticcheck 2> /dev/null),)
	go install honnef.co/go/tools/cmd/staticcheck@latest
endif
ifeq ($(shell command -v goimports 2> /dev/null),)
	go install golang.org/x/tools/cmd/goimports@latest
endif
ifeq ($(shell command -v errcheck 2> /dev/null),)
	go install github.com/kisielk/errcheck@latest
endif
ifeq ($(shell command -v gosec 2> /dev/null),)
	go install github.com/securego/gosec/v2/cmd/gosec@latest
endif
ifeq ($(shell command -v benchstat 2> /dev/null),)
	go install golang.org/x/perf/cmd/benchstat@latest
endif
ifeq ($(shell govulncheck ./... 2> /dev/null),)
	go install golang.org/x/vuln/cmd/govulncheck@latest
endif

.PHONY: gofmt
gofmt: ## Run gofmt.
	test -z "$(gofmt -s -l . | tee /dev/stderr)"

.PHONY: goimports
goimports: ## Run goimports.
	goimports -d $(find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: vet
vet: ## Run vet.
	go vet -v ./...

.PHONY: errcheck
errcheck: ## Run errcheck.
	errcheck -exclude errcheck_excludes.txt ./...

.PHONY: staticcheck
staticcheck: ## Run staticcheck.
	staticcheck ./...

.PHONY: gosec
gosec: ## Run gosec.
	gosec -exclude-dir=_examples ./...

.PHONY: govulncheck
govulncheck: ## Run govulncheck.
	govulncheck ./...

.PHONY: test
test: ## Run tests.
	go test -v -race ./...

.PHONY: test-cover
test-cover: ## Run tests with cover options. ex. make test-cover OUT="c.out"
	go test -v -race -cover -coverprofile=$(OUT) -covermode=atomic ./...

.PHONY: test-benchmark
test-benchmark: ## Run benchmark tests. ex. make test-benchmark CPU=1 COUNT=3
	go test -bench=. -cpu=$(CPU) -benchmem -count=$(COUNT)

.PHONY: test-benchmark-cpuprofile
test-benchmark-cpuprofile: ## Run benchmark tests with cpuprofile and run pprof. ex. Run benchmark tests with cpuprofile and run pprof. ex. make test-benchmark-cpuprofile CPU=1 COUNT=3
	go test -bench . -cpu=$(CPU) -count=$(COUNT) -cpuprofile cpu.out
	go tool pprof -http=":8888" cpu.out

.PHONY: test-benchmark-memprofile
test-benchmark-memprofile: ## Run benchmark tests with memprofile and run pprof. ex. make test-benchmark-memprofile CPU=1 COUNT=3
	go test -bench . -cpu=$(CPU) -count=$(COUNT) -memprofile mem.out
	go tool pprof -http=":8889" mem.out

benchstat: ## Run benchstat
	$(eval BRANCH := $(shell git rev-parse --abbrev-ref HEAD))
	git checkout main
	go test -bench . -benchmem -cpu=1 -count=1 > old.out
	git checkout $(BRANCH)
	go test -bench . -benchmem -cpu=1 -count=1 > new.out
	benchstat old.out new.out