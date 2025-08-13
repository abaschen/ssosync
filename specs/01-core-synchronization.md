# Core Synchronization Specification

## Overview

This specification defines the core synchronization functionality of SSO Sync, including the two primary sync methods and their behavior patterns.

## Sync Methods

### 1. Groups Method (Default)

**Description**: Group-centric synchronization that prioritizes group membership and syncs users as members of those groups.

**Process Flow**:
1. Query Google Workspace for groups matching the filter criteria
2. For each group, retrieve all members
3. Create/update users in AWS IAM Identity Center as needed
4. Create/update groups in AWS IAM Identity Center
5. Assign users to their respective groups
6. Remove users/groups that no longer exist in Google Workspace

**Use Cases**:
- Organizations with well-defined group structures
- Role-based access control scenarios
- Large directories where not all users need AWS access

**Configuration**:
```bash
--sync-method groups
--group-match "name:AWS*,email:engineering-*"
```

### 2. Users Groups Method (Legacy)

**Description**: User-centric synchronization that syncs all matching users first, then handles group memberships.

**Process Flow**:
1. Query Google Workspace for users matching the filter criteria
2. Create/update all matching users in AWS IAM Identity Center
3. Query Google Workspace for groups matching the filter criteria
4. Create/update groups in AWS IAM Identity Center
5. Assign users to groups based on membership
6. Remove users/groups that no longer exist in Google Workspace

**Use Cases**:
- Organizations needing all users synchronized regardless of group membership
- Migration scenarios from legacy systems
- Specific compliance requirements

**Configuration**:
```bash
--sync-method users_groups
--user-match "orgName=Engineering"
--include-groups "aws-admins@company.com,developers@company.com"
```

## Filtering Specifications

### Group Filtering

**Syntax**: Google Directory API query syntax
**Examples**:
- `name:AWS*` - Groups starting with "AWS"
- `email:engineering-*` - Groups with emails starting with "engineering-"
- `name=Administrators` - Exact group name match
- `*` - All groups (use with caution)

**Multiple Patterns**: Separate with commas
```bash
--group-match "name:Admin*,email:aws-*,name=Developers"
```

### User Filtering

**Syntax**: Google Directory API query syntax
**Examples**:
- `name:John*` - Users with names starting with "John"
- `email:admin*` - Users with emails starting with "admin"
- `orgName=Engineering` - Users in Engineering org unit
- `isAdmin=true` - Admin users only
- `*` - All users (use with caution)

**Complex Queries**:
```bash
--user-match "orgName=Engineering orgTitle:Manager,email:admin*"
```

## Ignore and Include Lists

### Ignore Users
**Purpose**: Exclude specific users from synchronization
**Format**: Comma-separated email addresses
**Example**: `--ignore-users "service@company.com,bot@company.com"`

### Ignore Groups
**Purpose**: Exclude specific groups from synchronization
**Format**: Comma-separated group emails
**Example**: `--ignore-groups "temp-group@company.com,test-group@company.com"`

### Include Groups (users_groups method only)
**Purpose**: Limit synchronization to specific groups when using users_groups method
**Format**: Comma-separated group emails
**Example**: `--include-groups "aws-admins@company.com,developers@company.com"`

## User Lifecycle Management

### User Creation
- Maps Google Workspace user attributes to SCIM user schema
- Sets user as active by default
- Includes primary email, display name, given name, family name
- Adds work address type

### User Updates
- Compares existing AWS user with Google Workspace user
- Updates changed attributes (name, email, active status)
- Maintains user ID consistency

### User Deletion
- Handles deleted users from Google Workspace
- Removes users from AWS IAM Identity Center
- Logs deletion operations

### User Suspension
- Handles suspended users in Google Workspace
- Sets user as inactive in AWS IAM Identity Center
- Maintains user record for potential reactivation

## Group Lifecycle Management

### Group Creation
- Maps Google Workspace group to SCIM group schema
- Uses group display name and email
- Initializes empty member list

### Group Updates
- Updates group display name if changed
- Manages group membership changes
- Handles nested group flattening

### Group Deletion
- Removes groups that no longer exist in Google Workspace
- Removes all group memberships
- Logs deletion operations

## Error Handling

### Transient Errors
- Implements retry logic with exponential backoff
- Handles rate limiting from both Google and AWS APIs
- Continues processing other items on individual failures

### Permanent Errors
- Logs detailed error information
- Skips problematic items and continues
- Reports summary of failures at completion

### Validation Errors
- Validates configuration before starting sync
- Checks required parameters
- Verifies API connectivity

## Performance Considerations

### Caching
- Implements user detail caching to reduce API calls
- Caches group membership information
- Configurable cache TTL

### Pagination
- Handles large result sets with pagination
- Configurable page sizes
- Memory-efficient processing

### Rate Limiting
- Respects Google Workspace API rate limits
- Implements AWS SCIM API rate limiting
- Configurable request throttling

## Logging and Monitoring

### Log Levels
- **DEBUG**: Detailed operation logs, API requests/responses
- **INFO**: High-level operation status, user/group counts
- **WARN**: Non-fatal issues, skipped items
- **ERROR**: Fatal errors, configuration issues

### Log Formats
- **TEXT**: Human-readable format for development
- **JSON**: Structured format for production monitoring

### Metrics
- Users created/updated/deleted counts
- Groups created/updated/deleted counts
- API call counts and response times
- Error counts by type

## Security Considerations

### Credential Management
- Google service account credentials stored securely
- AWS SCIM tokens managed through Secrets Manager
- No credentials logged or exposed

### Data Privacy
- Minimal user data synchronized (name, email, group membership)
- No sensitive personal information processed
- Audit trail of all operations

### Access Control
- Requires appropriate Google Workspace admin permissions
- Requires AWS IAM Identity Center admin permissions
- Principle of least privilege applied

## Configuration Validation

### Required Parameters
- Google admin email
- Google credentials file or content
- AWS SCIM endpoint URL
- AWS SCIM access token
- AWS region
- AWS Identity Store ID

### Optional Parameters
- Sync method (defaults to "groups")
- Filter queries (defaults to sync all)
- Ignore/include lists (defaults to empty)
- Logging configuration (defaults to info/text)

### Validation Rules
- Sync method must be "groups" or "users_groups"
- Include groups only valid with "users_groups" method
- Filter queries must use valid Google API syntax
- AWS region must be valid AWS region code