# Product Overview

SSO Sync is a CLI tool and AWS Lambda function that synchronizes Google Workspace (formerly G Suite) users and groups to AWS IAM Identity Center (formerly AWS SSO). 

## Key Features

- **Bidirectional sync**: Syncs users and groups from Google Workspace to AWS SSO
- **Multiple deployment options**: CLI tool, AWS Lambda, or AWS Serverless Application Repository
- **Flexible sync methods**: 
  - `groups` (default): Sync based on Google groups and their members
  - `users_groups`: Sync users first, then groups and memberships
- **Advanced filtering**: Support for user/group matching patterns and ignore lists
- **AWS Integration**: Uses AWS Identity Store API and SCIM endpoints
- **Production ready**: Includes retry logic, caching, and error handling

## Target Users

- AWS organizations using Google Workspace as their identity provider
- DevOps teams managing AWS SSO user provisioning
- Organizations requiring automated identity synchronization

## Core Value Proposition

Eliminates manual user management in AWS SSO by automatically syncing Google Workspace directory changes, supporting both one-time and scheduled synchronization patterns.