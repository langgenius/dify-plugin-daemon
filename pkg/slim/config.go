package slim

import (
	"encoding/json"
	"os"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

type RequestMeta struct {
	TenantID string          `json:"tenant_id"`
	UserID   string          `json:"user_id"`
	Data     json.RawMessage `json:"data"`
}

type LocalConfig struct {
	Folder               string `json:"folder"`
	PythonPath           string `json:"python_path"`
	UvPath               string `json:"uv_path"`
	PythonEnvInitTimeout int    `json:"python_env_init_timeout"`
	MaxExecutionTimeout  int    `json:"max_execution_timeout"`
	PipMirrorURL         string `json:"pip_mirror_url"`
	PipExtraArgs         string `json:"pip_extra_args"`
	MarketplaceURL       string `json:"marketplace_url"`
	IgnoreUvLock         bool   `json:"ignore_uv_lock"`
}

type RemoteConfig struct {
	DaemonAddr string `json:"daemon_addr"`
	DaemonKey  string `json:"daemon_key"`
}

type InvokeContext struct {
	PluginID string
	Action   string
	Request  RequestMeta
}

type SlimConfig struct {
	Mode   string       `json:"mode"`
	Local  LocalConfig  `json:"local"`
	Remote RemoteConfig `json:"remote"`
}

func NewInvokeContext(id, action, argsJSON string) (*InvokeContext, error) {
	var req RequestMeta
	if err := json.Unmarshal([]byte(argsJSON), &req); err != nil {
		return nil, NewError(ErrInvalidArgsJSON, err.Error())
	}
	if req.TenantID == "" {
		req.TenantID = uuid.Nil.String()
	}
	return &InvokeContext{
		PluginID: id,
		Action:   action,
		Request:  req,
	}, nil
}

func LoadConfigFromFile(path string) (*SlimConfig, error) {
	cfg, err := loadConfigFromFile(path)
	if err != nil {
		return nil, err
	}

	if err := fillDefaults(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadConfigFromFile(path string) (*SlimConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, NewError(ErrConfigLoad, err.Error())
	}

	var cfg SlimConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, NewError(ErrConfigLoad, err.Error())
	}

	return &cfg, nil
}

func LoadConfig() (*SlimConfig, error) {
	cfg := loadConfigFromEnv()
	if err := fillDefaults(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadConfigFromEnv() *SlimConfig {
	cfg := &SlimConfig{
		Mode: env("SLIM_MODE", ModeRemote),
	}

	switch cfg.Mode {
	case ModeLocal:
		cfg.Local = LocalConfig{
			Folder:               env("SLIM_FOLDER", ""),
			PythonPath:           env("SLIM_PYTHON_PATH", ""),
			UvPath:               env("SLIM_UV_PATH", ""),
			PythonEnvInitTimeout: envInt("SLIM_PYTHON_ENV_INIT_TIMEOUT", 0),
			MaxExecutionTimeout:  envInt("SLIM_MAX_EXECUTION_TIMEOUT", 0),
			PipMirrorURL:         env("SLIM_PIP_MIRROR_URL", ""),
			PipExtraArgs:         env("SLIM_PIP_EXTRA_ARGS", ""),
			MarketplaceURL:       env("SLIM_MARKETPLACE_URL", ""),
			IgnoreUvLock:         envBool("SLIM_IGNORE_UV_LOCK", false),
		}
	case ModeRemote:
		cfg.Remote = RemoteConfig{
			DaemonAddr: env("SLIM_DAEMON_ADDR", ""),
			DaemonKey:  env("SLIM_DAEMON_KEY", ""),
		}
	}

	return cfg
}

func LoadExtractConfig(configFile string, hasLocalPath bool) (*SlimConfig, error) {
	var cfg *SlimConfig
	var err error
	if configFile != "" {
		cfg, err = loadConfigFromFile(configFile)
	} else {
		cfg = loadConfigFromEnv()
	}
	if err != nil {
		return nil, err
	}

	if err := fillExtractDefaults(cfg, hasLocalPath); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (lc *LocalConfig) toAppConfig() *app.Config {
	return &app.Config{
		PluginWorkingPath:          lc.Folder,
		PluginInstalledPath:        lc.Folder,
		PluginPackageCachePath:     lc.Folder,
		PythonInterpreterPath:      lc.PythonPath,
		UvPath:                     lc.UvPath,
		PythonEnvInitTimeout:       lc.PythonEnvInitTimeout,
		PluginMaxExecutionTimeout:  lc.MaxExecutionTimeout,
		PipMirrorUrl:               lc.PipMirrorURL,
		PipExtraArgs:               lc.PipExtraArgs,
		PluginIgnoreUvLock:         lc.IgnoreUvLock,
		PipPreferBinary:            true,
		PipVerbose:                 true,
		PluginRuntimeBufferSize:    1024,
		PluginRuntimeMaxBufferSize: 5242880,
		Platform:                   app.PLATFORM_LOCAL,
	}
}

func fillDefaults(cfg *SlimConfig) error {
	if cfg.Mode == "" {
		cfg.Mode = ModeRemote
	}

	if cfg.Mode == ModeLocal {
		if cfg.Local.Folder == "" {
			return NewError(ErrConfigInvalid, "local.folder is required")
		}
		if cfg.Local.PythonPath == "" {
			cfg.Local.PythonPath = "python3"
		}
		if cfg.Local.PythonEnvInitTimeout == 0 {
			cfg.Local.PythonEnvInitTimeout = 120
		}
		if cfg.Local.MaxExecutionTimeout == 0 {
			cfg.Local.MaxExecutionTimeout = 600
		}
		if cfg.Local.MarketplaceURL == "" {
			cfg.Local.MarketplaceURL = "https://marketplace.dify.ai"
		}
	}

	if cfg.Mode == ModeRemote {
		if cfg.Remote.DaemonAddr == "" {
			return NewError(ErrConfigInvalid, "remote.daemon_addr is required")
		}
		if cfg.Remote.DaemonKey == "" {
			return NewError(ErrConfigInvalid, "remote.daemon_key is required")
		}
	}

	return nil
}

func fillExtractDefaults(cfg *SlimConfig, hasLocalPath bool) error {
	if cfg.Mode == "" {
		cfg.Mode = ModeRemote
	}

	switch cfg.Mode {
	case ModeLocal:
		if cfg.Local.Folder == "" && !hasLocalPath {
			return NewError(ErrConfigInvalid, "local.folder is required when extract uses -id")
		}
		if cfg.Local.MarketplaceURL == "" {
			cfg.Local.MarketplaceURL = "https://marketplace.dify.ai"
		}
	case ModeRemote:
		if cfg.Remote.DaemonAddr == "" {
			return NewError(ErrConfigInvalid, "remote.daemon_addr is required")
		}
		if cfg.Remote.DaemonKey == "" {
			return NewError(ErrConfigInvalid, "remote.daemon_key is required")
		}
	default:
		return NewError(ErrUnknownMode, cfg.Mode)
	}

	return nil
}
