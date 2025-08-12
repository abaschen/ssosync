# Project Structure

## Root Level
- `main.go`: Entry point, delegates to cmd package
- `go.mod/go.sum`: Go module definition and dependencies
- `Makefile`: Build automation and common tasks
- `README.md`: Comprehensive documentation with usage examples

## Core Application (`cmd/`)
- `cmd/root.go`: Cobra CLI setup, Lambda handler, and configuration initialization
- Handles both CLI execution and AWS Lambda runtime detection
- Manages environment variable binding and configuration unmarshaling

## Business Logic (`internal/`)
- `internal/sync.go`: Core synchronization logic and orchestration
- `internal/config/`: Configuration structures and defaults
- `internal/aws/`: AWS service clients and SCIM operations
- `internal/google/`: Google Workspace API integration
- `internal/interfaces/`: Shared interfaces and data structures
- `internal/http/`: HTTP client utilities with retry logic
- `internal/mocks/`: Generated mocks for testing (auto-generated)

## AWS Integration (`internal/aws/`)
- `client.go`: SCIM API client implementation
- `users.go/groups.go`: User and group management operations
- `identitystore/`: AWS Identity Store API wrappers
- `config.go`: AWS-specific configuration

## Deployment & CI/CD (`cicd/`)
- `cicd/cloudformation/`: CloudFormation templates for deployment
- `cicd/build/`: Build scripts and configurations
- `cicd/deploy_patterns/`: Different deployment patterns (single stack, staging, etc.)
- `cicd/tests/`: Integration and account-specific tests

## Configuration Files
- `.golangci.yml`: Linter configuration with security-focused rules
- `.mockery.yml`: Mock generation configuration
- `sar-template.json`: AWS Serverless Application Repository template
- `.goreleaser.yml`: Release automation configuration

## Development Conventions
- **Package naming**: Use descriptive, single-word package names
- **Interface placement**: Interfaces in `internal/interfaces/` for shared use
- **Error handling**: Wrap errors with context using `pkg/errors`
- **Logging**: Use structured logging with logrus and appropriate log levels
- **Testing**: Place tests alongside source files with `_test.go` suffix
- **Mocks**: Auto-generate in `internal/mocks/` using mockery

## File Naming Patterns
- `*_test.go`: Unit tests
- `mocks_*.go`: Generated mock implementations
- `*.yaml/*.yml`: Configuration files
- `*.json`: Templates and structured data