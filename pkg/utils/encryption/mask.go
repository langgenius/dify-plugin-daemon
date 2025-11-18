package encryption

import (
	"strings"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func MaskConfigCredentials(
	credentials map[string]any,
	provider_config []plugin_entities.ProviderConfig,
) map[string]any {
	/*
		Mask credentials based on provider config
	*/
	configsMap := make(map[string]plugin_entities.ProviderConfig)
	for _, config := range provider_config {
		configsMap[config.Name] = config
	}

	copiedCredentials := make(map[string]any)
	for key, value := range credentials {
		if config, ok := configsMap[key]; ok {
			if config.Type == plugin_entities.CONFIG_TYPE_SECRET_INPUT {
				if originalValue, ok := value.(string); ok {
					if len(originalValue) > 6 {
						copiedCredentials[key] = originalValue[:2] +
							strings.Repeat("*", len(originalValue)-4) +
							originalValue[len(originalValue)-2:]
					} else {
						copiedCredentials[key] = strings.Repeat("*", len(originalValue))
					}
				} else {
					copiedCredentials[key] = value
				}
			} else {
				copiedCredentials[key] = value
			}
		} else {
			copiedCredentials[key] = value
		}
	}

	return copiedCredentials
}
