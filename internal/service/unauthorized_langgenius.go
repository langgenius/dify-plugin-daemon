package service

import (
	"strings"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

// isUnauthorizedLanggenius checks if a plugin falsely claims to be from Langgenius
func isUnauthorizedLanggenius(declaration *plugin_entities.PluginDeclaration, verification *decoder.Verification) bool {
	// Check if plugin claims to be from Langgenius (case-insensitive)
	claimsLanggenius := strings.ToLower(declaration.Author) == string(decoder.AUTHORIZED_CATEGORY_LANGGENIUS)

	// If claims Langgenius but not properly authorized
	if claimsLanggenius {
		return verification == nil || // if no verification, it's unauthorized
			verification.AuthorizedCategory != decoder.AUTHORIZED_CATEGORY_LANGGENIUS
	}

	// Non-Langgenius plugins are allowed
	return false
}
