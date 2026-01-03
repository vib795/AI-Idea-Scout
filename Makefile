.PHONY: build test lint vuln clean install help

# Build variables
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building ideascout..."
	@go build -tags fts5 -ldflags "$(LDFLAGS)" -o bin/ideascout ./cmd/ideascout

install: build ## Install the binary to $(GOPATH)/bin
	@echo "Installing ideascout..."
	@cp bin/ideascout $(GOPATH)/bin/

test: ## Run tests
	@echo "Running tests..."
	@go test -tags fts5 -v -race -cover ./...

lint: ## Run linters
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout=5m

vuln: ## Check for vulnerabilities
	@echo "Checking for vulnerabilities..."
	@govulncheck ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ dist/

tidy: ## Tidy go.mod
	@go mod tidy

fmt: ## Format code
	@go fmt ./...

.DEFAULT_GOAL := help
