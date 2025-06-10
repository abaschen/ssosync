# SSO Sync CDK Infrastructure

This directory contains the AWS CDK infrastructure code for the SSO Sync project.

## Stacks

### SSOSyncPipelineStack

This stack creates the CI/CD pipeline for building, testing and deploying the SSO Sync application.

### SSOSyncSecretsStack

This stack creates the necessary secrets and KMS keys for storing credentials used by the SSO Sync application. It replaces the original `secrets.yaml` CloudFormation template.

## Prerequisites

- Node.js and pnpm installed
- AWS CDK CLI installed
- AWS credentials configured

## Installation

```bash
pnpm install
```

## Configuration

The stacks are configured through the context in `cdk.context.json`. You can either modify the file directly or provide the values through the CDK CLI.

### Secrets Stack Configuration

To deploy the secrets stack, you need to set the following context values in `cdk.context.json`:

```json
{
  "createSecrets": true,
  "managementAccount": "123456789012",
  "delegatedAccount": "234567890123",
  "nonDelegatedAccount": "345678901234",
  "googleAuthMethod": "Google Credentials", // or "Workload Identity Federation" or "Both"
  "googleCredentials": "", // Required if using Google Credentials
  "googleAdminEmail": "", // Required if using Google Credentials
  "wifServiceAccountEmail": "", // Required if using WIF
  "wifClientLibraryConfig": "", // Required if using WIF
  "scimEndpointUrl": "https://scim.region.amazonaws.com/instance-id/scim/v2/",
  "scimEndpointAccessToken": "",
  "identityStoreId": "d-1234567890"
}
```

Or through the CLI (which will override values from cdk.context.json):

```bash
cdk deploy SSOSyncSecretsStack \
  -c createSecrets=true \
  -c managementAccount=123456789012 \
  -c delegatedAccount=234567890123 \
  -c nonDelegatedAccount=345678901234 \
  -c googleAuthMethod="Google Credentials" \
  -c googleCredentials="..." \
  -c googleAdminEmail="admin@example.com" \
  -c scimEndpointUrl="https://scim.region.amazonaws.com/instance-id/scim/v2/" \
  -c scimEndpointAccessToken="..." \
  -c identityStoreId="d-1234567890"
```

### Pipeline Stack Configuration

The pipeline stack requires the following context values in `cdk.context.json`:

```json
{
  "repoName": "ssosync",
  "ownerName": "awslabs",
  "branchName": "master",
  "githubOAuthToken": "",
  "secretsConfig": "" // Output from the secrets stack
}
```

## Deployment

To deploy all stacks:

```bash
cdk deploy --all
```

To deploy a specific stack:

```bash
cdk deploy SSOSyncSecretsStack
# or
cdk deploy SSOSyncPipelineStack
```

## Useful commands

* `cdk diff`        compare deployed stack with current state
* `cdk synth`       emits the synthesized CloudFormation template
* `cdk destroy`     destroy the deployed stack(s)