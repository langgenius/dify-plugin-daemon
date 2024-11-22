package exception

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type genericError struct {
	Msg       string         `json:"msg"`
	ErrorType string         `json:"error_type"`
	Args      map[string]any `json:"args"`

	code int
}

func (e *genericError) Error() string {
	return e.Msg
}

func (e *genericError) ToResponse() *entities.Response {
	errorMsg := parser.MarshalJson(e)

	return entities.NewDaemonErrorResponse(e.code, errorMsg)
}

func Error(msg string) PluginDaemonError {
	return &genericError{Msg: msg, code: -500, ErrorType: "unknown"}
}

func ErrorWithCode(msg string, code int) PluginDaemonError {
	return &genericError{Msg: msg, code: code, ErrorType: "unknown"}
}

func ErrorWithType(msg string, errorType string) PluginDaemonError {
	return &genericError{Msg: msg, code: -500, ErrorType: errorType}
}

func ErrorWithTypeAndCode(msg string, errorType string, code int) PluginDaemonError {
	return &genericError{Msg: msg, code: code, ErrorType: errorType}
}

func ErrorWithTypeAndArgs(msg string, errorType string, args map[string]any) PluginDaemonError {
	return &genericError{Msg: msg, code: -500, ErrorType: errorType, Args: args}
}
