package manifest_entities

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type PluginTag string

const (
	PLUGIN_TAG_SEARCH        PluginTag = "search"
	PLUGIN_TAG_IMAGE         PluginTag = "image"
	PLUGIN_TAG_VIDEOS        PluginTag = "videos"
	PLUGIN_TAG_WEATHER       PluginTag = "weather"
	PLUGIN_TAG_FINANCE       PluginTag = "finance"
	PLUGIN_TAG_DESIGN        PluginTag = "design"
	PLUGIN_TAG_TRAVEL        PluginTag = "travel"
	PLUGIN_TAG_SOCIAL        PluginTag = "social"
	PLUGIN_TAG_NEWS          PluginTag = "news"
	PLUGIN_TAG_MEDICAL       PluginTag = "medical"
	PLUGIN_TAG_PRODUCTIVITY  PluginTag = "productivity"
	PLUGIN_TAG_EDUCATION     PluginTag = "education"
	PLUGIN_TAG_BUSINESS      PluginTag = "business"
	PLUGIN_TAG_ENTERTAINMENT PluginTag = "entertainment"
	PLUGIN_TAG_UTILITIES     PluginTag = "utilities"
	PLUGIN_TAG_AGENT         PluginTag = "agent"
	PLUGIN_TAG_RAG           PluginTag = "rag"
	PLUGIN_TAG_OTHER         PluginTag = "other"
)

func isPluginTag(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(PLUGIN_TAG_SEARCH),
		string(PLUGIN_TAG_IMAGE),
		string(PLUGIN_TAG_VIDEOS),
		string(PLUGIN_TAG_WEATHER),
		string(PLUGIN_TAG_FINANCE),
		string(PLUGIN_TAG_DESIGN),
		string(PLUGIN_TAG_TRAVEL),
		string(PLUGIN_TAG_SOCIAL),
		string(PLUGIN_TAG_NEWS),
		string(PLUGIN_TAG_MEDICAL),
		string(PLUGIN_TAG_PRODUCTIVITY),
		string(PLUGIN_TAG_EDUCATION),
		string(PLUGIN_TAG_BUSINESS),
		string(PLUGIN_TAG_ENTERTAINMENT),
		string(PLUGIN_TAG_UTILITIES),
		string(PLUGIN_TAG_OTHER),
		string(PLUGIN_TAG_AGENT),
		string(PLUGIN_TAG_RAG):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("plugin_tag", isPluginTag)
}
