APP_NAME    := cb
MODULE      := github.com/Morolis/cb
BUILD_DIR   := dist

# Version info from git
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE        := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS     := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

GO_BUILD_FLAGS := -ldflags="$(LDFLAGS)"

.PHONY: all build build-cli build-server frontend test lint clean install help

all: frontend build ## Build everything (frontend + CLI + server)

## --- Build ---

build: build-cli build-server ## Build CLI and server binaries

build-cli: ## Build CLI binary
	go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)"

build-server: frontend ## Build server binary (with embedded frontend)
	go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-server ./server/main.go
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)-server"

frontend: ## Build Vue frontend and copy to server embed path
	@if command -v npm >/dev/null 2>&1 && [ -d "web/node_modules" ]; then \
		echo "Building frontend..."; \
		cd web && npm run build && cd ..; \
		rm -rf server/web/dist; \
		cp -r web/dist server/web/dist; \
		echo "Frontend built: server/web/dist"; \
	else \
		echo "Skipping frontend (npm or node_modules not found)"; \
	fi

## --- Cross-compile ---

cross-compile: frontend ## Cross-compile for all platforms
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .
	CGO_ENABLED=1 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-server ./server/main.go
	@echo ""
	@ls -lh $(BUILD_DIR)/

## --- Test ---

test: ## Run unit tests
	go test ./... -v

test-integration: build ## Run integration tests
	bash test/integration.sh

test-all: test test-integration ## Run all tests

## --- Development ---

lint: ## Run go vet
	go vet ./...

tidy: ## Run go mod tidy
	go mod tidy

run-server: build-server ## Build and run server
	./$(BUILD_DIR)/$(APP_NAME)-server

## --- Install ---

install: ## Install CLI to GOPATH/bin
	go install $(GO_BUILD_FLAGS) .

## --- Clean ---

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
	rm -rf server/web/dist

## --- Release ---

release-check: ## Check goreleaser config
	goreleaser check

release-snapshot: ## Build a snapshot release (local testing)
	goreleaser release --snapshot --clean

release: ## Create a release (requires git tag)
	goreleaser release --clean


help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
