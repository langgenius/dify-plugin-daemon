package slim

type ErrorCode string

const (
	ErrInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrInvalidArgsJSON ErrorCode = "INVALID_ARGS_JSON"
	ErrConfigLoad      ErrorCode = "CONFIG_LOAD_ERROR"
	ErrConfigInvalid   ErrorCode = "CONFIG_INVALID"
	ErrUnknownMode     ErrorCode = "UNKNOWN_MODE"
	ErrUnknownAction   ErrorCode = "UNKNOWN_ACTION"
	ErrNotImplemented  ErrorCode = "NOT_IMPLEMENTED"
	ErrNetwork         ErrorCode = "NETWORK_ERROR"
	ErrDaemon          ErrorCode = "DAEMON_ERROR"
	ErrStreamRead      ErrorCode = "STREAM_READ_ERROR"
	ErrStreamParse          ErrorCode = "STREAM_PARSE_ERROR"
	ErrPluginNotFound       ErrorCode = "PLUGIN_NOT_FOUND"
	ErrPluginInit           ErrorCode = "PLUGIN_INIT_ERROR"
	ErrPluginExec           ErrorCode = "PLUGIN_EXEC_ERROR"
	ErrPluginDownload       ErrorCode = "PLUGIN_DOWNLOAD_ERROR"
	ErrPluginDownloadTimeout ErrorCode = "PLUGIN_DOWNLOAD_TIMEOUT"
	ErrPluginPackageInvalid ErrorCode = "PLUGIN_PACKAGE_INVALID"
	ErrPluginPackageTooLarge ErrorCode = "PLUGIN_PACKAGE_TOO_LARGE"
	ErrPluginExtract        ErrorCode = "PLUGIN_EXTRACT_ERROR"
)

const (
	ExitOK           = 0
	ExitPluginError  = 1
	ExitInputError   = 2
	ExitNetworkError = 3
	ExitDaemonError  = 4
)

const (
	ModeLocal  = "local"
	ModeRemote = "remote"
)

type SlimError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *SlimError) Error() string {
	return string(e.Code) + ": " + e.Message
}

func (e *SlimError) ExitCode() int {
	switch e.Code {
	case ErrInvalidInput, ErrInvalidArgsJSON, ErrConfigLoad,
		ErrConfigInvalid, ErrUnknownMode, ErrUnknownAction:
		return ExitInputError
	case ErrNetwork, ErrPluginDownload, ErrPluginDownloadTimeout:
		return ExitNetworkError
	case ErrDaemon:
		return ExitDaemonError
	default:
		return ExitPluginError
	}
}

func NewError(code ErrorCode, msg string) *SlimError {
	return &SlimError{Code: code, Message: msg}
}
