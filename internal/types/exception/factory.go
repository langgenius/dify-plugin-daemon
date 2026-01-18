package exception

import (
	"fmt"
	"runtime/debug"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
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

	ErrorTypePluginInstallation        = "PluginInstallationError"
	ErrorTypePluginConfiguration       = "PluginConfigurationError"
	ErrorTypePluginTimeout             = "PluginTimeoutError"
	ErrorTypePluginValidation          = "PluginValidationError"
	ErrorTypePluginDatabase            = "PluginDatabaseError"
	ErrorTypePluginStorage             = "PluginStorageError"
	ErrorTypePluginSignatureValidation = "PluginSignatureValidationError"
	ErrorTypePluginMemoryLimit         = "PluginMemoryLimitError"
	ErrorTypePluginNetwork             = "PluginNetworkError"
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

// Enhanced error factory functions with context and user-friendly messages

// PluginInstallationError creates an error during plugin installation with context
func PluginInstallationError(pluginID string, operation string, err error) PluginDaemonError {
	return ErrorWithTypeAndArgs(
		formatUserMessage("Failed to %s plugin '%s': %s", operation, pluginID, getUserFriendlyMessage(err)),
		ErrorTypePluginInstallation,
		map[string]any{
			"plugin_id":  pluginID,
			"operation":  operation,
			"suggestion": getSuggestionForOperation(operation),
		},
	)
}

// PluginConfigurationError creates an error for plugin configuration issues
func PluginConfigurationError(pluginID string, configIssue string) PluginDaemonError {
	return ErrorWithTypeAndArgs(
		formatUserMessage("Plugin configuration error for '%s': %s", pluginID, configIssue),
		ErrorTypePluginConfiguration,
		map[string]any{
			"plugin_id":  pluginID,
			"issue":      configIssue,
			"suggestion": "Please check the plugin configuration and ensure all required fields are properly set",
		},
	)
}

// PluginTimeoutError creates an error for timeout scenarios
func PluginTimeoutError(operation string, timeoutSeconds int) PluginDaemonError {
	return ErrorWithTypeAndArgs(
		formatUserMessage("Operation '%s' timed out after %d seconds", operation, timeoutSeconds),
		ErrorTypePluginTimeout,
		map[string]any{
			"operation":       operation,
			"timeout_seconds": timeoutSeconds,
			"suggestion":      "The operation took too long to complete. Try increasing the timeout setting or optimize the plugin performance",
		},
	)
}

// PluginValidationError creates an error for validation failures
func PluginValidationError(field string, value any, reason string) PluginDaemonError {
	return ErrorWithTypeAndArgs(
		formatUserMessage("Validation failed for field '%s': %s", field, reason),
		ErrorTypePluginValidation,
		map[string]any{
			"field":      field,
			"value":      value,
			"reason":     reason,
			"suggestion": getValidationSuggestion(field, reason),
		},
	)
}

// PluginDatabaseError creates an error for database operations
func PluginDatabaseError(operation string, err error) PluginDaemonError {
	return ErrorWithTypeAndArgs(
		formatUserMessage("Database operation '%s' failed: %s", operation, getUserFriendlyMessage(err)),
		ErrorTypePluginDatabase,
		map[string]any{
			"operation":  operation,
			"suggestion": "A database error occurred. Please check if the database service is running and try again",
		},
	)
}

// PluginNotFoundErrorWithContext creates a not found error with context
func PluginNotFoundErrorWithContext(resourceType string, resourceID string) PluginDaemonError {
	return ErrorWithTypeAndArgs(
		formatUserMessage("%s '%s' not found", resourceType, resourceID),
		PluginNotFoundError,
		map[string]any{
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"suggestion":    formatUserMessage("Please verify the %s identifier and try again", resourceType),
		},
	)
}

// PluginStorageError creates an error for storage-related issues
func PluginStorageError(operation string, err error) PluginDaemonError {
	return ErrorWithTypeAndArgs(
		formatUserMessage("Storage operation '%s' failed: %s", operation, getUserFriendlyMessage(err)),
		ErrorTypePluginStorage,
		map[string]any{
			"operation":  operation,
			"suggestion": "A storage error occurred. Please check available disk space and file permissions",
		},
	)
}

// PluginSignatureValidationError creates an error for signature validation failures
func PluginSignatureValidationError(pluginID string, reason string) PluginDaemonError {
	return ErrorWithTypeAndArgs(
		formatUserMessage("Plugin signature validation failed for '%s': %s", pluginID, reason),
		ErrorTypePluginSignatureValidation,
		map[string]any{
			"plugin_id":  pluginID,
			"reason":     reason,
			"suggestion": "The plugin signature could not be verified. If you trust this plugin, you can disable signature verification by setting ENFORCE_LANGGENIUS_PLUGIN_SIGNATURES=false (not recommended)",
		},
	)
}

// PluginNetworkError creates an error for network-related issues
func PluginNetworkError(operation string, target string, err error) PluginDaemonError {
	return ErrorWithTypeAndArgs(
		formatUserMessage("Network error during '%s' to %s: %s", operation, target, getUserFriendlyMessage(err)),
		ErrorTypePluginNetwork,
		map[string]any{
			"operation":  operation,
			"target":     target,
			"suggestion": "A network error occurred. Please check your network connection and ensure the target service is accessible",
		},
	)
}

// Helper functions

func formatUserMessage(format string, args ...any) string {
	// For now, just use fmt.Sprintf. In the future, this could support i18n
	return fmt.Sprintf(format, args...)
}

func getUserFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}
	// Remove technical details and stack traces from error messages
	msg := err.Error()
	// You could add more sophisticated filtering here
	if len(msg) > 200 {
		return msg[:200] + "..."
	}
	return msg
}

func getSuggestionForOperation(operation string) string {
	suggestions := map[string]string{
		"install":   "Ensure the plugin package is valid and all dependencies are available",
		"uninstall": "Make sure the plugin is not in use before uninstalling",
		"load":      "Verify the plugin files are intact and not corrupted",
		"invoke":    "Check if the plugin is properly installed and configured",
		"update":    "Ensure the new version is compatible with your current setup",
	}
	if suggestion, ok := suggestions[operation]; ok {
		return suggestion
	}
	return "Please check the plugin logs for more details and try again"
}

func getValidationSuggestion(field string, reason string) string {
	// Provide field-specific suggestions
	switch {
	case field == "plugin_unique_identifier" && reason == "already exists":
		return "A plugin with this identifier already exists. Please use a unique identifier"
	case field == "version" && reason == "invalid format":
		return "Version must follow semantic versioning format (e.g., 1.0.0)"
	case field == "memory_limit" && reason == "exceeds maximum":
		return "The requested memory limit exceeds the maximum allowed. Please reduce the memory requirement"
	default:
		return "Please check the field value and ensure it meets the requirements"
	}
}
