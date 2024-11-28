package plugin_entities

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

type I18nObject struct {
	EnUS   string `json:"en_US" yaml:"en_US" validate:"required,gt=0,lt=1024"`
	JaJp   string `json:"ja_JP,omitempty" yaml:"ja_JP,omitempty" validate:"lt=1024"`
	ZhHans string `json:"zh_Hans,omitempty" yaml:"zh_Hans,omitempty" validate:"lt=1024"`
	PtBr   string `json:"pt_BR,omitempty" yaml:"pt_BR,omitempty" validate:"lt=1024"`
}

func NewI18nObject(def string) I18nObject {
	return I18nObject{
		EnUS:   def,
		ZhHans: def,
		JaJp:   def,
		PtBr:   def,
	}
}

func isBasicType(fl validator.FieldLevel) bool {
	// allowed int, string, bool, float64
	switch fl.Field().Kind() {
	case reflect.Int, reflect.String, reflect.Bool,
		reflect.Float64, reflect.Float32,
		reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
		reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return true
	case reflect.Ptr:
		// check if the pointer is nil
		if fl.Field().IsNil() {
			return true
		}
	}

	return false
}
