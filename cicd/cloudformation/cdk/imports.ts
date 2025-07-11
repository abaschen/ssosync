import { CfnOutput, Fn } from "aws-cdk-lib";
import { Construct } from "constructs";

export const SecretPrefix = `ssosync`;
export const ParamPrefix = `/SSOSync`;

export const SecretNames = {
    GoogleServiceCredentialsSecret: `${SecretPrefix}/google/ServiceAccountCredentials`,
    WIFCredsSecret: `${SecretPrefix}/google/WIFCredentials`,
    SCIMAccessTokenSecret: `${SecretPrefix}/aws/SCIMAccessToken`,
}
export const ParameterNames = {
    GoogleAdminEmailParam: `${ParamPrefix}/google/AdminEmail`,
    WIFEmailParam: `${ParamPrefix}/google/WIFServiceAccountEmail`,
    SCIMEndpointUrlParam: `${ParamPrefix}/aws/SCIMEndpointUrl`,
    IdentityStoreIdParam: `${ParamPrefix}/aws/IdentityStoreId`,
    SSORegionParam: `${ParamPrefix}/aws/SSORegion`,
    AppVersionParam: `${ParamPrefix}/Staging/Version`
}
export const Keys = {
    GoogleServiceCredentialsSecret: 'SSOSync:GoogleServiceCredentials',
    GoogleAdminEmailParam: 'SSOSync:GoogleAdminEmail',
    WIFCredsSecret: 'SSOSync:WIFCredentials',
    WIFEmailParam: 'SSOSync:WIFServiceAccountEmail',
    SCIMEndpointParam: 'SSOSync:SCIMEndpoint',
    SCIMAccessTokenSecret: 'SSOSync:SCIMAccessToken',
    SSORegion: 'SSOSync:SSORegion',
    IdentityStoreIDParam: 'SSOSync:IdentityStoreId',
    KeyForSecretsParam: 'SSOSync:KeyForSecrets',
}

export const SSOSync = {
    imports: {
        GoogleServiceCredentials: () => Fn.importValue(Keys.GoogleServiceCredentialsSecret),
        GoogleAdminEmail: () => Fn.importValue(Keys.GoogleAdminEmailParam),
        WIFCredsSecret: () => Fn.importValue(Keys.WIFCredsSecret),
        WIFEmailParam: () => Fn.importValue(Keys.WIFEmailParam),
        SCIMEndpointParam: () => Fn.importValue(Keys.SCIMEndpointParam),
        SCIMAccessTokenSecret: () => Fn.importValue(Keys.SCIMAccessTokenSecret),
        SSORegion: () => Fn.importValue(Keys.SSORegion),
        IdentityStoreIDParam: () => Fn.importValue(Keys.IdentityStoreIDParam),
        KeyForSecretsParam: () => Fn.importValue(Keys.KeyForSecretsParam),
    },
    exports: {
        GoogleServiceAccountSecret: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-GoogleServiceAccount-Secret', {
            description: 'ARN of the secret containing the JSON of Google Service Account credentials',
            value: data,
            exportName: Keys.GoogleServiceCredentialsSecret,
        }),
        WIFCredsSecret: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-WIFCreds-Secret', {
            description: 'ARN of the secret containing the Google Workspace ID Federation credentials',
            value: data,
            exportName: Keys.WIFCredsSecret,
        }),
        GoogleAdminEmailParam: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-GoogleAdminEmail-Param', {
            description: 'ARN of the SSM Parameter containing the Google Admin email',
            value: data,
            exportName: Keys.GoogleAdminEmailParam,
        }),
        WIFEmailParam: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-WIFEmail-Param', {
            description: 'ARN of the SSM Parameter containing the Google Workspace ID Federation email',
            value: data,
            exportName: Keys.WIFEmailParam,
        }),
        SCIMEndpointParam: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-SCIMEndpoint-Param', {
            description: 'ARN of the SSM Parameter containing the SCIM Endpoint',
            value: data,
            exportName: Keys.SCIMEndpointParam,
        }),
        SCIMAccessTokenSecret: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-SCIMAccessToken-Secret', {
            description: 'ARN of the secret containing the SCIM Access Token',
            value: data,
            exportName: Keys.SCIMAccessTokenSecret,
        }),
        SSORegion: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-SSORegion-Param', {
            description: 'ARN of the SSM Parameter containing the region where SSO is located',
            value: data,
            exportName: Keys.SSORegion,
        }),
        IdentityStoreIDParam: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-IdentityStoreId-Param', {
            description: 'ARN of the SSM Parameter containing the AWS Identity Center Identity Store ID',
            value: data,
            exportName: Keys.IdentityStoreIDParam,
        }),
        KeyForSecretsParam: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-KeyForSecrets-Param', {
            description: 'ARN of the SSM Parameter containing the KMS ARN for secret encryption',
            value: data,
            exportName: Keys.KeyForSecretsParam,
        }),
    }
}