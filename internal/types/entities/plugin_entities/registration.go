package plugin_entities

type PluginRegistration struct {
	PluginName    string                      `json:"plugin_name"`
	PluginVersion string                      `json:"plugin_version"`
	Models        []ModelProviderRegistration `json:"models"`
	Tools         []ToolProviderRegistration  `json:"tools"`
}

type ToolProviderRegistration struct {
}

type ToolRegistration struct {
}

type ModelProviderRegistration struct {
}

type ModelRegistration struct {
}
