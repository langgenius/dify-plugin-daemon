package exception

import (
	"runtime/debug"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

const (
	PluginDaemonInternalServerError   = "PluginDaemonInternalServerError"
	PluginDaemonBadRequestError       = "PluginDaemonBadRequestError"
	PluginDaemonNotFoundError         = "PluginDaemonNotFoundError"
	PluginDaemonUnauthorizedError     = "PluginDaemonUnauthorizedError"
	PluginDaemonPermissionDeniedError = "PluginDaemonPermissionDeniedError"
	PluginDaemonInvokeError           = "PluginDaemonInvokeError"
	PluginUniqueIdentifierError       = "PluginUniqueIdentifierError"
	PluginNotFoundError               = "PluginNotFoundError"
	PluginUnauthorizedError           = "PluginUnauthorizedError"
	PluginPermissionDeniedError       = "PluginPermissionDeniedError"
	PluginInvokeError                 = "PluginInvokeError"
	PluginConnectionClosedError       = "ConnectionClosedError"
)

func InternalServerError(err error) PluginDaemonError {
	// log the error
	// get traceback
	traceback := string(debug.Stack())
	log.Error("PluginDaemonInternalServerError: %v\n%s", err, traceback)

	return ErrorWithTypeAndCode(err.Error(), PluginDaemonInternalServerError, -500)
}

func BadRequestError(err error) PluginDaemonError {
	return ErrorWithTypeAndCode(err.Error(), PluginDaemonBadRequestError, -400)
}

func NotFoundError(err error) PluginDaemonError {
	return ErrorWithTypeAndCode(err.Error(), PluginDaemonNotFoundError, -404)
}

func UniqueIdentifierError(err error) PluginDaemonError {
	return ErrorWithTypeAndCode(err.Error(), PluginUniqueIdentifierError, -400)
}

// the difference between NotFoundError and ErrPluginNotFound is that the latter is used to notify
// the caller that the plugin is not installed, while the former is a generic NotFound error.
func ErrPluginNotFound() PluginDaemonError {
	return ErrorWithTypeAndCode("plugin not found", PluginNotFoundError, -404)
}

func UnauthorizedError() PluginDaemonError {
	return ErrorWithTypeAndCode("unauthorized", PluginDaemonUnauthorizedError, -401)
}

func PermissionDeniedError(msg string) PluginDaemonError {
	return ErrorWithTypeAndCode(msg, PluginPermissionDeniedError, -403)
}

func InvokePluginError(err error) PluginDaemonError {
	return ErrorWithTypeAndCode(err.Error(), PluginInvokeError, -500)
}

// ConnectionClosedError is designed to be used when the connection was closed unexpectedly
// but the session is not closed yet.
func ConnectionClosedError() PluginDaemonError {
	return ErrorWithTypeAndCode("connection closed", PluginConnectionClosedError, -500)
}
