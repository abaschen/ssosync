# Testing and Quality Assurance Specification

## Overview

This specification defines the comprehensive testing strategy, quality assurance processes, and continuous integration practices for SSO Sync.

## Testing Strategy

### Testing Pyramid

```
                    ┌─────────────────┐
                    │   E2E Tests     │ (5%)
                    │                 │
                ┌───┴─────────────────┴───┐
                │   Integration Tests     │ (15%)
                │                         │
            ┌───┴─────────────────────────┴───┐
            │        Unit Tests               │ (80%)
            │                                 │
            └─────────────────────────────────┘
```

### Test Categories

#### Unit Tests (80% of test suite)

**Scope**: Individual functions, methods, and components
**Coverage Target**: 85%+
**Execution Time**: < 30 seconds total

**Test Structure**:
```go
func TestNewGroup(t *testing.T) {
    tests := []struct {
        name        string
        displayName string
        want        *Group
        wantErr     bool
    }{
        {
            name:        "valid group creation",
            displayName: "Test Group",
            want: &Group{
                Schemas:     []string{constants.SCIMSchemaGroup},
                DisplayName: "Test Group",
                Members:     []string{},
            },
            wantErr: false,
        },
        {
            name:        "empty display name",
            displayName: "",
            want: &Group{
                Schemas:     []string{constants.SCIMSchemaGroup},
                DisplayName: "",
                Members:     []string{},
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := NewGroup(tt.displayName)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

**Unit Test Categories**:
- **Core Logic Tests**: Synchronization algorithms, data transformations
- **Utility Function Tests**: Helper functions, validators, formatters
- **Error Handling Tests**: Error conditions, edge cases, boundary values
- **Configuration Tests**: Config parsing, validation, defaults
- **Mock Tests**: External dependency interactions

#### Integration Tests (15% of test suite)

**Scope**: Component interactions, API integrations
**Coverage Target**: Key integration paths
**Execution Time**: < 5 minutes total

**Integration Test Types**:

**API Integration Tests**:
```go
func TestGoogleAPIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    client := setupGoogleTestClient(t)
    
    // Test user listing
    users, err := client.ListUsers("*")
    require.NoError(t, err)
    assert.NotEmpty(t, users)
    
    // Test group listing
    groups, err := client.ListGroups("*")
    require.NoError(t, err)
    assert.NotEmpty(t, groups)
}
```

**Database Integration Tests**:
```go
func TestAWSIdentityStoreIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    cfg := loadTestConfig(t)
    client := setupAWSTestClient(t, cfg)
    
    // Test user creation
    user := createTestUser()
    createdUser, err := client.CreateUser(user)
    require.NoError(t, err)
    defer cleanupUser(t, client, createdUser.ID)
    
    assert.Equal(t, user.DisplayName, createdUser.DisplayName)
}
```

**Configuration Integration Tests**:
```go
func TestConfigurationValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  *config.Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid configuration",
            config: &config.Config{
                GoogleAdmin:     "admin@test.com",
                SCIMEndpoint:    "https://scim.test.com",
                SCIMAccessToken: "token123",
                Region:          "us-east-1",
                IdentityStoreID: "d-1234567890",
            },
            wantErr: false,
        },
        {
            name: "missing google admin",
            config: &config.Config{
                SCIMEndpoint:    "https://scim.test.com",
                SCIMAccessToken: "token123",
                Region:          "us-east-1",
                IdentityStoreID: "d-1234567890",
            },
            wantErr: true,
            errMsg:  "google admin email is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### End-to-End Tests (5% of test suite)

**Scope**: Complete user workflows, system behavior
**Coverage Target**: Critical user journeys
**Execution Time**: < 15 minutes total

**E2E Test Scenarios**:
```go
func TestCompleteUserSyncWorkflow(t *testing.T) {
    if !isE2EEnabled() {
        t.Skip("E2E tests disabled")
    }

    // Setup test environment
    testEnv := setupE2EEnvironment(t)
    defer testEnv.Cleanup()

    // Create test users in Google Workspace
    testUsers := createTestUsers(t, testEnv.GoogleClient)
    defer cleanupTestUsers(t, testEnv.GoogleClient, testUsers)

    // Run synchronization
    syncConfig := &config.Config{
        GoogleAdmin:     testEnv.GoogleAdmin,
        SCIMEndpoint:    testEnv.SCIMEndpoint,
        SCIMAccessToken: testEnv.SCIMAccessToken,
        Region:          testEnv.Region,
        IdentityStoreID: testEnv.IdentityStoreID,
        SyncMethod:      "groups",
        GroupMatch:      "name:TestGroup*",
        DryRun:          false,
    }

    err := internal.DoSync(context.Background(), syncConfig)
    require.NoError(t, err)

    // Verify users were created in AWS
    for _, testUser := range testUsers {
        awsUser, err := testEnv.AWSClient.FindUserByEmail(testUser.Email)
        require.NoError(t, err)
        assert.Equal(t, testUser.DisplayName, awsUser.DisplayName)
    }
}
```

## Test Implementation

### Mock Framework

**Mockery Configuration** (`.mockery.yml`):
```yaml
with-expecter: true
packages:
  github.com/awslabs/ssosync/internal/interfaces:
    interfaces:
      IdentityStoreAPI:
        config:
          filename: "mocks_{{.InterfaceName}}.go"
          dir: "internal/mocks"
      IdentityStorePaginators:
        config:
          filename: "mocks_{{.InterfaceName}}.go"
          dir: "internal/mocks"
  github.com/awslabs/ssosync/internal/http:
    interfaces:
      Client:
        config:
          filename: "mocks_{{.InterfaceName}}.go"
          dir: "internal/mocks"
```

**Mock Usage Examples**:
```go
func TestSyncUsers(t *testing.T) {
    // Create mocks
    mockAWSClient := mocks.NewMockClient(t)
    mockGoogleClient := mocks.NewMockGoogleClient(t)
    mockIdentityStore := mocks.NewMockIdentityStoreAPI(t)

    // Setup expectations
    mockGoogleClient.EXPECT().
        GetDeletedUsers().
        Return([]*admin.User{}, nil).
        Once()

    mockGoogleClient.EXPECT().
        GetUsers("*").
        Return([]*admin.User{
            {
                PrimaryEmail: "test@example.com",
                Name: &admin.UserName{
                    GivenName:  "Test",
                    FamilyName: "User",
                    FullName:   "Test User",
                },
            },
        }, nil).
        Once()

    // Create sync instance
    sync := internal.New(cfg, mockAWSClient, mockGoogleClient, mockIdentityStore)

    // Execute test
    err := sync.SyncUsers("*")
    assert.NoError(t, err)

    // Verify all expectations were met
    mockAWSClient.AssertExpectations(t)
    mockGoogleClient.AssertExpectations(t)
    mockIdentityStore.AssertExpectations(t)
}
```

### Test Data Management

**Test Fixtures**:
```go
// testdata/users.json
{
  "users": [
    {
      "primaryEmail": "john.doe@example.com",
      "name": {
        "givenName": "John",
        "familyName": "Doe",
        "fullName": "John Doe"
      },
      "suspended": false,
      "orgUnitPath": "/Engineering"
    }
  ]
}

// Loading test data
func loadTestUsers(t *testing.T) []*admin.User {
    data, err := os.ReadFile("testdata/users.json")
    require.NoError(t, err)
    
    var testData struct {
        Users []*admin.User `json:"users"`
    }
    
    err = json.Unmarshal(data, &testData)
    require.NoError(t, err)
    
    return testData.Users
}
```

**Test Builders**:
```go
type UserBuilder struct {
    user *interfaces.User
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        user: &interfaces.User{
            Schemas: []string{constants.SCIMSchemaUser},
            Active:  true,
            Emails:  []interfaces.UserEmail{},
            Addresses: []interfaces.UserAddress{
                {Type: "work"},
            },
        },
    }
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.user.Username = email
    b.user.Emails = []interfaces.UserEmail{
        {
            Value:   email,
            Type:    "work",
            Primary: true,
        },
    }
    return b
}

func (b *UserBuilder) WithName(given, family string) *UserBuilder {
    b.user.Name.GivenName = given
    b.user.Name.FamilyName = family
    b.user.DisplayName = fmt.Sprintf("%s %s", given, family)
    return b
}

func (b *UserBuilder) Build() *interfaces.User {
    return b.user
}

// Usage
testUser := NewUserBuilder().
    WithEmail("test@example.com").
    WithName("Test", "User").
    Build()
```

### Dry-Run Testing

**Dry-Run Implementation**:
```go
type DryRunClient struct {
    realClient aws.Client
    logger     *log.Logger
}

func NewDryRunClient(realClient aws.Client) *DryRunClient {
    return &DryRunClient{
        realClient: realClient,
        logger:     log.WithField("mode", "dry-run"),
    }
}

func (d *DryRunClient) CreateUser(user *interfaces.User) (*interfaces.User, error) {
    d.logger.WithFields(log.Fields{
        "operation": "create_user",
        "email":     user.Username,
        "name":      user.DisplayName,
    }).Info("DRY RUN: Would create user")
    
    // Return mock response
    return &interfaces.User{
        ID:          "dry-run-user-id",
        Username:    user.Username,
        DisplayName: user.DisplayName,
        Active:      user.Active,
    }, nil
}
```

**Dry-Run Tests**:
```go
func TestDryRunMode(t *testing.T) {
    // Setup dry-run configuration
    cfg := &config.Config{
        DryRun: true,
        // ... other config
    }

    // Create clients
    realAWSClient := aws.NewClient(cfg)
    dryRunClient := aws.NewDryRunClient(realAWSClient)

    // Test dry-run behavior
    user := NewUserBuilder().
        WithEmail("test@example.com").
        WithName("Test", "User").
        Build()

    createdUser, err := dryRunClient.CreateUser(user)
    require.NoError(t, err)
    assert.Equal(t, "dry-run-user-id", createdUser.ID)

    // Verify no actual API calls were made
    // (this would be verified through mocks or API call counting)
}
```

## Quality Assurance

### Code Quality Standards

#### Linting Configuration

**golangci-lint Configuration** (`.golangci.yml`):
```yaml
run:
  timeout: 5m
  modules-download-mode: readonly

linters-settings:
  govet:
    check-shadowing: true
  golint:
    min-confidence: 0
  gocyclo:
    min-complexity: 15
  maligned:
    suggest-new: true
  dupl:
    threshold: 100
  goconst:
    min-len: 2
    min-occurrences: 2
  misspell:
    locale: US
  lll:
    line-length: 140
  goimports:
    local-prefixes: github.com/awslabs/ssosync
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style

linters:
  enable:
    - bodyclose
    - deadcode
    - depguard
    - dogsled
    - dupl
    - errcheck
    - exportloopref
    - exhaustive
    - funlen
    - gochecknoinits
    - goconst
    - gocritic
    - gocyclo
    - gofmt
    - goimports
    - golint
    - gomnd
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - interfacer
    - lll
    - misspell
    - nakedret
    - noctx
    - nolintlint
    - rowserrcheck
    - scopelint
    - staticcheck
    - structcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - varcheck
    - whitespace

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gomnd
        - funlen
        - lll
```

#### Security Scanning

**gosec Configuration**:
```yaml
# .gosec.json
{
  "severity": "medium",
  "confidence": "medium",
  "rules": {
    "G101": "Look for hardcoded credentials",
    "G102": "Bind to all interfaces",
    "G103": "Audit the use of unsafe block",
    "G104": "Audit errors not checked",
    "G105": "Audit the use of math/big.Int.Exp",
    "G106": "Audit the use of ssh.InsecureIgnoreHostKey",
    "G107": "Url provided to HTTP request as taint input",
    "G108": "Profiling endpoint automatically exposed on /debug/pprof",
    "G109": "Potential Integer overflow made by strconv.Atoi result conversion to int16/32",
    "G110": "Potential DoS vulnerability via decompression bomb",
    "G201": "SQL query construction using format string",
    "G202": "SQL query construction using string concatenation",
    "G203": "Use of unescaped data in HTML templates",
    "G204": "Audit use of command execution",
    "G301": "Poor file permissions used when creating a directory",
    "G302": "Poor file permissions used with chmod",
    "G303": "Creating tempfile using a predictable path",
    "G304": "File path provided as taint input",
    "G305": "File traversal when extracting zip/tar archive",
    "G306": "Poor file permissions used when writing to a new file",
    "G307": "Deferring a method which returns an error",
    "G401": "Detect the usage of DES, RC4, MD5 or SHA1",
    "G402": "Look for bad TLS connection settings",
    "G403": "Ensure minimum RSA key length of 2048 bits",
    "G404": "Insecure random number source (rand)",
    "G501": "Import blocklist: crypto/md5",
    "G502": "Import blocklist: crypto/des",
    "G503": "Import blocklist: crypto/rc4",
    "G504": "Import blocklist: net/http/cgi",
    "G505": "Import blocklist: crypto/sha1",
    "G601": "Implicit memory aliasing of items from a range statement"
  }
}
```

### Performance Testing

#### Benchmark Tests

**Synchronization Performance**:
```go
func BenchmarkUserSync(b *testing.B) {
    // Setup
    cfg := loadBenchmarkConfig()
    mockClient := setupMockClient(1000) // 1000 users
    sync := internal.New(cfg, mockClient, mockClient, mockClient)

    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        err := sync.SyncUsers("*")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkGroupSync(b *testing.B) {
    cfg := loadBenchmarkConfig()
    mockClient := setupMockClient(100) // 100 groups
    sync := internal.New(cfg, mockClient, mockClient, mockClient)

    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        err := sync.SyncGroups("*")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**Memory Usage Tests**:
```go
func TestMemoryUsage(t *testing.T) {
    var m1, m2 runtime.MemStats
    
    // Measure initial memory
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Run synchronization with large dataset
    cfg := loadTestConfig()
    sync := setupSyncWithLargeDataset(cfg, 10000) // 10k users
    
    err := sync.SyncUsers("*")
    require.NoError(t, err)
    
    // Measure final memory
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // Assert memory usage is within acceptable limits
    memoryUsed := m2.Alloc - m1.Alloc
    maxMemoryMB := uint64(500 * 1024 * 1024) // 500MB limit
    
    assert.Less(t, memoryUsed, maxMemoryMB, 
        "Memory usage exceeded limit: %d MB", memoryUsed/(1024*1024))
}
```

#### Load Testing

**Concurrent Execution Tests**:
```go
func TestConcurrentSync(t *testing.T) {
    const numGoroutines = 10
    const usersPerGoroutine = 100
    
    cfg := loadTestConfig()
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            mockClient := setupMockClient(usersPerGoroutine)
            sync := internal.New(cfg, mockClient, mockClient, mockClient)
            
            if err := sync.SyncUsers("*"); err != nil {
                errors <- fmt.Errorf("goroutine %d failed: %w", id, err)
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        t.Error(err)
    }
}
```

### Test Automation

#### Continuous Integration Pipeline

**GitHub Actions Workflow** (`.github/workflows/test.yml`):
```yaml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.17, 1.18, 1.19]
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: make install-deps
    
    - name: Run linters
      run: make lint
    
    - name: Run security scan
      run: |
        go install github.com/securecodewarrior/github-action-gosec@latest
        gosec -fmt sarif -out results.sarif ./...
    
    - name: Run unit tests
      run: make test
    
    - name: Run integration tests
      run: go test -tags=integration ./... -v
      env:
        INTEGRATION_TEST_ENABLED: true
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella
    
    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: results.sarif

  e2e-test:
    runs-on: ubuntu-latest
    needs: test
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.19
    
    - name: Run E2E tests
      run: go test -tags=e2e ./... -v
      env:
        E2E_TEST_ENABLED: true
        GOOGLE_ADMIN: ${{ secrets.E2E_GOOGLE_ADMIN }}
        GOOGLE_CREDENTIALS: ${{ secrets.E2E_GOOGLE_CREDENTIALS }}
        SCIM_ENDPOINT: ${{ secrets.E2E_SCIM_ENDPOINT }}
        SCIM_ACCESS_TOKEN: ${{ secrets.E2E_SCIM_ACCESS_TOKEN }}
        REGION: ${{ secrets.E2E_REGION }}
        IDENTITY_STORE_ID: ${{ secrets.E2E_IDENTITY_STORE_ID }}
```

#### Test Reporting

**Coverage Reporting**:
```go
//go:build tools
// +build tools

package tools

import (
    _ "github.com/axw/gocov/gocov"
    _ "github.com/AlekSi/gocov-xml"
    _ "github.com/matm/gocov-html"
)

// Generate coverage reports
//go:generate gocov test ./... | gocov-xml > coverage.xml
//go:generate gocov test ./... | gocov-html > coverage.html
```

**Test Result Aggregation**:
```bash
#!/bin/bash
# scripts/test-report.sh

set -e

echo "Running comprehensive test suite..."

# Unit tests with coverage
echo "Running unit tests..."
go test -v -race -coverprofile=unit.out -covermode=atomic ./...

# Integration tests
echo "Running integration tests..."
go test -v -tags=integration -coverprofile=integration.out -covermode=atomic ./...

# Benchmark tests
echo "Running benchmark tests..."
go test -v -bench=. -benchmem ./... > benchmark.out

# Combine coverage reports
echo "Combining coverage reports..."
gocovmerge unit.out integration.out > coverage.out

# Generate reports
echo "Generating coverage reports..."
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out

# Test summary
echo "Test Summary:"
echo "============="
go tool cover -func=coverage.out | tail -1
echo "Benchmark results saved to benchmark.out"
echo "Coverage report saved to coverage.html"
```

### Quality Gates

#### Coverage Requirements

**Minimum Coverage Thresholds**:
- Overall coverage: 80%
- New code coverage: 90%
- Critical path coverage: 95%

**Coverage Validation**:
```bash
#!/bin/bash
# scripts/check-coverage.sh

COVERAGE_THRESHOLD=80
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
    echo "Coverage $COVERAGE% is below threshold $COVERAGE_THRESHOLD%"
    exit 1
fi

echo "Coverage $COVERAGE% meets threshold $COVERAGE_THRESHOLD%"
```

#### Performance Requirements

**Performance Benchmarks**:
```go
// Performance requirements
const (
    MaxSyncTimePerUser  = 100 * time.Millisecond
    MaxSyncTimePerGroup = 50 * time.Millisecond
    MaxMemoryPerUser    = 1024 // bytes
    MaxAPICallsPerUser  = 3
)

func TestPerformanceRequirements(t *testing.T) {
    // Test sync time per user
    start := time.Now()
    err := syncSingleUser()
    duration := time.Since(start)
    
    require.NoError(t, err)
    assert.Less(t, duration, MaxSyncTimePerUser,
        "User sync took too long: %v", duration)
}
```

#### Security Requirements

**Security Test Gates**:
```bash
#!/bin/bash
# scripts/security-check.sh

echo "Running security checks..."

# Static security analysis
gosec -quiet -fmt json -out gosec.json ./...
SECURITY_ISSUES=$(jq '.Issues | length' gosec.json)

if [ "$SECURITY_ISSUES" -gt 0 ]; then
    echo "Security issues found: $SECURITY_ISSUES"
    jq '.Issues[] | {severity: .severity, confidence: .confidence, rule_id: .rule_id, details: .details}' gosec.json
    exit 1
fi

# Dependency vulnerability check
govulncheck ./...

echo "Security checks passed"
```