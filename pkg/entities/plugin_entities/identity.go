package plugin_entities

import (
	"errors"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type PluginUniqueIdentifier string

var (
	// pluginUniqueIdentifierRegexp is a regular expression to validate the plugin unique identifier.
	// It must be in the format of "author/plugin_id:version@checksum".
	// all lowercase. the length of plugin_id must be less than 256, and for version part, it must be ^\d{1,4}(\.\d{1,4}){1,3}(-\w{1,16})?$
	// for checksum, it must be a 32-character hexadecimal string.
	// the author part is optional, if not specified, it will be empty.
	pluginUniqueIdentifierRegexp = regexp.MustCompile(
		`^(?:([a-z0-9_-]{1,64})\/)?([a-z0-9_-]{1,255}):([0-9]{1,4})(\.[0-9]{1,4}){1,3}(-\w{1,16})?@[a-f0-9]{32,64}$`,
	)
)

func NewPluginUniqueIdentifier(identifier string) (PluginUniqueIdentifier, error) {
	if !pluginUniqueIdentifierRegexp.MatchString(identifier) {
		return "", errors.New("plugin_unique_identifier is not valid: " + identifier)
	}
	return PluginUniqueIdentifier(identifier), nil
}

func (p PluginUniqueIdentifier) PluginID() string {
	// try find :
	split := strings.Split(p.String(), ":")
	if len(split) == 2 {
		return split[0]
	}
	return p.String()
}

func (p PluginUniqueIdentifier) Version() manifest_entities.Version {
	// extract version part from the string
	split := strings.Split(p.String(), "@")
	if len(split) == 2 {
		split = strings.Split(split[0], ":")
		if len(split) == 2 {
			return manifest_entities.Version(split[1])
		}
	}
	return ""
}

func (p PluginUniqueIdentifier) RemoteLike() bool {
	// check if the author is a uuid
	_, err := uuid.Parse(p.Author())
	return err == nil
}

func (p PluginUniqueIdentifier) Author() string {
	// extract author part from the string
	split := strings.Split(p.String(), ":")
	if len(split) == 2 {
		split = strings.Split(split[0], "/")
		if len(split) == 2 {
			return split[0]
		}
	}
	return ""
}

func (p PluginUniqueIdentifier) Checksum() string {
	// extract checksum part from the string
	split := strings.Split(p.String(), "@")
	if len(split) == 2 {
		return split[1]
	}
	return ""
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
