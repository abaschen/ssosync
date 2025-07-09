OUTPUT = main # Referenced as Handler in sar-template.json

PACKAGED_TEMPLATE = packaged.yaml
STACK_NAME := $(STACK_NAME)
S3_BUCKET := $(S3_BUCKET)
S3_PREFIX := $(S3_PREFIX)
TEMPLATE = sar-template.json
APP_NAME ?= ssosync
GOREL_ARGS ?= 
GOREL ?= goreleaser

.PHONY: generate-mock
generate-mock:
	mockery

.PHONY: test
test: generate-mock
	go test ./... -coverprofile=coverage.out

.PHONY: go-build
go-build:
	$(GOREL) build --snapshot --clean --id ssosync $(GOREL_ARGS)

.PHONY: clean
clean:
	rm -f $(OUTPUT) $(PACKAGED_TEMPLATE) bootstrap coverage.out
	rm -rf dist/ internal/mocks/*

build-SSOSyncFunction: go-build
	cp dist/ssosync_linux_arm64_v8.2/ssosync $(ARTIFACTS_DIR)/bootstrap

.PHONY: config
config:
	go mod download

.PHONY: vet
vet: generate-mock
	golangci-lint run

main: main.go
	echo $(GOREL) build --clean $(GOREL_ARGS)
	$(GOREL) build --clean $(GOREL_ARGS)

# compile the code to run in Lambda (local or real)
.PHONY: lambda
lambda: main

.PHONY: build
build: clean main

.PHONY: release
release: 
	$(GOREL) release --clean $(GOREL_ARGS)

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
	cp dist/ssosync_linux_arm64_v8.2/ssosync ./bootstrap
	sam package --s3-bucket $(S3_BUCKET) --output-template-file $(PACKAGED_TEMPLATE) --s3-prefix $(S3_PREFIX)

.PHONY: deploy
deploy: package
	sam deploy --stack-name $(STACK_NAME) --template-file $(PACKAGED_TEMPLATE) --capabilities CAPABILITY_IAM
