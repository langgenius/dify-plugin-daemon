package encryption

import (
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func MaskConfigCredentials(
	credentials map[string]any,
	provider_config map[string]plugin_entities.ProviderConfig,
) map[string]any {
	/*
		Mask credentials based on provider config
	*/
	copied_credentials := make(map[string]any)
	for key, value := range credentials {
		if config, ok := provider_config[key]; ok {
			if config.Type == plugin_entities.CONFIG_TYPE_SECRET_INPUT {
				if original_value, ok := value.(string); ok {
					if len(original_value) > 6 {
						copied_credentials[key] = original_value[:2] +
							strings.Repeat("*", len(original_value)-4) +
							original_value[len(original_value)-2:]
					} else {
						copied_credentials[key] = strings.Repeat("*", len(original_value))
					}
				} else {
					copied_credentials[key] = value
				}
			} else {
				copied_credentials[key] = value
			}
		} else {
			copied_credentials[key] = value
		}
	}

	return copied_credentials
}
