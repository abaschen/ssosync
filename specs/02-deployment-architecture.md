# Deployment Architecture Specification

## Overview

This specification defines the various deployment architectures supported by SSO Sync, including CLI, AWS Lambda, and AWS SAM deployments.

## Deployment Options

### 1. CLI Deployment

**Description**: Direct execution on local machines or CI/CD systems

**Architecture**:
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Local Machine   │    │ Google          │    │ AWS IAM         │
│ or CI/CD        │───▶│ Workspace       │    │ Identity Center │
│                 │    │ Directory API   │    │                 │
│ • ssosync CLI   │    └─────────────────┘    │ • SCIM API      │
│ • Credentials   │                           │ • Identity Store│
│ • Config        │    ┌─────────────────┐    │   API           │
└─────────────────┘───▶│ AWS Services    │───▶└─────────────────┘
                       │ • IAM           │
                       │ • Secrets Mgr   │
                       └─────────────────┘
```

**Use Cases**:
- Development and testing
- CI/CD pipeline integration
- One-time migrations
- Troubleshooting and debugging

**Requirements**:
- Go 1.17+ runtime (if building from source)
- AWS credentials configured
- Google service account credentials
- Network access to Google and AWS APIs

**Configuration Methods**:
1. Command-line flags
2. Environment variables
3. Configuration files (future enhancement)

### 2. AWS Lambda Deployment

**Description**: Serverless execution with scheduled triggers

**Architecture**:
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ CloudWatch      │    │ AWS Lambda      │    │ Google          │
│ Events          │───▶│                 │───▶│ Workspace       │
│                 │    │ • ssosync       │    │ Directory API   │
│ • Scheduled     │    │ • Runtime       │    └─────────────────┘
│ • Manual        │    │ • Env Vars      │
└─────────────────┘    └─────────────────┘    ┌─────────────────┐
                              │               │ AWS IAM         │
                              │               │ Identity Center │
                              └──────────────▶│                 │
                                              │ • SCIM API      │
┌─────────────────┐    ┌─────────────────┐    │ • Identity Store│
│ AWS Secrets     │    │ CloudWatch      │    │   API           │
│ Manager         │    │ Logs            │    └─────────────────┘
│                 │    │                 │
│ • Google Creds  │    │ • Execution     │
│ • SCIM Token    │    │   Logs          │
│ • Config        │    │ • Metrics       │
└─────────────────┘    └─────────────────┘
```

**Use Cases**:
- Production automated synchronization
- Scheduled regular updates
- Event-driven synchronization
- Centralized management

**Components**:
- Lambda function with ARM64 runtime
- CloudWatch Events for scheduling
- Secrets Manager for credential storage
- CloudWatch Logs for monitoring
- IAM roles and policies

**Deployment Patterns**:

#### Pattern 1: App + Secrets (Default)
- Creates Lambda function and all secrets
- Single CloudFormation stack
- Suitable for single-account deployments

#### Pattern 2: App Only
- Creates Lambda function only
- Expects existing secrets in same account
- Suitable for shared secret scenarios

#### Pattern 3: Secrets Only
- Creates secrets without Lambda function
- Used with separate app deployment
- Enables secret sharing across environments

#### Pattern 4: Cross-Account
- Lambda in one account, secrets in another
- Enhanced security separation
- Suitable for multi-account organizations

### 3. AWS SAM Deployment

**Description**: Infrastructure as Code deployment using AWS SAM

**Architecture**:
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Developer       │    │ AWS SAM         │    │ AWS             │
│ Machine         │───▶│                 │───▶│ CloudFormation  │
│                 │    │ • template.yaml │    │                 │
│ • SAM CLI       │    │ • Build         │    │ • Stack         │
│ • Source Code   │    │ • Package       │    │ • Resources     │
│ • Config        │    │ • Deploy        │    │ • Outputs       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │ S3 Bucket       │
                       │                 │
                       │ • Artifacts     │
                       │ • Templates     │
                       └─────────────────┘
```

**Use Cases**:
- Custom deployment requirements
- Infrastructure as Code workflows
- Development environment setup
- Advanced configuration needs

**Components**:
- SAM template (template.yaml)
- Build configuration
- Parameter files
- Custom resource definitions

## Runtime Specifications

### CLI Runtime

**Binary Specifications**:
- Cross-platform support (Linux, macOS, Windows)
- ARM64 and AMD64 architectures
- Static binary with no external dependencies
- Compressed with UPX for smaller size

**Resource Requirements**:
- Memory: 64MB minimum, 256MB recommended
- CPU: Single core sufficient
- Network: HTTPS access to Google and AWS APIs
- Storage: 10MB for binary, minimal temp space

### Lambda Runtime

**Runtime Configuration**:
- Runtime: `provided.al2`
- Architecture: ARM64
- Memory: 512MB (configurable)
- Timeout: 15 minutes (configurable)
- Environment: Amazon Linux 2

**Resource Limits**:
- Maximum execution time: 15 minutes
- Memory allocation: 128MB - 10GB
- Temporary storage: 512MB - 10GB
- Concurrent executions: Account limits apply

**Environment Variables**:
```bash
# Required (from Secrets Manager)
GOOGLE_ADMIN=<secret-arn>
GOOGLE_CREDENTIALS=<secret-arn>
SCIM_ENDPOINT=<secret-arn>
SCIM_ACCESS_TOKEN=<secret-arn>
REGION=<secret-arn>
IDENTITY_STORE_ID=<secret-arn>

# Optional
LOG_LEVEL=info
LOG_FORMAT=json
SYNC_METHOD=groups
GROUP_MATCH=*
USER_MATCH=
IGNORE_USERS=
IGNORE_GROUPS=
DRY_RUN=false
```

## Security Architecture

### Network Security

**CLI Deployment**:
- Outbound HTTPS to Google APIs (443)
- Outbound HTTPS to AWS APIs (443)
- No inbound connections required

**Lambda Deployment**:
- VPC deployment optional
- NAT Gateway required if in private subnet
- Security groups for outbound HTTPS only

### IAM Permissions

**CLI Deployment**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "identitystore:*",
        "sso:*"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "arn:aws:secretsmanager:*:*:secret:ssosync/*"
    }
  ]
}
```

**Lambda Deployment**:
- Basic Lambda execution role
- Identity Store API permissions
- Secrets Manager read permissions
- CloudWatch Logs write permissions

### Credential Management

**Google Credentials**:
- Service account JSON stored in Secrets Manager
- Domain-wide delegation configured
- Minimal required scopes granted

**AWS Credentials**:
- SCIM access token stored in Secrets Manager
- IAM roles for Lambda execution
- Principle of least privilege applied

## Monitoring and Observability

### CloudWatch Integration

**Metrics**:
- Lambda execution duration
- Memory utilization
- Error rates
- Custom business metrics

**Logs**:
- Structured JSON logging
- Configurable log levels
- Correlation IDs for tracing
- Error stack traces

**Alarms**:
- Execution failures
- Duration thresholds
- Error rate thresholds
- Custom metric alarms

### Dashboards

**Operational Dashboard**:
- Execution frequency and success rate
- User/group sync counts
- API response times
- Error trends

**Business Dashboard**:
- Total users synchronized
- Group membership changes
- Sync operation history
- Compliance metrics

## Scalability Considerations

### Performance Limits

**Google Workspace API**:
- Rate limits: 100 requests/second/user
- Daily quotas: 1,000,000 requests/day
- Batch operations: Up to 1,000 items

**AWS SCIM API**:
- Rate limits: 20 requests/second
- Burst capacity: 40 requests
- Throttling behavior: Exponential backoff

**Lambda Limits**:
- Maximum execution time: 15 minutes
- Memory allocation: Up to 10GB
- Concurrent executions: Account limits

### Optimization Strategies

**Caching**:
- User detail caching
- Group membership caching
- API response caching

**Batching**:
- Batch user operations
- Batch group operations
- Parallel processing where possible

**Filtering**:
- Precise filter queries
- Incremental synchronization
- Delta change detection

## Disaster Recovery

### Backup Strategies

**Configuration Backup**:
- CloudFormation templates in version control
- Parameter files backed up
- Secrets documented (not stored)

**State Recovery**:
- Idempotent operations
- Full resync capability
- Audit trail preservation

### Failure Scenarios

**Google API Outage**:
- Graceful degradation
- Retry with exponential backoff
- Alert on extended failures

**AWS API Outage**:
- Queue operations for retry
- Maintain operation logs
- Resume from last successful state

**Lambda Timeout**:
- Checkpoint progress
- Resume from checkpoint
- Split large operations

## Cost Optimization

### Lambda Costs

**Factors**:
- Execution frequency
- Memory allocation
- Execution duration
- Request volume

**Optimization**:
- Right-size memory allocation
- Optimize execution time
- Use appropriate scheduling
- Monitor and adjust

### API Costs

**Google Workspace**:
- No direct API costs
- Consider admin license costs

**AWS Services**:
- Lambda execution costs
- Secrets Manager storage costs
- CloudWatch Logs storage costs
- Data transfer costs (minimal)

## Compliance and Governance

### Data Residency

**Google Workspace**:
- Data processed in Google's global infrastructure
- Subject to Google's data processing agreements

**AWS Services**:
- Data processed in specified AWS region
- Subject to AWS data processing agreements

### Audit Requirements

**Logging**:
- All operations logged
- User/group changes tracked
- API calls audited
- Error conditions recorded

**Retention**:
- CloudWatch Logs retention configurable
- Audit trail preservation
- Compliance with organizational policies