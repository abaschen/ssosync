# API Integration Specification

## Overview

This specification defines the integration patterns and requirements for Google Workspace Directory API and AWS IAM Identity Center APIs used by SSO Sync.

## Google Workspace Directory API Integration

### Authentication and Authorization

**Service Account Setup**:
- Service account with domain-wide delegation
- JSON key file for authentication
- OAuth 2.0 JWT assertion flow

**Required Scopes**:
```
https://www.googleapis.com/auth/admin.directory.user.readonly
https://www.googleapis.com/auth/admin.directory.group.readonly
https://www.googleapis.com/auth/admin.directory.group.member.readonly
```

**Domain-Wide Delegation**:
- Service account authorized for domain-wide delegation
- Admin user email specified for impersonation
- Scopes granted in Google Admin Console

### User API Operations

#### List Users
**Endpoint**: `GET /admin/directory/v1/users`

**Parameters**:
- `domain`: Organization domain
- `query`: Filter query string
- `maxResults`: Page size (1-500)
- `pageToken`: Pagination token
- `orderBy`: Sort order
- `projection`: Data fields to include

**Query Syntax Examples**:
```
name:John*                    # Users with names starting with "John"
email:admin*                  # Users with emails starting with "admin"
orgName=Engineering           # Users in Engineering org unit
isAdmin=true                  # Admin users only
isSuspended=false            # Active users only
isArchived=false             # Non-archived users only
```

**Response Format**:
```json
{
  "kind": "admin#directory#users",
  "users": [
    {
      "id": "user-id",
      "primaryEmail": "user@domain.com",
      "name": {
        "givenName": "John",
        "familyName": "Doe",
        "fullName": "John Doe"
      },
      "suspended": false,
      "archived": false,
      "orgUnitPath": "/Engineering",
      "isAdmin": false
    }
  ],
  "nextPageToken": "token"
}
```

#### Get Deleted Users
**Endpoint**: `GET /admin/directory/v1/users`

**Parameters**:
- `domain`: Organization domain
- `showDeleted`: "true"
- `maxResults`: Page size
- `pageToken`: Pagination token

### Group API Operations

#### List Groups
**Endpoint**: `GET /admin/directory/v1/groups`

**Parameters**:
- `domain`: Organization domain
- `query`: Filter query string
- `maxResults`: Page size (1-200)
- `pageToken`: Pagination token

**Query Syntax Examples**:
```
name:AWS*                     # Groups with names starting with "AWS"
email:engineering-*           # Groups with emails starting with "engineering-"
name=Administrators          # Exact group name match
```

**Response Format**:
```json
{
  "kind": "admin#directory#groups",
  "groups": [
    {
      "id": "group-id",
      "email": "group@domain.com",
      "name": "Group Name",
      "description": "Group description",
      "directMembersCount": "10"
    }
  ],
  "nextPageToken": "token"
}
```

#### List Group Members
**Endpoint**: `GET /admin/directory/v1/groups/{groupKey}/members`

**Parameters**:
- `groupKey`: Group ID or email
- `maxResults`: Page size (1-200)
- `pageToken`: Pagination token
- `roles`: Member roles to include

**Response Format**:
```json
{
  "kind": "admin#directory#members",
  "members": [
    {
      "id": "member-id",
      "email": "member@domain.com",
      "role": "MEMBER",
      "type": "USER",
      "status": "ACTIVE"
    }
  ],
  "nextPageToken": "token"
}
```

### Rate Limiting and Quotas

**Rate Limits**:
- 100 requests per second per user
- 1,500 requests per 100 seconds per user
- Burst capacity available

**Daily Quotas**:
- 1,000,000 requests per day (default)
- Higher quotas available on request

**Best Practices**:
- Implement exponential backoff
- Use batch operations where available
- Cache results when appropriate
- Monitor quota usage

### Error Handling

**Common Error Codes**:
- `400`: Bad Request - Invalid parameters
- `401`: Unauthorized - Authentication failure
- `403`: Forbidden - Insufficient permissions or quota exceeded
- `404`: Not Found - Resource doesn't exist
- `429`: Too Many Requests - Rate limit exceeded
- `500`: Internal Server Error - Google server error

**Retry Strategy**:
```go
// Exponential backoff with jitter
baseDelay := 1 * time.Second
maxDelay := 32 * time.Second
maxRetries := 5

for attempt := 0; attempt < maxRetries; attempt++ {
    if err := makeRequest(); err == nil {
        return nil
    }
    
    if !isRetryableError(err) {
        return err
    }
    
    delay := min(baseDelay * (1 << attempt), maxDelay)
    jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
    time.Sleep(delay + jitter)
}
```

## AWS IAM Identity Center API Integration

### Authentication and Authorization

**SCIM API Authentication**:
- Bearer token authentication
- Token obtained from IAM Identity Center console
- Token has limited lifetime (typically 1 year)

**Identity Store API Authentication**:
- AWS SDK v2 with standard credential chain
- IAM roles or access keys
- Regional endpoint configuration

**Required Permissions**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "identitystore:CreateUser",
        "identitystore:UpdateUser",
        "identitystore:DeleteUser",
        "identitystore:ListUsers",
        "identitystore:CreateGroup",
        "identitystore:UpdateGroup",
        "identitystore:DeleteGroup",
        "identitystore:ListGroups",
        "identitystore:CreateGroupMembership",
        "identitystore:DeleteGroupMembership",
        "identitystore:ListGroupMemberships",
        "identitystore:IsMemberInGroups"
      ],
      "Resource": "*"
    }
  ]
}
```

### SCIM API Operations (Legacy)

#### Create User
**Endpoint**: `POST /scim/v2/Users`

**Request Format**:
```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "user@domain.com",
  "name": {
    "givenName": "John",
    "familyName": "Doe"
  },
  "displayName": "John Doe",
  "active": true,
  "emails": [
    {
      "value": "user@domain.com",
      "type": "work",
      "primary": true
    }
  ],
  "addresses": [
    {
      "type": "work"
    }
  ]
}
```

#### Update User
**Endpoint**: `PUT /scim/v2/Users/{id}`

**Request Format**: Same as create user with ID field

#### Create Group
**Endpoint**: `POST /scim/v2/Groups`

**Request Format**:
```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:Group"],
  "displayName": "Group Name",
  "members": []
}
```

### Identity Store API Operations (Current)

#### Create User
**Operation**: `CreateUser`

**Input**:
```go
&identitystore.CreateUserInput{
    IdentityStoreId: aws.String("d-1234567890"),
    UserName:        aws.String("user@domain.com"),
    DisplayName:     aws.String("John Doe"),
    Name: &types.Name{
        GivenName:  aws.String("John"),
        FamilyName: aws.String("Doe"),
    },
    Emails: []types.Email{
        {
            Value:   aws.String("user@domain.com"),
            Type:    aws.String("work"),
            Primary: aws.Bool(true),
        },
    },
}
```

#### List Users
**Operation**: `ListUsers`

**Input**:
```go
&identitystore.ListUsersInput{
    IdentityStoreId: aws.String("d-1234567890"),
    MaxResults:      aws.Int32(50),
    NextToken:       nextToken,
}
```

**Pagination**:
```go
paginator := identitystore.NewListUsersPaginator(client, input)
for paginator.HasMorePages() {
    output, err := paginator.NextPage(ctx)
    if err != nil {
        return err
    }
    
    for _, user := range output.Users {
        // Process user
    }
}
```

#### Create Group
**Operation**: `CreateGroup`

**Input**:
```go
&identitystore.CreateGroupInput{
    IdentityStoreId: aws.String("d-1234567890"),
    DisplayName:     aws.String("Group Name"),
    Description:     aws.String("Group description"),
}
```

#### Create Group Membership
**Operation**: `CreateGroupMembership`

**Input**:
```go
&identitystore.CreateGroupMembershipInput{
    IdentityStoreId: aws.String("d-1234567890"),
    GroupId:         aws.String("group-id"),
    MemberId: &types.MemberIdMemberUserId{
        Value: "user-id",
    },
}
```

### Rate Limiting and Quotas

**SCIM API Limits**:
- 20 requests per second
- Burst capacity: 40 requests
- Throttling: HTTP 429 responses

**Identity Store API Limits**:
- Service-specific throttling
- Regional limits apply
- Exponential backoff recommended

**Throttling Handling**:
```go
func handleThrottling(err error) bool {
    var throttleErr *types.ThrottlingException
    if errors.As(err, &throttleErr) {
        return true
    }
    
    var serviceErr *smithy.GenericAPIError
    if errors.As(err, &serviceErr) {
        return serviceErr.Code == "Throttling"
    }
    
    return false
}
```

### Error Handling

**Common Error Types**:
- `ConflictException`: Resource already exists
- `ResourceNotFoundException`: Resource not found
- `ValidationException`: Invalid input parameters
- `ThrottlingException`: Rate limit exceeded
- `InternalServerException`: AWS service error

**Error Response Format**:
```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
}
```

**Retry Logic**:
```go
func retryWithBackoff(operation func() error) error {
    backoff := &backoff.ExponentialBackOff{
        InitialInterval:     1 * time.Second,
        RandomizationFactor: 0.1,
        Multiplier:          2.0,
        MaxInterval:         30 * time.Second,
        MaxElapsedTime:      5 * time.Minutes,
        Clock:               backoff.SystemClock,
    }
    
    return backoff.Retry(operation, backoff)
}
```

## Data Mapping Specifications

### User Mapping

**Google Workspace → AWS IAM Identity Center**:
```go
type UserMapping struct {
    // Google fields → AWS fields
    PrimaryEmail    → UserName, Email.Value
    Name.GivenName  → Name.GivenName
    Name.FamilyName → Name.FamilyName
    Name.FullName   → DisplayName
    Suspended       → !Active (inverted)
    OrgUnitPath     → Address.Type = "work"
}
```

**Field Transformations**:
- Email normalization (lowercase)
- Name field concatenation for display name
- Boolean inversion for suspended/active status
- Default values for required fields

### Group Mapping

**Google Workspace → AWS IAM Identity Center**:
```go
type GroupMapping struct {
    // Google fields → AWS fields
    Name        → DisplayName
    Email       → ExternalId (optional)
    Description → Description
    Members     → Group memberships (separate API calls)
}
```

**Member Mapping**:
- Google group members → AWS group memberships
- Nested group flattening
- External user filtering
- Owner/member role normalization

## Caching Strategy

### User Caching

**Cache Key**: User email address
**Cache Duration**: Configurable (default: 1 hour)
**Cache Invalidation**: Time-based expiration

**Implementation**:
```go
type UserCache struct {
    cache map[string]*CachedUser
    mutex sync.RWMutex
    ttl   time.Duration
}

type CachedUser struct {
    User      *admin.User
    Timestamp time.Time
}

func (c *UserCache) Get(email string) (*admin.User, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    cached, exists := c.cache[email]
    if !exists || time.Since(cached.Timestamp) > c.ttl {
        return nil, false
    }
    
    return cached.User, true
}
```

### Group Membership Caching

**Cache Key**: Group email address
**Cache Duration**: Configurable (default: 30 minutes)
**Cache Strategy**: Lazy loading with refresh

## API Client Configuration

### HTTP Client Settings

**Timeouts**:
- Connection timeout: 30 seconds
- Request timeout: 60 seconds
- Keep-alive timeout: 90 seconds

**Retry Configuration**:
- Maximum retries: 3
- Backoff strategy: Exponential with jitter
- Retry conditions: Network errors, 5xx responses, 429 responses

**Connection Pooling**:
- Maximum idle connections: 100
- Maximum connections per host: 10
- Idle connection timeout: 90 seconds

### TLS Configuration

**Security Settings**:
- Minimum TLS version: 1.2
- Certificate validation: Enabled
- Hostname verification: Enabled
- Cipher suites: Modern secure ciphers only

## Monitoring and Observability

### API Metrics

**Request Metrics**:
- Request count by API and operation
- Response time percentiles (p50, p95, p99)
- Error rate by error type
- Retry attempt counts

**Business Metrics**:
- Users synchronized per operation
- Groups synchronized per operation
- API quota utilization
- Cache hit/miss ratios

### Logging

**Request Logging**:
```go
type APIRequest struct {
    Timestamp   time.Time `json:"timestamp"`
    Service     string    `json:"service"`
    Operation   string    `json:"operation"`
    Duration    int64     `json:"duration_ms"`
    StatusCode  int       `json:"status_code"`
    Error       string    `json:"error,omitempty"`
    RequestID   string    `json:"request_id,omitempty"`
}
```

**Structured Logging**:
- Correlation IDs for request tracing
- Contextual information (user count, group count)
- Error details with stack traces
- Performance metrics

### Health Checks

**API Connectivity**:
- Google Workspace API health check
- AWS IAM Identity Center API health check
- Credential validation
- Network connectivity verification

**Health Check Endpoints**:
```go
func healthCheck() HealthStatus {
    return HealthStatus{
        GoogleAPI: checkGoogleAPI(),
        AWSAPI:    checkAWSAPI(),
        Overall:   calculateOverallHealth(),
    }
}
```