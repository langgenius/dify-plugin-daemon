package app

type Config struct {
	DifyPluginHost string `envconfig:"DIFY_PLUGIN_HOST"`
	DifyPluginPort int16  `envconfig:"DIFY_PLUGIN_PORT"`
	DifyPluginKey  string `envconfig:"DIFY_PLUGIN_KEY"`
	StoragePath    string `envconfig:"STORAGE_PATH"`
}
