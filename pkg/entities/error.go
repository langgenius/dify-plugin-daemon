package entities

import "github.com/langgenius/dify-plugin-daemon/internal/utils/parser"

type Error struct {
	ErrorType string `json:"error_type"`
	Message   string `json:"message"`
	Args      any    `json:"args"`
}

func (e *Error) Error() string {
	return parser.MarshalJson(e)
}

func NewError(error_type string, message string, args ...any) *Error {
	return &Error{
		ErrorType: error_type,
		Message:   message,
		Args:      args,
	}
}
