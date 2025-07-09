// Copyright (c) 2020, Amazon.com, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmd ...
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-secretsmanager-caching-go/v2/secretcache"
	"github.com/awslabs/ssosync/internal"
	"github.com/awslabs/ssosync/internal/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

var cfg *config.Config
var awsConfig aws.Config
var codePipelineClient *codepipeline.Client

var rootCmd = &cobra.Command{
	Version: "dev",
	Use:     "ssosync",
	Short:   "SSO Sync, making AWS SSO be populated automagically",
	Long: `A command line tool to enable you to synchronise your Google
Apps (Google Workspace) users to AWS Single Sign-on (AWS SSO)
Complete documentation is available at https://github.com/awslabs/ssosync`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := internal.DoSync(ctx, cfg)
		if err != nil {
			return err
		}
		return nil
	},
}

// Execute is the entry point of the command. If we are
// running inside of AWS Lambda, we use the Lambda
// execution path.
func Execute() {
	if cfg.IsLambda {
		log.Info("Executing as Lambda")
		lambda.Start(Handler)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// Handler for when executing as a lambda
func Handler(ctx context.Context, event events.CodePipelineEvent) (string, error) {
	log.Debug(event)
	err := rootCmd.Execute()

	cfg.IsLambdaRunningInCodePipeline = len(event.CodePipelineJob.ID) > 0

	if cfg.IsLambdaRunningInCodePipeline {
		log.Info("Lambda has been invoked by CodePipeline")

		if err != nil {
			// notify codepipeline and mark its job execution as Failure
			log.Fatal(errors.Wrap(err, "Notifying CodePipeline and mark its job execution as Failure").Error())
			jobID := event.CodePipelineJob.ID
			if len(jobID) == 0 {
				panic("CodePipeline Job ID is not set")
			}
			// mark the job as Failure.
			cplFailure := &codepipeline.PutJobFailureResultInput{
				JobId: aws.String(jobID),
				FailureDetails: &types.FailureDetails{
					Message: aws.String(err.Error()),
					Type:    types.FailureTypeJobFailed,
				},
			}
			_, cplErr := codePipelineClient.PutJobFailureResult(ctx, cplFailure)
			if cplErr != nil {
				log.Fatal(errors.Wrap(err, "Failed to update CodePipeline jobID status").Error())
			}
			return "Failure", err
		}

		log.Info("Notifying CodePipeline and mark its job execution as Success")
		jobID := event.CodePipelineJob.ID
		if len(jobID) == 0 {
			panic("CodePipeline Job ID is not set")
		}
		// mark the job as Success.
		cplSuccess := &codepipeline.PutJobSuccessResultInput{
			JobId: aws.String(jobID),
		}
		_, cplErr := codePipelineClient.PutJobSuccessResult(ctx, cplSuccess)
		if cplErr != nil {
			log.Fatal(errors.Wrap(err, "Failed to update CodePipeline jobID status").Error())
		}

		return "Success", nil
	}

	if err != nil {
		log.Fatal(errors.Wrap(err, "Notifying Lambda and mark this execution as Failure").Error())
		return "Failure", err
	}
	return "Success", nil
}

func init() {
	// init config
	cfg = config.New()
	cfg.IsLambda = len(os.Getenv("AWS_LAMBDA_FUNCTION_NAME")) > 0

	// initialize cobra
	cobra.OnInitialize(initConfig)
	addFlags(rootCmd, cfg)

	rootCmd.SetVersionTemplate(fmt.Sprintf("%s, commit %s, built at %s by %s\n", version, commit, date, builtBy))

	// silence on the root cmd
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// allow to read in from environment
	viper.SetEnvPrefix("ssosync")
	viper.AutomaticEnv()

	appEnvVars := []string{
		"google_admin",
		"google_credentials",
		"scim_access_token",
		"scim_endpoint",
		"log_level",
		"log_format",
		"ignore_users",
		"ignore_groups",
		"include_groups",
		"user_match",
		"group_match",
		"sync_method",
		"region",
		"identity_store_id",
	}

	for _, e := range appEnvVars {
		if err := viper.BindEnv(e); err != nil {
			log.Fatal(errors.Wrap(err, "cannot bind environment variable").Error())
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(errors.Wrap(err, "cannot unmarshal config").Error())
	}

	if cfg.IsLambda {
		configLambda()
	}

	// config logger
	logConfig(cfg)

}

var parameterCache = make(map[string]string)
var secretCache *secretcache.Cache

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func configLambda() {
	ctx := context.Background()

	// Load AWS SDK configuration once
	var err error
	awsConfig, err = aws_config.LoadDefaultConfig(ctx, aws_config.WithRegion(cfg.Region))
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to load AWS SDK configuration").Error())
	}

	// Create clients once
	ssmClient := ssm.NewFromConfig(awsConfig)
	codePipelineClient = codepipeline.NewFromConfig(awsConfig)
	secretCache, err = secretcache.New(func(c *secretcache.Cache) {
		c.Client = secretsmanager.NewFromConfig(awsConfig)
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to create secret cache").Error())
	}

	// Get values from SSM Parameter Store with caching
	cfg.GoogleAdmin = getParameterWithCache(ctx, ssmClient, getEnv("GOOGLE_ADMIN", "/SSOSync/google/AdminEmail"))
	cfg.SCIMEndpoint = getParameterWithCache(ctx, ssmClient, getEnv("SCIM_ENDPOINT", "/SSOSync/aws/SCIMEndpointUrl"))
	cfg.IdentityStoreID = getParameterWithCache(ctx, ssmClient, getEnv("IDENTITY_STORE_ID", "/SSOSync/aws/IdentityStoreId"))
	cfg.Region = getParameterWithCache(ctx, ssmClient, getEnv("REGION", "/SSOSync/aws/Region"))

	// Get sensitive values from Secrets Manager with caching
	cfg.GoogleCredentials = getSecretFromCache(getEnv("GOOGLE_CREDENTIALS", "ssosync/google/ServiceAccountCredentials"))
	cfg.SCIMAccessToken = getSecretFromCache(getEnv("SCIM_ACCESS_TOKEN", "ssosync/aws/SCIMAccessToken"))

	// Handle environment variables for other settings
	if unwrap := os.Getenv("LOG_LEVEL"); unwrap != "" {
		cfg.LogLevel = unwrap
		log.WithField("LogLevel", unwrap).Debug("from EnvVar")
	}

	if unwrap := os.Getenv("LOG_FORMAT"); unwrap != "" {
		cfg.LogFormat = unwrap
		log.WithField("LogFormat", unwrap).Debug("from EnvVar")
	}

	if unwrap := os.Getenv("SYNC_METHOD"); unwrap != "" {
		cfg.SyncMethod = unwrap
		log.WithField("SyncMethod", unwrap).Debug("from EnvVar")
	}

	if unwrap := os.Getenv("USER_MATCH"); unwrap != "" {
		cfg.UserMatch = unwrap
		log.WithField("UserMatch", unwrap).Debug("from EnvVar")
	}

	if unwrap := os.Getenv("GROUP_MATCH"); unwrap != "" {
		cfg.GroupMatch = unwrap
		log.WithField("GroupMatch", unwrap).Debug("from EnvVar")
	}

	if unwrap := os.Getenv("IGNORE_GROUPS"); unwrap != "" {
		cfg.IgnoreGroups = strings.Split(unwrap, ",")
		log.WithField("IgnoreGroups", unwrap).Debug("from EnvVar")
	}

	if unwrap := os.Getenv("IGNORE_USERS"); unwrap != "" {
		cfg.IgnoreUsers = strings.Split(unwrap, ",")
		log.WithField("IgnoreUsers", unwrap).Debug("from EnvVar")
	}

	if unwrap := os.Getenv("INCLUDE_GROUPS"); unwrap != "" {
		cfg.IncludeGroups = strings.Split(unwrap, ",")
		log.WithField("IncludeGroups", unwrap).Debug("from EnvVar")
	}
}

func getParameterWithCache(ctx context.Context, client *ssm.Client, paramName string) string {
	if value, exists := parameterCache[paramName]; exists {
		return value
	}

	resp, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(paramName),
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, fmt.Sprintf("cannot read parameter: %s", paramName)).Error())
	}

	value := *resp.Parameter.Value
	parameterCache[paramName] = value
	return value
}

func getSecretFromCache(secretName string) string {
	value, err := secretCache.GetSecretString(secretName)
	if err != nil {
		log.Fatal(errors.Wrap(err, fmt.Sprintf("cannot read secret: %s", secretName)).Error())
	}
	return value
}

func addFlags(_ *cobra.Command, cfg *config.Config) {
	rootCmd.PersistentFlags().StringVarP(&cfg.GoogleCredentials, "google-admin", "a", config.DefaultGoogleCredentials, "path to find credentials file for Google Workspace")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Debug, "debug", "d", config.DefaultDebug, "enable verbose / debug logging")
	rootCmd.PersistentFlags().StringVarP(&cfg.LogFormat, "log-format", "", config.DefaultLogFormat, "log format")
	rootCmd.PersistentFlags().StringVarP(&cfg.LogLevel, "log-level", "", config.DefaultLogLevel, "log level")
	rootCmd.Flags().StringVarP(&cfg.SCIMAccessToken, "access-token", "t", "", "AWS SSO SCIM API Access Token")
	rootCmd.Flags().StringVarP(&cfg.SCIMEndpoint, "endpoint", "e", "", "AWS SSO SCIM API Endpoint")
	rootCmd.Flags().StringVarP(&cfg.GoogleCredentials, "google-credentials", "c", config.DefaultGoogleCredentials, "path to Google Workspace credentials file")
	rootCmd.Flags().StringVarP(&cfg.GoogleAdmin, "google-admin", "u", "", "Google Workspace admin user email")
	rootCmd.Flags().StringSliceVar(&cfg.IgnoreUsers, "ignore-users", []string{}, "ignores these Google Workspace users")
	rootCmd.Flags().StringSliceVar(&cfg.IgnoreGroups, "ignore-groups", []string{}, "ignores these Google Workspace groups")
	rootCmd.Flags().StringSliceVar(&cfg.IncludeGroups, "include-groups", []string{}, "include only these Google Workspace groups, NOTE: only works when --sync-method 'users_groups'")
	rootCmd.Flags().StringVarP(&cfg.UserMatch, "user-match", "m", "", "Google Workspace Users filter query parameter, example: 'name:John*' 'name=John Doe,email:admin*', to sync all users in the directory specify '*'. For query syntax and more examples see: https://developers.google.com/admin-sdk/directory/v1/guides/search-users")
	rootCmd.Flags().StringVarP(&cfg.GroupMatch, "group-match", "g", "*", "Google Workspace Groups filter query parameter, example: 'name:Admin*' 'name=Admins,email:aws-*', to sync all groups (and their member users) specify '*'. For query syntax and more examples see: https://developers.google.com/admin-sdk/directory/v1/guides/search-groups")
	rootCmd.Flags().StringVarP(&cfg.SyncMethod, "sync-method", "s", config.DefaultSyncMethod, "Sync method to use (users_groups|groups)")
	rootCmd.Flags().StringVarP(&cfg.Region, "region", "r", "", "AWS Region where AWS SSO is enabled")
	rootCmd.Flags().StringVarP(&cfg.IdentityStoreID, "identity-store-id", "i", "", "Identifier of Identity Store in AWS SSO")
}

func logConfig(cfg *config.Config) {
	// reset log format
	if cfg.LogFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if cfg.Debug {
		cfg.LogLevel = "debug"
	}

	// set the configured log level
	if level, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.SetLevel(level)
	}
}
