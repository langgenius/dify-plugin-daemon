package plugin_entities

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type PluginUniqueIdentifier string

var (
	// pluginUniqueIdentifierRegexp is a regular expression to validate the plugin unique identifier.
	// It must be in the format of "plugin_id:version@checksum".
	// all lowercase. the length of plugin_id must be less than 128, and for version part, it must be ^\d{1,4}(\.\d{1,4}){1,3}(-\w{1,16})?$
	// for checksum, it must be a 32-character hexadecimal string.
	pluginUniqueIdentifierRegexp = regexp.MustCompile(
		`^[a-z0-9_-]{1,128}:[0-9]{1,4}(\.[0-9]{1,4}){1,3}(-\w{1,16})?@[a-f0-9]{32}$`,
	)
)

func (p PluginUniqueIdentifier) PluginID() string {
	// try find @
	split := strings.Split(p.String(), "@")
	if len(split) == 2 {
		return split[0]
	}
	return p.String()
}

func (p PluginUniqueIdentifier) String() string {
	return string(p)
}

func (p PluginUniqueIdentifier) Validate() error {
	return validators.GlobalEntitiesValidator.Var(p, "plugin_unique_identifier")
}

func isValidPluginUniqueIdentifier(fl validator.FieldLevel) bool {
	return pluginUniqueIdentifierRegexp.MatchString(fl.Field().String())
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("plugin_unique_identifier", isValidPluginUniqueIdentifier)
}
