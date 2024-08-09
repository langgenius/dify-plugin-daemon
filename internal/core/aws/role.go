package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

const (
	DIFY_PLUGIN_LAMBDA_EXECUTION_ROLE = "dify-plugin-lambda-execution-role"
)

// getOrCreateLambdaExecutionRole creates a new lambda execution role if it doesn't exist
// or returns the existing role's ARN
func getOrCreateLambdaExecutionRole(ctx context.Context) (string, error) {
	iam_client := iam.NewFromConfig(*aws_lambda_config)

	// Check if the role already exists
	_, err := iam_client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(DIFY_PLUGIN_LAMBDA_EXECUTION_ROLE),
	})

	if err == nil {
		// Role already exists, return its ARN
		return fmt.Sprintf("arn:aws:iam::%s:role/%s", lambda_account_id, DIFY_PLUGIN_LAMBDA_EXECUTION_ROLE), nil
	}

	// Create the role if it doesn't exist
	assume_role_policy_document := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": "lambda.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`

	create_role_output, err := iam_client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(DIFY_PLUGIN_LAMBDA_EXECUTION_ROLE),
		AssumeRolePolicyDocument: aws.String(assume_role_policy_document),
	})
	if err != nil {
		return "", err
	}

	// Attach the AWSLambdaBasicExecutionRole policy
	_, err = iam_client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(DIFY_PLUGIN_LAMBDA_EXECUTION_ROLE),
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
	})
	if err != nil {
		// Delete the role if the policy attachment fails
		_, err1 := iam_client.DeleteRole(ctx, &iam.DeleteRoleInput{
			RoleName: aws.String(DIFY_PLUGIN_LAMBDA_EXECUTION_ROLE),
		})
		if err1 != nil {
			return "", errors.Join(err, err1)
		}

		return "", err
	}

	return *create_role_output.Role.Arn, nil
}
