# Technology Stack

## Language & Runtime
- **Go 1.24**: Primary language with modern Go features
- **AWS Lambda Runtime**: `provided.al2023` on ARM64 architecture

## Key Dependencies
- **AWS SDK v2**: Core AWS service integration
  - `identitystore`: AWS Identity Store API client
  - `secretsmanager`: Secrets management with caching
  - `ssm`: Parameter Store integration
- **Google APIs**: `google.golang.org/api/admin/directory/v1` for Workspace integration
- **CLI Framework**: Cobra for command-line interface
- **Configuration**: Viper for config management
- **Logging**: Logrus with structured logging
- **HTTP**: Retryable HTTP client with backoff
- **Testing**: Testify framework with Mockery for mocks

## Build System & Tools
- **GoReleaser**: Release automation and cross-platform builds
- **Makefile**: Build orchestration
- **AWS SAM**: Serverless application deployment
- **golangci-lint**: Code quality with specific linters:
  - errcheck, govet, ineffassign, staticcheck, unused, gosec, testifylint

## Common Commands

### Development
```bash
make go-build          # Build for current platform
make test              # Run tests with coverage
make generate-mock     # Generate mocks using mockery
make vet               # Run linters
```

### Release & Deployment
```bash
make release           # Full release with GoReleaser
make dry-run           # Test release without publishing
make package           # Package for AWS SAM deployment
make deploy            # Deploy to AWS with SAM
```

### Local Testing
```bash
make dev               # Run with development config (auto-detects architecture)
```

## Architecture Patterns
- **Interface-driven design**: Heavy use of interfaces for testability
- **Dependency injection**: Clean separation of concerns
- **Context propagation**: Proper context handling throughout
- **Error wrapping**: Structured error handling with pkg/errors
- **Configuration layers**: Environment variables, CLI flags, and config files