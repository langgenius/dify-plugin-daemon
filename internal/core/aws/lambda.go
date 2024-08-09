package aws

// This file contains functions for interacting with AWS Lambda
// it take a docker image and push it to ECR, create a lambda function and deploy it
// also, it will create a function url for the lambda function with auth enabled

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

var (
	aws_lambda_config *aws.Config
	lambda_client     *lambda.Client
	lambda_account_id string
)

// InitLambda initializes the AWS configuration and validates the credentials
// It takes a pointer to the app.Config struct as an argument
func InitLambda(app *app.Config) {
	// Check if required AWS Lambda configuration is provided
	if app.AWSLambdaRegion == nil || app.AWSLambdaAccessKey == nil || app.AWSLambdaSecretKey == nil {
		log.Panic("AWSLambdaRegion, AWSLambdaAccessKey, and AWSLambdaSecretKey must be set")
	}

	// Load AWS configuration with provided credentials
	c, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(*app.AWSLambdaRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			*app.AWSLambdaAccessKey,
			*app.AWSLambdaSecretKey,
			"",
		)),
	)

	// Handle error if AWS config loading fails
	if err != nil {
		log.Panic("Failed to load AWS Lambda config: %v", err)
	}

	log.Info("AWS Lambda config loaded")

	// Create STS client to validate AWS credentials
	stsClient := sts.NewFromConfig(c)
	identity, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Panic("Failed to validate AWS Lambda credentials: %v", err)
	}

	// Get the account ID
	lambda_account_id = *identity.Account

	// Create the Lambda client
	lambda_client = lambda.NewFromConfig(c)

	log.Info("AWS Lambda credentials validated successfully")

	// Store the AWS configuration globally
	aws_lambda_config = &c
}

type LambdaFunction struct {
	FunctionName string
	FunctionARN  string
	FunctionURL  string
}

// PushImageToECR pushes a Docker image to ECR
func PushImageToECR(ctx context.Context, plugin_runtime entities.PluginRuntimeInterface) (string, error) {
	ecr_client := ecr.NewFromConfig(*aws_lambda_config)

	// Create ECR repository if it doesn't exist
	identity, err := plugin_runtime.Identity()
	if err != nil {
		return "", fmt.Errorf("failed to get plugin identity: %v", err)
	}
	image_name := fmt.Sprintf("dify-plugin-%s-%s", identity, plugin_runtime.Checksum())
	repo_name := fmt.Sprintf("dify-plugin-%s", image_name)
	_, err = ecr_client.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repo_name),
	})
	if err != nil && !strings.Contains(err.Error(), "RepositoryAlreadyExistsException") {
		return "", fmt.Errorf("failed to create ECR repository: %v", err)
	}

	// Get ECR authorization token
	auth_output, err := ecr_client.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get ECR authorization token: %v", err)
	}

	if len(auth_output.AuthorizationData) == 0 || auth_output.AuthorizationData[0].AuthorizationToken == nil {
		return "", fmt.Errorf("invalid ECR authorization data")
	}

	auth_token, err := base64.StdEncoding.DecodeString(*auth_output.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode ECR authorization token: %v", err)
	}

	// Extract username and password from auth token
	credentials := strings.SplitN(string(auth_token), ":", 2)
	if len(credentials) != 2 {
		return "", fmt.Errorf("invalid ECR credentials format")
	}

	// TODO: Use the extracted credentials to push the Docker image to ECR
	// This step typically involves using a Docker client library or executing Docker CLI commands

	if auth_output.AuthorizationData[0].ProxyEndpoint == nil {
		return "", fmt.Errorf("invalid ECR proxy endpoint")
	}

	return fmt.Sprintf("%s/%s:latest", *auth_output.AuthorizationData[0].ProxyEndpoint, repo_name), nil
}

// CreateLambdaFunction creates a Lambda function from an ECR image
func CreateLambdaFunction(ctx context.Context, plugin_runtime entities.PluginRuntimeInterface, image_uri string) (*LambdaFunction, error) {
	function_name := fmt.Sprintf("dify-plugin-%s", plugin_runtime.Checksum())

	// Get or create the lambda execution role
	role_arn, err := getOrCreateLambdaExecutionRole(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create Lambda execution role: %v", err)
	}

	create_output, err := lambda_client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(function_name),
		Role:         aws.String(role_arn),
		PackageType:  lambdatypes.PackageTypeImage,
		Code: &lambdatypes.FunctionCode{
			ImageUri: aws.String(image_uri),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Lambda function: %v", err)
	}

	if create_output.FunctionArn == nil {
		return nil, fmt.Errorf("invalid Lambda function creation output")
	}

	// Create function URL
	url_output, err := lambda_client.CreateFunctionUrlConfig(ctx, &lambda.CreateFunctionUrlConfigInput{
		FunctionName: aws.String(function_name),
		AuthType:     lambdatypes.FunctionUrlAuthTypeAwsIam,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create function URL: %v", err)
	}

	if url_output.FunctionUrl == nil {
		return nil, fmt.Errorf("invalid function URL creation output")
	}

	return &LambdaFunction{
		FunctionName: function_name,
		FunctionARN:  *create_output.FunctionArn,
		FunctionURL:  *url_output.FunctionUrl,
	}, nil
}

// ListLambdaFunctions lists all Lambda functions with the "dify-plugin-" prefix
func ListLambdaFunctions(ctx context.Context) ([]*LambdaFunction, error) {
	var functions []*LambdaFunction
	var marker *string

	for {
		output, err := lambda_client.ListFunctions(ctx, &lambda.ListFunctionsInput{
			Marker: marker,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list Lambda functions: %v", err)
		}

		for _, f := range output.Functions {
			if f.FunctionName == nil || f.FunctionArn == nil {
				continue
			}
			if strings.HasPrefix(*f.FunctionName, "dify-plugin-") {
				url_output, err := lambda_client.GetFunctionUrlConfig(ctx, &lambda.GetFunctionUrlConfigInput{
					FunctionName: f.FunctionName,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get function URL for %s: %v", *f.FunctionName, err)
				}

				if url_output.FunctionUrl == nil {
					return nil, fmt.Errorf("invalid function URL output for %s", *f.FunctionName)
				}

				functions = append(functions, &LambdaFunction{
					FunctionName: *f.FunctionName,
					FunctionARN:  *f.FunctionArn,
					FunctionURL:  *url_output.FunctionUrl,
				})
			}
		}

		if output.NextMarker == nil {
			break
		}
		marker = output.NextMarker
	}

	return functions, nil
}

// GetLambdaFunction retrieves a specific Lambda function by its checksum
func GetLambdaFunction(ctx context.Context, identity string, checksum string) (*LambdaFunction, error) {
	function_name := fmt.Sprintf("dify-plugin-%s-%s", identity, checksum)

	output, err := lambda_client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(function_name),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda function: %v", err)
	}

	if output.Configuration == nil || output.Configuration.FunctionName == nil || output.Configuration.FunctionArn == nil {
		return nil, fmt.Errorf("invalid GetFunction output")
	}

	url_output, err := lambda_client.GetFunctionUrlConfig(ctx, &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(function_name),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get function URL: %v", err)
	}

	if url_output.FunctionUrl == nil {
		return nil, fmt.Errorf("invalid function URL output")
	}

	return &LambdaFunction{
		FunctionName: *output.Configuration.FunctionName,
		FunctionARN:  *output.Configuration.FunctionArn,
		FunctionURL:  *url_output.FunctionUrl,
	}, nil
}

// UpdateLambdaFunction updates an existing Lambda function with a new image
func UpdateLambdaFunction(ctx context.Context, plugin_runtime entities.PluginRuntimeInterface, image_uri string) error {
	// Get the function name
	identity, err := plugin_runtime.Identity()
	if err != nil {
		return fmt.Errorf("failed to get plugin identity: %v", err)
	}
	function_name := fmt.Sprintf("dify-plugin-%s-%s", identity, plugin_runtime.Checksum())

	_, err = lambda_client.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(function_name),
		ImageUri:     aws.String(image_uri),
	})
	if err != nil {
		return fmt.Errorf("failed to update Lambda function: %v", err)
	}

	return nil
}

// DeleteLambdaFunction deletes a Lambda function and its associated function URL
func DeleteLambdaFunction(ctx context.Context, plugin_runtime entities.PluginRuntimeInterface) error {
	// Get the function name
	identity, err := plugin_runtime.Identity()
	if err != nil {
		return fmt.Errorf("failed to get plugin identity: %v", err)
	}
	function_name := fmt.Sprintf("dify-plugin-%s-%s", identity, plugin_runtime.Checksum())

	// Delete function URL
	_, err = lambda_client.DeleteFunctionUrlConfig(ctx, &lambda.DeleteFunctionUrlConfigInput{
		FunctionName: aws.String(function_name),
	})
	if err != nil {
		return fmt.Errorf("failed to delete function URL: %v", err)
	}

	// Delete Lambda function
	_, err = lambda_client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(function_name),
	})
	if err != nil {
		return fmt.Errorf("failed to delete Lambda function: %v", err)
	}

	return nil
}
