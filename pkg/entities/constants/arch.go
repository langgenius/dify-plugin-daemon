package constants

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type Arch string

const (
	AMD64 Arch = "amd64"
	ARM64 Arch = "arm64"
)

func isAvailableArch(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == string(AMD64) || value == string(ARM64)
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("is_available_arch", isAvailableArch)
}
