import * as cdk from 'aws-cdk-lib';
import * as codebuild from 'aws-cdk-lib/aws-codebuild';
import * as codepipeline from 'aws-cdk-lib/aws-codepipeline';
import * as codepipeline_actions from 'aws-cdk-lib/aws-codepipeline-actions';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as kms from 'aws-cdk-lib/aws-kms';
import { Function } from 'aws-cdk-lib/aws-lambda';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { Construct } from 'constructs';
import { SSOSync } from './imports.ts';

const cicdFolder = '../../../cicd'

export interface SSOSyncPipelineStackProps extends cdk.StackProps {
  repoName: string;
  ownerName: string;
  branchName: string;
  githubConnectionArn: string;
}

export class SSOSyncPipelineStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: SSOSyncPipelineStackProps) {
    super(scope, id, props);

    // Create CloudWatch Log Group
    const pipelineLogGroup = new logs.LogGroup(this, 'CodePipelineLogGroup', {
      retention: logs.RetentionDays.ONE_MONTH,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Create KMS Key for artifacts
    const artifactBucketKey = new kms.Key(this, 'ArtifactBucketKey', {
      description: 'Key for this CodePipeline',
      enableKeyRotation: true,
      pendingWindow: cdk.Duration.days(7),
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Create S3 Bucket for artifacts
    const artifactBucket = new s3.Bucket(this, 'ArtifactBucket', {
      encryption: s3.BucketEncryption.KMS,
      encryptionKey: artifactBucketKey,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      enforceSSL: true,
    });

    // Add bucket policy for SAR access
    artifactBucket.addToResourcePolicy(new iam.PolicyStatement({
      actions: ['s3:GetObject'],
      principals: [new iam.ServicePrincipal('serverlessrepo.amazonaws.com')],
      resources: [artifactBucket.arnForObjects('*')],
      conditions: {
        StringEquals: {
          'aws:SourceAccount': this.account
        }
      }
    }));

    // Create S3 Bucket for application
    const appBucket = new s3.Bucket(this, 'AppBucket', {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      encryption: s3.BucketEncryption.S3_MANAGED,
      enforceSSL: true,
      versioned: true,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Add bucket policy for SAR access
    appBucket.addToResourcePolicy(new iam.PolicyStatement({
      actions: ['s3:GetObject'],
      principals: [new iam.ServicePrincipal('serverlessrepo.amazonaws.com')],
      resources: [appBucket.arnForObjects('*')],
      conditions: {
        StringEquals: {
          'aws:SourceAccount': this.account
        }
      }
    }));

    // Create CodeBuild Projects
    const buildApp = new codebuild.Project(this, 'CodeBuildApp', {
      projectName: 'SSOSync-Build-App',
      description: 'Build project for SSOSync',
      buildSpec: codebuild.BuildSpec.fromAsset(`${cicdFolder}/build/build/buildspec.yml`),
      environment: {
        buildImage: codebuild.LinuxBuildImage.STANDARD_7_0,
        computeType: codebuild.ComputeType.SMALL,
      },
      environmentVariables: {
        ARTIFACT_S3_BUCKET: { value: artifactBucket.bucketName },
        OUTPUT: { value: 'main' },
        APP_NAME: { value: 'ssosync' }
      },
      logging: {
        cloudWatch: {
          logGroup: pipelineLogGroup,
          prefix: 'SSOSync-Build-App'
        }
      }
    });

    const buildPackage = new codebuild.Project(this, 'CodeBuildPackage', {
      projectName: 'SSOSync-Package',
      description: 'SAM package for SSOSync',
      buildSpec: codebuild.BuildSpec.fromAsset(`${cicdFolder}/build/package/buildspec.yml`),
      environment: {
        buildImage: codebuild.LinuxBuildImage.STANDARD_7_0,
        computeType: codebuild.ComputeType.SMALL,
      },
      environmentVariables: {
        ARTIFACT_S3_BUCKET: { value: artifactBucket.bucketName },
        S3Bucket: { value: appBucket.bucketName },
        Template: { value: 'sar-template.yaml' }
      },
      logging: {
        cloudWatch: {
          logGroup: pipelineLogGroup,
          prefix: 'SSOSync-Package'
        }
      }
    });



    const sourceOutput = new codepipeline.Artifact('Source');


    const actionSource = new codepipeline_actions.CodeStarConnectionsSourceAction({
      actionName: 'GitHub',
      owner: props.ownerName,
      repo: props.repoName,
      branch: props.branchName,
      output: sourceOutput,
      connectionArn: props.githubConnectionArn,
      codeBuildCloneOutput: true
    });

    //actions
    console.log(`GitHub source action: ${props.ownerName}/${props.repoName}#${props.branchName}`)


    const buildOutput = new codepipeline.Artifact('Built');
    const actionBuild_goBuild = new codepipeline_actions.CodeBuildAction({
      actionName: 'GoLang-Build',
      project: buildApp,
      input: sourceOutput,
      runOrder: 1,
      outputs: [buildOutput],
    });
    const samPackageRole = new iam.Role(this, 'SAMPackageRole', {
      assumedBy: new iam.ServicePrincipal('codepipeline.amazonaws.com'),
      inlinePolicies: {

        'PipelineAccess': new iam.PolicyDocument({
          statements: [
            new iam.PolicyStatement({
              actions: [
                "s3:Abort*",
                "s3:DeleteObject*",
                "s3:GetBucket*",
                "s3:GetObject*",
                "s3:List*",
                "s3:PutObject",
                "s3:PutObjectLegalHold",
                "s3:PutObjectRetention",
                "s3:PutObjectTagging",
                "s3:PutObjectVersionTagging"
              ],
              resources: [artifactBucket.arnForObjects('*'), artifactBucket.bucketArn],
            }),
            new iam.PolicyStatement({
              actions: ["s3:PutObject",
                "s3:GetObject",
                "s3:GetObjectVersion",
                "s3:GetBucketAcl",
                "s3:GetBucketLocation"],
              resources: [`${appBucket.bucketArn}/*`],
            }),
            new iam.PolicyStatement({
              actions: [
                "kms:Decrypt",
                "kms:DescribeKey",
                "kms:Encrypt",
                "kms:GenerateDataKey*",
                "kms:ReEncrypt*"
              ],
              resources: [artifactBucketKey.keyArn],
            }),
            new iam.PolicyStatement({
              actions: [
                "codebuild:CreateReportGroup",
                "codebuild:CreateReport",
                "codebuild:UpdateReport",
                "codebuild:BatchPutTestCases",
                "codebuild:BatchPutCodeCoverages"
              ],
              resources: [`arn:aws:codebuild:${this.region}:${this.account}:report-group/${buildPackage.projectName}-*`],
            }),
            new iam.PolicyStatement({
              actions: [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
              ],
              resources: [
                `arn:aws:logs:${this.region}:${this.account}:log-group:/aws/codebuild/${buildPackage.projectName}-*`,
                `arn:aws:logs:${this.region}:${this.account}:log-group:/aws/codebuild/${buildPackage.projectName}`
              ],
            }),
          ]
        }),
        'SARAccess': new iam.PolicyDocument({
          statements: [
            new iam.PolicyStatement({
              actions: ['serverlessrepo:CreateApplicationVersion', 'serverlessrepo:UpdateApplicationVersion'],
              resources: ['*'],
            }),
          ]
        }),
        'SSMAccess': new iam.PolicyDocument({
          statements: [
            new iam.PolicyStatement({
              actions: ['ssm:GetParameter'],
              resources: [`arn:aws:ssm:${SSOSync.imports.SecretRegion()}:${this.account}:parameter/ssosync/*`],
            }),
            new iam.PolicyStatement({
              actions: ['ssm:PutParameter'],
              resources: [`arn:aws:ssm:${SSOSync.imports.SecretRegion()}:${this.account}:parameter/ssosync/Staging/*`],
            }),
          ]
        })
      }
    });
    const packageOutput = new codepipeline.Artifact('Packaged');
    const actionBuild_SAMPackage = new codepipeline_actions.CodeBuildAction({
      actionName: 'SAM-Package-SAR-Stage',
      project: buildPackage,
      input: sourceOutput,
      extraInputs: [buildOutput],
      runOrder: 2,
      role: samPackageRole,
      outputs: [packageOutput],

      environmentVariables: {
        AppVersion: { value: actionBuild_goBuild.variable('AppVersion}') },
        AppCommit: { value: actionBuild_goBuild.variable('AppCommit}') },
        AppTag: { value: actionBuild_goBuild.variable('AppTag') }
      }
    });

    const codebuildStagingProjectName = 'SSOSync-Staging';

    const buildStaging = new codebuild.Project(this, 'CodeBuildStaging', {
      projectName: codebuildStagingProjectName,
      description: 'Publish SSOSync to Serverless Application Repository in Staging',
      buildSpec: codebuild.BuildSpec.fromAsset(`${cicdFolder}/account_execution/staging/buildspec.yml`),
      environment: {
        buildImage: codebuild.LinuxBuildImage.STANDARD_7_0,
        computeType: codebuild.ComputeType.SMALL,
      },

      environmentVariables: {
        AppArn: { value: `arn:aws:serverlessrepo:\${this.region}:\${this.account}:applications/SSOSync-Staging` },
        AppVersion: { value: actionBuild_goBuild.variable('AppVersion}') },
        AppCommit: { value: actionBuild_goBuild.variable('AppCommit}') },
        AppTag: { value: actionBuild_goBuild.variable('AppTag') }
      },
      logging: {
        cloudWatch: {
          logGroup: pipelineLogGroup,
          prefix: 'SSOSync-Staging'
        }
      }
    });

    const buildSmokeCLI = new codebuild.Project(this, 'CodeBuildSmokeCLI', {
      projectName: 'SSOSync-Smoke-CLI',
      description: 'Execute within a container on the cli to prove cli invocation',
      buildSpec: codebuild.BuildSpec.fromAsset(`${cicdFolder}/tests/account_execution/cli/buildspec.yml`),
      environment: {
        buildImage: codebuild.LinuxBuildImage.STANDARD_7_0,
        computeType: codebuild.ComputeType.SMALL,
      },
      logging: {
        cloudWatch: {
          logGroup: pipelineLogGroup,
          prefix: 'SSOSync-Smoke-CLI'
        }
      }
    });

    const buildSmokeLambda = new codebuild.Project(this, 'CodeBuildSmokeLambda', {
      projectName: 'SSOSync-Smoke-Lambda',
      description: 'Execute Lambda from within a container, to test invocation without codepipeline event handling',
      buildSpec: codebuild.BuildSpec.fromAsset(`${cicdFolder}/tests/account_execution/lambda/buildspec.yml`),
      environment: {
        buildImage: codebuild.LinuxBuildImage.STANDARD_7_0,
        computeType: codebuild.ComputeType.SMALL,
      },
      logging: {
        cloudWatch: {
          logGroup: pipelineLogGroup,
          prefix: 'SSOSync-Smoke-Lambda'
        }
      }
    });

    // Deploy Stage
    const testsOutput = new codepipeline.Artifact('Tests');

    const actionDeploy_Staging = new codepipeline_actions.CodeBuildAction({
      actionName: 'Staging',
      runOrder: 1,
      project: buildStaging,
      input: sourceOutput,
      extraInputs: [packageOutput, buildOutput],
      outputs: [testsOutput],
    });
    const actionDeploy_StackDeploy = new codepipeline_actions.CloudFormationCreateUpdateStackAction({
      actionName: 'Deploy',
      runOrder: 2,
      stackName: 'TestAccountExecution',
      templatePath: testsOutput.atPath('deploy/stack.yml'),
      adminPermissions: true,
      parameterOverrides: {
        // Add parameters if needed
      },
      extraInputs: [testsOutput],
    })

    // Smoke Tests Stage
    const smokeLambdaOutput = new codepipeline.Artifact('SmokeLambda');
    const smokeCLIOutput = new codepipeline.Artifact('SmokeCLI');

    const actionSmokeTests_Lambda = new codepipeline_actions.CodeBuildAction({
      actionName: 'Lambda',
      project: buildSmokeLambda,
      input: sourceOutput,
      extraInputs: [testsOutput],
      outputs: [smokeLambdaOutput],
    });
    const actionSmokeTests_CLI = new codepipeline_actions.CodeBuildAction({
      actionName: 'CLI',
      project: buildSmokeCLI,
      input: sourceOutput,
      extraInputs: [testsOutput],
      outputs: [smokeCLIOutput],
    });
    const actionSmokeTests_CodePipeline = new codepipeline_actions.LambdaInvokeAction({
      actionName: 'CodePipeline',
      lambda: Function.fromFunctionName(this, 'SSOSyncFunction', 'SSOSyncFunction'),
      inputs: [testsOutput],
    })


    const actionCleanUp = new codepipeline_actions.CloudFormationDeleteStackAction({
      actionName: 'RemoveStack',
      stackName: 'TestAccountExecution',
      adminPermissions: true,
      extraInputs: [testsOutput],
    });

    // Create Pipeline
    const pipeline = new codepipeline.Pipeline(this, 'CodePipeline', {
      pipelineName: 'SSOSync',
      pipelineType: codepipeline.PipelineType.V2,
      artifactBucket: artifactBucket,
      crossAccountKeys: false,
      triggers: [
        {
          providerType: codepipeline.ProviderType.CODE_STAR_SOURCE_CONNECTION,
          gitConfiguration: {
            sourceAction: actionSource,
            pushFilter: [
              { branchesIncludes: [props.branchName] },
              { tagsIncludes: ["^v*"] }
            ]
          },
        }
      ],
      stages: [
        {
          stageName: 'Source',
          actions: [
            actionSource
          ],
        },
        {
          stageName: 'Build',
          actions: [
            actionBuild_goBuild,
            actionBuild_SAMPackage
          ],
        },
        {
          stageName: 'Deploy',
          actions: [
            actionDeploy_Staging,
            actionDeploy_StackDeploy
          ],
        },
        {
          stageName: 'SmokeTests',
          actions: [
            actionSmokeTests_Lambda,
            actionSmokeTests_CLI,
            actionSmokeTests_CodePipeline
          ],
        },
        {
          stageName: 'CleanUp',
          actions: [
            actionCleanUp
          ]
        }
      ]
    });


    pipeline.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['ssm:GetParameters'],
        resources: [`arn:aws:ssm:${SSOSync.imports.SecretRegion()}:${this.account}:parameter/ssosync/*`]
      }));
    //grant read for secrets
    pipeline.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['secretsmanager:GetSecretValue'],
        resources: [`arn:aws:secretsmanager:${SSOSync.imports.SecretRegion()}:${this.account}:secret/ssosync/*`]
      }));
    //grant access to KMS Key to decrypt secret
    pipeline.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['kms:Decrypt', 'kms:DescribeKey'],
        resources: [SSOSync.imports.KeyForSecretsParam()]
      }));

    pipeline.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["codestar-connections:UseConnection"],
        resources: [props.githubConnectionArn]
      }));

    buildApp.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['codepipeline:ListActionExecutions'],
        resources: [pipeline.pipelineArn]
      }));


  }
}