package plugin_entities

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

type I18nObject struct {
	EnUS   string `json:"en_US" yaml:"en_US" validate:"required,gt=0,lt=1024"`
	JaJp   string `json:"ja_JP" yaml:"ja_JP" validate:"lt=1024"`
	ZhHans string `json:"zh_Hans" yaml:"zh_Hans" validate:"lt=1024"`
	PtBr   string `json:"pt_BR" yaml:"pt_BR" validate:"lt=1024"`
}

func isBasicType(fl validator.FieldLevel) bool {
	// allowed int, string, bool, float64
	switch fl.Field().Kind() {
	case reflect.Int, reflect.String, reflect.Bool, reflect.Float64:
		return true
	}

	return false
}
