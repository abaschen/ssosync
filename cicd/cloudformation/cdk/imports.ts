import { CfnOutput, Fn } from "aws-cdk-lib"
import { Construct } from "constructs"

export const Keys = {
    GoogleServiceCredentialsSecret: 'SSOSync:GoogleServiceCredentials',
    WIFCredsSecret: 'SSOSync:WIFCredentials',
    WIFEmailParam: 'SSOSync:WIFServiceAccountEmail',
    SCIMEndpointParam: 'SSOSync:SCIMEndpoint',
    SCIMAccessTokenSecret: 'SSOSync:SCIMAccessToken',
    SecretRegion: 'SSOSync:Region',
    IdentityStoreIDParam: 'SSOSync:IdentityStoreId',
    KeyForSecretsParam: 'SSOSync:KeyForSecrets',
}

export const SSOSync = {
    imports: {
        GoogleServiceCredentials: () => Fn.importValue(Keys.GoogleServiceCredentialsSecret),
        WIFCredsSecret: () => Fn.importValue(Keys.WIFCredsSecret),
        WIFEmailParam: () => Fn.importValue(Keys.WIFEmailParam),
        SCIMEndpointParam: () => Fn.importValue(Keys.SCIMEndpointParam),
        SCIMAccessTokenSecret: () => Fn.importValue(Keys.SCIMAccessTokenSecret),
        SecretRegion: () => Fn.importValue(Keys.SecretRegion),
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
            description: 'ARN of the secret containing the JSON of Google Service Account credentials',
            value: data,
            exportName: Keys.WIFCredsSecret,
        }),
        WIFEmailParam: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-WIFEmail-Param', {
            description: 'ARN of the secret containing the JSON of Google Service Account credentials',
            value: data,
            exportName: Keys.WIFEmailParam,
        }),
        SCIMEndpointParam: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-SCIMEndpoint-Param', {
            description: 'ARN of the secret containing the JSON of Google Service Account credentials',
            value: data,
            exportName: Keys.SCIMEndpointParam,
        }),
        SCIMAccessTokenSecret: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-SCIMAccessToken-Secret', {
            description: 'ARN of the secret containing the JSON of Google Service Account credentials',
            value: data,
            exportName: Keys.SCIMAccessTokenSecret,
        }),
        SecretRegion: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-Region-Param', {
            description: 'ARN of the secret containing the JSON of Google Service Account credentials',
            value: data,
            exportName: Keys.SecretRegion,
        }),
        IdentityStoreIDParam: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-IdentityStoreID-Param', {
            description: 'ARN of the secret containing the JSON of Google Service Account credentials',
            value: data,
            exportName: Keys.IdentityStoreIDParam,
        }),
        KeyForSecretsParam: (stack: Construct, data: any) => new CfnOutput(stack, 'OUT-KeyForSecrets-Param', {
            description: 'ARN of the secret containing the JSON of Google Service Account credentials',
            value: data,
            exportName: Keys.KeyForSecretsParam,
        }),
    }
}