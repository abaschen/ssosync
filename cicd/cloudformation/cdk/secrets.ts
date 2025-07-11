import * as cdk from 'aws-cdk-lib';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as secretsmanager from 'aws-cdk-lib/aws-secretsmanager';
import { StringParameter } from 'aws-cdk-lib/aws-ssm';
import { Construct } from 'constructs';
import { readFileSync } from 'node:fs';
import { SSOSync } from './imports';
import { ParameterNames, SecretNames } from './imports.ts';

//import google secrets from json

export interface SSOSyncSecretsStackProps extends cdk.StackProps {
  managementAccount: string;
  delegatedAccount: string;
  nonDelegatedAccount: string;
  googleAuthMethod: 'Google Credentials' | 'Workload Identity Federation' | 'Both';
  googleCredentialsSecretFile?: string;
  googleAdminEmail?: string;
  wifServiceAccountEmail?: string;
  wifClientLibraryConfig?: string;
  scimEndpointUrl: string;
  scimEndpointAccessToken: string;
  identityStoreId: string;
  ssoRegion: string;
}

export class SSOSyncSecretsStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: SSOSyncSecretsStackProps) {
    super(scope, id, props);
    console.log(props);
    // Create KMS Key for secrets
    const keyForSecrets = new kms.Key(this, 'KeyForSecrets', {
      description: 'Key for protecting SSOSync Secrets in cross-account deployment',
      enabled: true,
      keySpec: kms.KeySpec.SYMMETRIC_DEFAULT,
      keyUsage: kms.KeyUsage.ENCRYPT_DECRYPT,
      pendingWindow: cdk.Duration.days(7),
      removalPolicy: cdk.RemovalPolicy.RETAIN,
    });

    // Create KMS Alias
    new kms.Alias(this, 'KeyAlias', {
      aliasName: 'alias/SSOSync',
      targetKey: keyForSecrets,
    });

    // Add key policy for cross-account access
    keyForSecrets.addToResourcePolicy(new cdk.aws_iam.PolicyStatement({
      effect: cdk.aws_iam.Effect.ALLOW,
      principals: [new cdk.aws_iam.AccountPrincipal(props.managementAccount)],
      actions: ['kms:Decrypt', 'kms:DescribeKey'],
      resources: ['*'],
    }));

    keyForSecrets.addToResourcePolicy(new cdk.aws_iam.PolicyStatement({
      effect: cdk.aws_iam.Effect.ALLOW,
      principals: [new cdk.aws_iam.AccountPrincipal(props.delegatedAccount)],
      actions: ['kms:Decrypt', 'kms:DescribeKey'],
      resources: ['*'],
    }));

    keyForSecrets.addToResourcePolicy(new cdk.aws_iam.PolicyStatement({
      effect: cdk.aws_iam.Effect.ALLOW,
      principals: [new cdk.aws_iam.AccountPrincipal(props.nonDelegatedAccount)],
      actions: ['kms:Decrypt', 'kms:DescribeKey'],
      resources: ['*'],
    }));

    // Helper function to create cross-account secret policy
    const createSecretPolicy = (secret: secretsmanager.ISecret) => {
      secret.addToResourcePolicy(new cdk.aws_iam.PolicyStatement({
        effect: cdk.aws_iam.Effect.ALLOW,
        principals: [
          new cdk.aws_iam.AccountPrincipal(props.managementAccount),
          new cdk.aws_iam.AccountPrincipal(props.delegatedAccount),
          new cdk.aws_iam.AccountPrincipal(props.nonDelegatedAccount),
        ],
        actions: ['secretsmanager:GetSecretValue'],
        resources: ['*'],
      }));
    };

    const grantRead = (parameter: StringParameter) => {
      parameter.grantRead(new cdk.aws_iam.AccountPrincipal(props.managementAccount));
      parameter.grantRead(new cdk.aws_iam.AccountPrincipal(props.delegatedAccount));
      parameter.grantRead(new cdk.aws_iam.AccountPrincipal(props.nonDelegatedAccount));
    }

    // Create secrets based on Google auth method
    const createGoogleCreds = props.googleAuthMethod === 'Google Credentials' || props.googleAuthMethod === 'Both';
    const createWIF = props.googleAuthMethod === 'Workload Identity Federation' || props.googleAuthMethod === 'Both';
    let appConfig = '';
    if (createGoogleCreds) {
      if (!props.googleCredentialsSecretFile || !props.googleAdminEmail) {
        throw new Error('Google credentials and admin email are required when using Google Credentials authentication');
      }

      const credFile = readFileSync(new URL(props.googleCredentialsSecretFile, import.meta.url));
      //parse json in credFile
      const credFileJson: { [key: string]: string } = JSON.parse(credFile.toString());

      const secretGoogleCreds = new secretsmanager.Secret(this, 'GoogleSSOSyncServiceAccountSecret', {
        encryptionKey: keyForSecrets,
        removalPolicy: cdk.RemovalPolicy.RETAIN,
        secretName: SecretNames.GoogleServiceCredentialsSecret,
        secretObjectValue: Object.fromEntries(
          Object.entries(credFileJson).map(([key, value]) => [key, cdk.SecretValue.unsafePlainText(value)])
        ),
      });
      createSecretPolicy(secretGoogleCreds);

      const paramGoogleAdminEmail = new StringParameter(this, 'ParamGoogleAdminEmail', {
        parameterName: ParameterNames.GoogleAdminEmailParam,
        stringValue: props.googleAdminEmail,
      });
      grantRead(paramGoogleAdminEmail);
      SSOSync.exports.GoogleServiceAccountSecret(this, secretGoogleCreds.secretArn);
      SSOSync.exports.GoogleAdminEmailParam(this, paramGoogleAdminEmail.parameterArn);
      appConfig = `${secretGoogleCreds.secretArn},${props.googleAdminEmail},`;
    } else {
      SSOSync.exports.GoogleServiceAccountSecret(this, '-');
      SSOSync.exports.GoogleAdminEmailParam(this, '-');
      appConfig = '-,-,';
    }

    if (createWIF) {
      if (!props.wifServiceAccountEmail || !props.wifClientLibraryConfig) {
        throw new Error('WIF service account email and client library config are required when using WIF authentication');
      }

      const paramWIFEmail = new StringParameter(this, 'SecretWIFServiceAccountEmail', {
        parameterName: ParameterNames.WIFEmailParam,
        stringValue: props.wifServiceAccountEmail,
      });
      grantRead(paramWIFEmail);

      const secretWIFCreds = new secretsmanager.Secret(this, 'SecretWIFClientLibraryCreds', {
        secretName: SecretNames.WIFCredsSecret,
        secretStringValue: cdk.SecretValue.unsafePlainText(props.wifClientLibraryConfig),
        encryptionKey: keyForSecrets,
        removalPolicy: cdk.RemovalPolicy.RETAIN,
      });
      createSecretPolicy(secretWIFCreds);

      SSOSync.exports.WIFCredsSecret(this, secretWIFCreds);
      SSOSync.exports.WIFEmailParam(this, paramWIFEmail.parameterArn);

    } else {
      SSOSync.exports.WIFCredsSecret(this, '-');
      SSOSync.exports.WIFEmailParam(this, '-');
    }
    //FIXME should be in AppConfig too?

    // Create SCIM secrets
    const paramSCIMEndpoint = new StringParameter(this, 'SecretSCIMEndpoint', {
      parameterName: ParameterNames.SCIMEndpointUrlParam,
      stringValue: props.scimEndpointUrl,
    });
    grantRead(paramSCIMEndpoint);

    const secretSCIMToken = new secretsmanager.Secret(this, 'SecretSCIMAccessToken', {
      secretName: SecretNames.SCIMAccessTokenSecret,
      secretStringValue: cdk.SecretValue.unsafePlainText(props.scimEndpointAccessToken),
      encryptionKey: keyForSecrets,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
    });
    createSecretPolicy(secretSCIMToken);



    const paramIdentityStore = new StringParameter(this, 'SecretIdentityStoreID', {
      parameterName: ParameterNames.IdentityStoreIdParam,
      stringValue: props.identityStoreId,
    });
    grantRead(paramIdentityStore);

    const ssoRegion = new StringParameter(this, 'SSORegion', {
      parameterName: ParameterNames.SSORegionParam,
      stringValue: props.ssoRegion,
    });
    grantRead(ssoRegion);
    // output these

    SSOSync.exports.SCIMEndpointParam(this, paramSCIMEndpoint.parameterArn);
    SSOSync.exports.SCIMAccessTokenSecret(this, secretSCIMToken.secretArn);
    SSOSync.exports.SSORegion(this, ssoRegion.parameterArn);
    SSOSync.exports.IdentityStoreIDParam(this, paramIdentityStore.parameterArn);
    SSOSync.exports.KeyForSecretsParam(this, keyForSecrets.keyArn);

    //export as string
    //${SecretGoogleCredentials},${ParamGoogleAdminEmail},${ParamSCIMEndpoint},${SecretSCIMAccessToken},${SSORegion},${ParamIdentityStoreId},arn:aws:kms:${AWS::Region}:${AWS::AccountId}:key/${KeyForSecrets}
    new cdk.CfnOutput(this, 'AppConfig', {
      value: `${appConfig}${paramSCIMEndpoint.parameterArn},${secretSCIMToken.secretArn},${ssoRegion.parameterArn},${paramIdentityStore.parameterArn},${keyForSecrets.keyArn}`,
      exportName: 'AppConfig'
    });
  }
}