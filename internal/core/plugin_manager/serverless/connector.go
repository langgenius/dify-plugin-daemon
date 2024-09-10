package serverless

import (
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

var ()

type LambdaFunction struct {
	FunctionName string `json:"function_name" validate:"required"`
	FunctionARN  string `json:"function_arn" validate:"required"`
	FunctionURL  string `json:"function_url" validate:"required"`
}

// Ping the serverless connector, return error if failed
func Ping() error {
	url, err := url.JoinPath(baseurl.String(), "/ping")
	if err != nil {
		return err
	}
	response, err := http_requests.PostAndParse[entities.GenericResponse[string]](
		client,
		url,
		http_requests.HttpHeader(map[string]string{
			"Authorization": SERVERLESS_CONNECTOR_API_KEY,
		}),
	)
	if err != nil {
		return err
	}

	if response.Code != 0 {
		return fmt.Errorf("unexpected response from serverless connector: %s", response.Message)
	}

	if response.Data != "pong" {
		return fmt.Errorf("unexpected response from serverless connector: %s", response.Data)
	}
	return nil
}

var (
	ErrNoLambdaFunction = errors.New("no lambda function found")
)

// Fetch the lambda function from serverless connector, return error if failed
func FetchLambda(identity string, checksum string) (*LambdaFunction, error) {
	request := map[string]any{
		"config": map[string]any{
			"identity": identity,
			"checksum": checksum,
		},
	}

	url, err := url.JoinPath(baseurl.String(), "/v1/lambda/fetch")
	if err != nil {
		return nil, err
	}

	response, err := http_requests.PostAndParse[entities.GenericResponse[LambdaFunction]](
		client,
		url,
		http_requests.HttpHeader(map[string]string{
			"Authorization": SERVERLESS_CONNECTOR_API_KEY,
		}),
		http_requests.HttpPayloadJson(request),
	)
	if err != nil {
		return nil, err
	}

	if response.Code != 0 {
		if response.Code == -404 {
			return nil, ErrNoLambdaFunction
		}
		return nil, fmt.Errorf("unexpected response from serverless connector: %s", response.Message)
	}

	return &response.Data, nil
}

type LaunchAWSLambdaFunctionEvent string

const (
	Error     LaunchAWSLambdaFunctionEvent = "error"
	Info      LaunchAWSLambdaFunctionEvent = "info"
	Lambda    LaunchAWSLambdaFunctionEvent = "lambda"
	LambdaUrl LaunchAWSLambdaFunctionEvent = "lambda_url"
	Done      LaunchAWSLambdaFunctionEvent = "done"
)

type LaunchAWSLambdaFunctionResponse struct {
	Event   LaunchAWSLambdaFunctionEvent `json:"event"`
	Message string                       `json:"message"`
}

// Launch the lambda function from serverless connector, it will receive the context_tar as the input
// and build it a docker image, then run it on serverless platform like AWS Lambda
// it returns a event stream, the caller should consider it as a async operation
func LaunchLambda(identity string, checksum string, context_tar io.Reader) (*stream.StreamResponse[LaunchAWSLambdaFunctionResponse], error) {
	url, err := url.JoinPath(baseurl.String(), "/v1/lambda/launch")
	if err != nil {
		return nil, err
	}

	response, err := http_requests.PostAndParseStream[LaunchAWSLambdaFunctionResponse](
		client,
		url,
		http_requests.HttpHeader(map[string]string{
			"Authorization": SERVERLESS_CONNECTOR_API_KEY,
		}),
		http_requests.HttpReadTimeout(300),
		http_requests.HttpWriteTimeout(300),
		http_requests.HttpPayloadMultipart(
			map[string]string{
				"identity": identity,
				"checksum": checksum,
			},
			map[string]io.Reader{
				"context": context_tar,
			},
		),
	)

	if err != nil {
		return nil, err
	}

	return response, nil
}
