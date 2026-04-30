BINARY            := terraform-provider-clouding
GO                := go
GOLANGCI_LINT     := golangci-lint
DOCKER_IMAGE      := terraform-provider-clouding-dev
GOLANGCI_VERSION  := v1.64.0

.PHONY: build test lint tidy check install dev-install clean \
        docker-build docker-test docker-lint docker-tidy docker-check docker-shell

# ── native targets (require Go + golangci-lint installed locally) ─────────────

build: ## Compile the provider binary
	$(GO) build -o $(BINARY) .

test: ## Run unit tests
	$(GO) test ./...

lint: ## Run golangci-lint
	$(GOLANGCI_LINT) run ./...

tidy: ## Tidy and verify Go module graph
	$(GO) mod tidy
	git diff --exit-code go.mod go.sum

check: tidy build test lint ## Run all checks (tidy + build + test + lint)

install: build ## Install provider binary to GOPATH/bin
	mv $(BINARY) $(GOPATH)/bin/$(BINARY)

dev-install: build ## Install provider binary to the directory used by the dev override
	@echo "Binary built at ./$(BINARY)"
	@echo "Ensure your ~/.terraformrc dev_overrides path matches: $$(pwd)"

clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -rf dist/

# ── Docker targets (no local Go or golangci-lint required) ───────────────────

docker-build: ## Build the dev Docker image (Go + golangci-lint)
	docker build \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_VERSION) \
		--target dev \
		-t $(DOCKER_IMAGE) .

docker-tidy: ## Verify go.mod/go.sum are tidy (Docker)
	docker build --target tidy .

docker-test: ## Run unit tests (Docker)
	docker build --target test .

docker-lint: ## Run golangci-lint (Docker)
	docker build \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_VERSION) \
		--target lint .

docker-check: ## Run all checks: tidy + build + test + lint (Docker)
	docker build \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_VERSION) \
		--target check .

docker-shell: docker-build ## Open an interactive shell in the dev container (source mounted)
	docker run --rm -it \
		-v "$(CURDIR):/src" \
		-w /src \
		$(DOCKER_IMAGE) sh

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
