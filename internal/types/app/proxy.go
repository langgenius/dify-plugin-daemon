package app

import (
	"fmt"
	"os"
)

func (config *Config) SetProxyEnvFallbacks() {
	if config.HttpProxy == "" {
		config.HttpProxy = os.Getenv("http_proxy")
	}
	if config.HttpsProxy == "" {
		config.HttpsProxy = os.Getenv("https_proxy")
	}
	if config.NoProxy == "" {
		config.NoProxy = os.Getenv("no_proxy")
	}
}

func (config *Config) ProxyEnv() []string {
	env := []string{}
	if config.HttpProxy != "" {
		env = append(env,
			fmt.Sprintf("HTTP_PROXY=%s", config.HttpProxy),
			fmt.Sprintf("http_proxy=%s", config.HttpProxy),
		)
	}
	if config.HttpsProxy != "" {
		env = append(env,
			fmt.Sprintf("HTTPS_PROXY=%s", config.HttpsProxy),
			fmt.Sprintf("https_proxy=%s", config.HttpsProxy),
		)
	}
	if config.NoProxy != "" {
		env = append(env,
			fmt.Sprintf("NO_PROXY=%s", config.NoProxy),
			fmt.Sprintf("no_proxy=%s", config.NoProxy),
		)
	}
	return env
}
