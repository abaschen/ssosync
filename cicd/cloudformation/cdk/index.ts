#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { SSOSyncPipelineStack } from './developer';
import { SSOSyncSecretsStack } from './secrets';

const app = new cdk.App();

// Create the secrets stack if the required parameters are provided
new SSOSyncSecretsStack(app, 'SSOSyncSecretsStack', {
  managementAccount: app.node.tryGetContext('ssosync/secrets/managementAccount'),
  delegatedAccount: app.node.tryGetContext('ssosync/secrets/delegatedAccount'),
  nonDelegatedAccount: app.node.tryGetContext('ssosync/secrets/nonDelegatedAccount'),

  googleAuthMethod: app.node.tryGetContext('ssosync/secrets/googleAuthMethod') ?? 'Google Credentials',

  googleCredentialsSecretFile: app.node.tryGetContext('ssosync/secrets/GoogleSSOSyncServiceAccount-File'),
  googleAdminEmail: app.node.tryGetContext('ssosync/secrets/googleAdminEmail'),

  wifServiceAccountEmail: app.node.tryGetContext('ssosync/secrets/wifServiceAccountEmail'),
  wifClientLibraryConfig: app.node.tryGetContext('ssosync/secrets/wifClientLibraryConfig'),

  scimEndpointUrl: app.node.getContext('ssosync/aws-sso/scimEndpointUrl'),
  scimEndpointAccessToken: app.node.getContext('ssosync/aws-sso/scimEndpointAccessToken'),
  identityStoreId: app.node.getContext('ssosync/aws-sso/identityStoreId'),
  ssoRegion: app.node.getContext('ssosync/aws-sso/region'),
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: 'eu-west-1'
  }
});

// Create the pipeline stack
new SSOSyncPipelineStack(app, 'SSOSyncPipelineStack', {
  repoName: app.node.tryGetContext('ssosync/repoName') || 'ssosync',
  ownerName: app.node.tryGetContext('ssosync/ownerName') || 'awslabs',
  branchName: app.node.tryGetContext('ssosync/branchName') || 'master',
  githubConnectionArn: app.node.getContext('ssosync/githubConnectionArn'),
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: 'eu-west-1'
  }
});