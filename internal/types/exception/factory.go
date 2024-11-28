package exception

func InternalServerError(err error) PluginDaemonError {
	return ErrorWithTypeAndCode(err.Error(), "PluginDaemonInternalServerError", -500)
}

func BadRequestError(err error) PluginDaemonError {
	return ErrorWithTypeAndCode(err.Error(), "PluginDaemonBadRequestError", -400)
}

func NotFoundError(err error) PluginDaemonError {
	return ErrorWithTypeAndCode(err.Error(), "PluginDaemonNotFoundError", -404)
}

func PluginUniqueIdentifierError(err error) PluginDaemonError {
	return ErrorWithTypeAndCode(err.Error(), "PluginUniqueIdentifierError", -400)
}

// the difference between NotFoundError and ErrPluginNotFound is that the latter is used to notify
// the caller that the plugin is not installed, while the former is a generic NotFound error.
func ErrPluginNotFound() PluginDaemonError {
	return ErrorWithTypeAndCode("plugin not found", "PluginNotFoundError", -404)
}

func UnauthorizedError() PluginDaemonError {
	return ErrorWithTypeAndCode("unauthorized", "PluginDaemonUnauthorizedError", -401)
}

func PermissionDeniedError(msg string) PluginDaemonError {
	return ErrorWithTypeAndCode(msg, "PluginPermissionDeniedError", -403)
}

func InvokePluginError(err error) PluginDaemonError {
	return ErrorWithTypeAndCode(err.Error(), "PluginInvokeError", -500)
}
