package exception

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
)

type genericError struct {
	Message   string         `json:"message"`
	ErrorType string         `json:"error_type"`
	Args      map[string]any `json:"args"`

	code int
}

func (e *genericError) Error() string {
	return e.Message
}

func (e *genericError) ToResponse() *entities.Response {
	// TODO: using struct instead, currently, for compatibility with old code
	errorMsg := parser.MarshalJson(e)

	return entities.NewDaemonErrorResponse(e.code, errorMsg)
}

func Error(msg string) PluginDaemonError {
	return &genericError{Message: msg, code: -500, ErrorType: "unknown"}
}

func ErrorWithCode(msg string, code int) PluginDaemonError {
	return &genericError{Message: msg, code: code, ErrorType: "unknown"}
}

func ErrorWithType(msg string, errorType string) PluginDaemonError {
	return &genericError{Message: msg, code: -500, ErrorType: errorType}
}

func ErrorWithTypeAndCode(msg string, errorType string, code int) PluginDaemonError {
	return &genericError{Message: msg, code: code, ErrorType: errorType}
}

func ErrorWithTypeAndArgs(msg string, errorType string, args map[string]any) PluginDaemonError {
	return &genericError{Message: msg, code: -500, ErrorType: errorType, Args: args}
}
