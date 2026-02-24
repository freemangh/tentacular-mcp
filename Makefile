.PHONY: build test test-unit test-integration test-e2e test-all lint clean

BINARY := bin/tentacular-mcp
MODULE := github.com/randybias/tentacular-mcp
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build: ## Build the binary
	go build $(LDFLAGS) -o $(BINARY) ./cmd/tentacular-mcp

test: test-unit ## Run unit tests (default)

test-unit: ## Unit tests (no cluster)
	go test ./pkg/... -v -count=1

test-integration: ## Integration tests (kind cluster)
	@test/integration/setup_kind.sh
	go test -tags=integration ./test/integration/... -v -timeout 300s -count=1
	@test/integration/teardown_kind.sh

test-e2e: ## E2E tests (production k0s cluster)
	@test -n "$$TENTACULAR_E2E_KUBECONFIG" || { echo "Set TENTACULAR_E2E_KUBECONFIG"; exit 1; }
	go test -tags=e2e ./test/e2e/... -v -timeout 600s -count=1

test-all: test-unit test-integration test-e2e ## Run all test tiers

lint: ## Run linters
	golangci-lint run ./...
	go vet ./...

clean: ## Clean build artifacts
	rm -rf bin/

docker-build: ## Build Docker image
	docker build -t tentacular-mcp:$(VERSION) .

deploy: ## Deploy to current cluster
	kubectl apply -k deploy/manifests/

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
