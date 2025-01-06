package manifest_entities

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type Version string

func NewVersion(version string) (Version, error) {
	if !PluginDeclarationVersionRegex.MatchString(version) {
		return "", fmt.Errorf("invalid version")
	}
	return Version(version), nil
}

func (v Version) String() string {
	return string(v)
}

const (
	VERSION_PATTERN   = `\d{1,4}(\.\d{1,4}){2}(-\w{1,16})?`
	VERSION_X_PATTERN = `(\d{1,4}|[xX])`
)

var PluginDeclarationVersionRegex = regexp.MustCompile("^" + VERSION_PATTERN + "$")

func isVersion(fl validator.FieldLevel) bool {
	// version format must be like x.x.x, at least 2 digits and most 5 digits, and it can be ends with a letter
	value := fl.Field().String()
	return PluginDeclarationVersionRegex.MatchString(value)
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("version", isVersion)
}
