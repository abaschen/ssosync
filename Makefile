OUTPUT = main # Referenced as Handler in sar-template.json

PACKAGED_TEMPLATE = packaged.yaml
STACK_NAME := $(STACK_NAME)
S3_BUCKET := $(S3_BUCKET)
S3_PREFIX := $(S3_PREFIX)
TEMPLATE = sar-template.json
APP_NAME ?= ssosync
GOREL_ARGS ?= 
GOREL ?= goreleaser

# Tool versions
MOCKERY_VERSION ?= v3.5.2
GOLANGCI_LINT_VERSION ?= v2.3.1
GORELEASER_VERSION ?= v2.11.2
UPX_VERSION ?= v4.2.4

# Tool installation paths
TOOLS_DIR := $(shell pwd)/.bin
MOCKERY := $(TOOLS_DIR)/mockery
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint
GORELEASER_BIN := $(TOOLS_DIR)/goreleaser
UPX := $(TOOLS_DIR)/upx

# Detect OS and architecture
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
	ARCH := amd64
endif
ifeq ($(ARCH),aarch64)
	ARCH := arm64
endif

.PHONY: install-deps
install-deps: install-mockery install-golangci-lint install-goreleaser install-upx
	@echo "All development dependencies installed"

.PHONY: install-mockery
install-mockery:
	@echo "Installing mockery $(MOCKERY_VERSION)..."
	@mkdir -p $(TOOLS_DIR)
	@if [ ! -f $(MOCKERY) ] || [ "$$($(MOCKERY) version 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+')" != "$(MOCKERY_VERSION)" ]; then \
		MOCKERY_ARCH=$(ARCH); \
		MOCKERY_OS=$(OS); \
		if [ "$(ARCH)" = "amd64" ]; then MOCKERY_ARCH="x86_64"; fi; \
		if [ "$(OS)" = "darwin" ]; then MOCKERY_OS="Darwin"; fi; \
		if [ "$(OS)" = "linux" ]; then MOCKERY_OS="Linux"; fi; \
		curl -sSfL https://github.com/vektra/mockery/releases/download/$(MOCKERY_VERSION)/mockery_$(MOCKERY_VERSION:v%=%)_$${MOCKERY_OS}_$${MOCKERY_ARCH}.tar.gz | tar -xz -C $(TOOLS_DIR) mockery; \
		chmod +x $(MOCKERY); \
		echo "mockery $(MOCKERY_VERSION) installed"; \
	else \
		echo "mockery $(MOCKERY_VERSION) already installed"; \
	fi

.PHONY: install-golangci-lint
install-golangci-lint:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@mkdir -p $(TOOLS_DIR)
	@if [ ! -f $(GOLANGCI_LINT) ] || [ "$$($(GOLANGCI_LINT) --version 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+')" != "$(GOLANGCI_LINT_VERSION)" ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DIR) $(GOLANGCI_LINT_VERSION); \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) installed"; \
	else \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) already installed"; \
	fi

.PHONY: install-goreleaser
install-goreleaser:
	@echo "Installing goreleaser $(GORELEASER_VERSION)..."
	@mkdir -p $(TOOLS_DIR)
	@if [ ! -f $(GORELEASER_BIN) ] || [ "$$($(GORELEASER_BIN) --version 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+')" != "$(GORELEASER_VERSION)" ]; then \
		curl -sSfL https://github.com/goreleaser/goreleaser/releases/download/$(GORELEASER_VERSION)/goreleaser_$(OS)_$(ARCH).tar.gz | tar -xz -C $(TOOLS_DIR) goreleaser; \
		chmod +x $(GORELEASER_BIN); \
		echo "goreleaser $(GORELEASER_VERSION) installed"; \
	else \
		echo "goreleaser $(GORELEASER_VERSION) already installed"; \
	fi

.PHONY: install-upx
install-upx:
	@echo "Installing upx $(UPX_VERSION)..."
	@mkdir -p $(TOOLS_DIR)
	@if [ ! -f $(UPX) ] || [ "$$($(UPX) --version 2>/dev/null | head -1 | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+')" != "$(UPX_VERSION)" ]; then \
		UPX_ARCH=$(ARCH); \
		if [ "$(ARCH)" = "amd64" ]; then UPX_ARCH="amd64"; fi; \
		if [ "$(ARCH)" = "arm64" ]; then UPX_ARCH="arm64"; fi; \
		if [ "$(OS)" = "linux" ]; then \
			UPX_FILE="upx-$(UPX_VERSION:v%=%)-$${UPX_ARCH}_linux"; \
		elif [ "$(OS)" = "darwin" ]; then \
			UPX_FILE="upx-$(UPX_VERSION:v%=%)-amd64_macos"; \
		else \
			echo "Error: Unsupported OS $(OS) for UPX installation"; \
			exit 1; \
		fi; \
		curl -sSfL https://github.com/upx/upx/releases/download/$(UPX_VERSION)/$${UPX_FILE}.tar.xz | tar -xJ -C $(TOOLS_DIR) --strip-components=1 $${UPX_FILE}/upx; \
		chmod +x $(UPX); \
		echo "upx $(UPX_VERSION) installed"; \
	else \
		echo "upx $(UPX_VERSION) already installed"; \
	fi

.PHONY: generate-mock
generate-mock: install-mockery
	$(MOCKERY)

.PHONY: test
test: generate-mock
	go test ./... -coverprofile=coverage.out

.PHONY: test-verbose
test-verbose: generate-mock
	go test -v ./... -coverprofile=coverage.out

.PHONY: test-coverage
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: go-build
go-build: install-goreleaser
	$(GORELEASER_BIN) build --snapshot --clean --id ssosync $(GOREL_ARGS)

.PHONY: clean
clean:
	rm -f $(OUTPUT) $(PACKAGED_TEMPLATE) bootstrap coverage.out coverage.html
	rm -rf dist/ internal/mocks/*

.PHONY: clean-all
clean-all: clean
	rm -rf $(TOOLS_DIR)

build-SSOSyncFunction: go-build
	cp dist/ssosync_linux_arm64_v8.2/ssosync $(ARTIFACTS_DIR)/bootstrap

.PHONY: config
config:
	go mod download

.PHONY: vet
vet: install-golangci-lint generate-mock
	$(GOLANGCI_LINT) run

.PHONY: lint
lint: vet

.PHONY: fmt
fmt:
	go fmt ./...
	go mod tidy

main: main.go install-goreleaser
	echo $(GORELEASER_BIN) build --clean $(GOREL_ARGS)
	$(GORELEASER_BIN) build --clean $(GOREL_ARGS)

# compile the code to run in Lambda (local or real)
.PHONY: lambda
lambda: main

.PHONY: build
build: clean main

.PHONY: release
release: install-goreleaser
	$(GORELEASER_BIN) release --clean $(GOREL_ARGS)

.PHONY: dry-run
dry-run: 
	$(MAKE) GOREL_ARGS=--skip=publish release

.PHONY: api
api: build
	sam local start-api

.PHONY: publish
publish:
	sam publish -t packaged.yaml

.PHONY: package
package: build
	@if [ "$(ARCH)" = "arm64" ]; then \
		cp dist/ssosync_$(OS)_arm64_v8.2/ssosync ./bootstrap; \
	elif [ "$(ARCH)" = "amd64" ]; then \
		cp dist/ssosync_$(OS)_amd64_v1/ssosync ./bootstrap; \
	else \
		echo "Error: Unsupported architecture $(ARCH)"; \
		exit 1; \
	fi
	sam package --s3-bucket $(S3_BUCKET) --output-template-file $(PACKAGED_TEMPLATE) --s3-prefix $(S3_PREFIX)

.PHONY: deploy
deploy: package
	sam deploy --stack-name $(STACK_NAME) --template-file $(PACKAGED_TEMPLATE) --capabilities CAPABILITY_IAM

.PHONY: dev
dev: go-build
	@echo "Running development build for $(OS)/$(ARCH)..."
	@if [ "$(ARCH)" = "arm64" ]; then \
		BINARY_PATH="dist/ssosync_$(OS)_arm64_v8.2/ssosync"; \
	elif [ "$(ARCH)" = "amd64" ]; then \
		BINARY_PATH="dist/ssosync_$(OS)_amd64_v1/ssosync"; \
	else \
		BINARY_PATH=$$(find dist/ -name "ssosync_$(OS)_*" -type f | head -1); \
	fi; \
	if [ ! -f "$$BINARY_PATH" ]; then \
		echo "Error: Binary not found at $$BINARY_PATH. Available binaries:"; \
		find dist/ -name "ssosync*" -type f || echo "No binaries found in dist/"; \
		exit 1; \
	fi; \
	echo "Using binary: $$BINARY_PATH"; \
	$$BINARY_PATH -g "name:AWS*" \
		-t $$(jq '.["ssosync/aws-sso/scimEndpointAccessToken"]' ./cicd/cloudformation/cdk/cdk.context.json -r) \
		-r $$(jq '.["ssosync/aws-sso/region"]' ./cicd/cloudformation/cdk/cdk.context.json -r) \
		-e $$(jq '.["ssosync/aws-sso/scimEndpointUrl"]' ./cicd/cloudformation/cdk/cdk.context.json -r) \
		-u $$(jq '.["ssosync/secrets/googleAdminEmail"]' ./cicd/cloudformation/cdk/cdk.context.json -r) \
		-i $$(jq '.["ssosync/aws-sso/identityStoreId"]' ./cicd/cloudformation/cdk/cdk.context.json -r) \
		-c ./cicd/cloudformation/cdk/google-service-account.json \
		--log-level debug.PH

ONY: check-tools
check-tools:
	@echo "Checking installed tools..."
	@if [ -f $(MOCKERY) ]; then echo "✓ mockery: $$($(MOCKERY) version)"; else echo "✗ mockery: not installed"; fi
	@if [ -f $(GOLANGCI_LINT) ]; then echo "✓ golangci-lint: $$($(GOLANGCI_LINT) --version)"; else echo "✗ golangci-lint: not installed"; fi
	@if [ -f $(GORELEASER_BIN) ]; then echo "✓ goreleaser: $$($(GORELEASER_BIN) --version)"; else echo "✗ goreleaser: not installed"; fi
	@if [ -f $(UPX) ]; then echo "✓ upx: $$($(UPX) --version 2>/dev/null | head -1)"; else echo "✗ upx: not installed"; fi
	@echo "Go version: $$(go version)"

.PHONY: setup
setup: install-deps config
	@echo "Development environment setup complete"

.PHONY: ci
ci: fmt vet test
	@echo "CI pipeline completed successfully"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  setup           - Install all dependencies and setup development environment"
	@echo "  install-deps    - Install all development dependencies (mockery, golangci-lint, goreleaser, upx)"
	@echo "  check-tools     - Check status of installed tools"
	@echo "  fmt             - Format code and tidy modules"
	@echo "  generate-mock   - Generate mocks using mockery"
	@echo "  test            - Run tests with coverage"
	@echo "  test-verbose    - Run tests with verbose output"
	@echo "  test-coverage   - Generate HTML coverage report"
	@echo "  vet/lint        - Run linters"
	@echo "  go-build        - Build application using goreleaser"
	@echo "  build           - Clean and build application"
	@echo "  release         - Create release using goreleaser"
	@echo "  dry-run         - Test release without publishing"
	@echo "  clean           - Clean build artifacts"
	@echo "  clean-all       - Clean everything including development tools (.bin/)"
	@echo "  ci              - Run CI pipeline (fmt, vet, test)"
	@echo "  dev             - Run development build (auto-detects $(OS)/$(ARCH))"
	@echo "  dev-help        - Show help for development build"
	@echo "  help            - Show this help message"

.PHONY: dev-help
dev-help: go-build
	@echo "Running development build help for $(OS)/$(ARCH)..."
	@if [ "$(ARCH)" = "arm64" ]; then \
		BINARY_PATH="dist/ssosync_$(OS)_arm64_v8.2/ssosync"; \
	elif [ "$(ARCH)" = "amd64" ]; then \
		BINARY_PATH="dist/ssosync_$(OS)_amd64_v1/ssosync"; \
	else \
		BINARY_PATH=$$(find dist/ -name "ssosync_$(OS)_*" -type f | head -1); \
	fi; \
	if [ ! -f "$$BINARY_PATH" ]; then \
		echo "Error: Binary not found at $$BINARY_PATH"; \
		exit 1; \
	fi; \
	echo "Using binary: $$BINARY_PATH"; \
	$$BINARY_PATH --help