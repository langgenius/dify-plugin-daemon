package parser

import (
	"reflect"
	"unicode"
)

func StructToMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := val.Type().Field(i)
		fieldName := toSnakeCase(typeField.Name)

		if typeField.Anonymous {
			embeddedFields := StructToMap(field.Interface())
			for k, v := range embeddedFields {
				result[k] = v
			}
		} else {
			result[fieldName] = field.Interface()
		}
	}
	return result
}

func toSnakeCase(str string) string {
	runes := []rune(str)
	length := len(runes)
	var out []rune

	for i := 0; i < length; i++ {
		if unicode.IsUpper(runes[i]) {
			if i > 0 {
				out = append(out, '_')
			}
			out = append(out, unicode.ToLower(runes[i]))
		} else {
			out = append(out, runes[i])
		}
	}

	return string(out)
}
