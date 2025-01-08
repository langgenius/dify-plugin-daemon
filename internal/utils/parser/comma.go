package parser

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

// ParserCommaSeparatedValues parses the comma separated values
// and returns a map of key-value pairs
// examples:
// data: a=1,b=2
//
//	T: type struct {
//		A int `comma:"a"`
//		B string `comma:"b"`
//	}
//
//	return:
//	T{A: 1, B: "2"}
func ParserCommaSeparatedValues[T any](data []byte) (T, error) {
	var result T
	if len(data) == 0 {
		return result, nil
	}

	// Split by comma
	pairs := bytes.Split(data, []byte(","))

	// Create map to store key-value pairs
	values := make(map[string]string)

	// Parse each key-value pair
	for _, pair := range pairs {
		kv := bytes.Split(pair, []byte("="))
		if len(kv) != 2 {
			return result, fmt.Errorf("invalid key-value pair: %s", pair)
		}
		key := string(bytes.TrimSpace(kv[0]))
		value := string(bytes.TrimSpace(kv[1]))
		values[key] = value
	}

	// Convert map to struct using reflection
	resultValue := reflect.ValueOf(&result).Elem()
	resultType := resultValue.Type()

	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)
		fieldValue := resultValue.Field(i)

		// Get comma tag value
		tag := field.Tag.Get("comma")
		if tag == "" {
			tag = field.Name
		}

		if value, ok := values[tag]; ok {
			switch field.Type.Kind() {
			case reflect.String:
				fieldValue.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
					fieldValue.SetInt(intVal)
				}
			case reflect.Float32, reflect.Float64:
				if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					fieldValue.SetFloat(floatVal)
				}
			case reflect.Bool:
				if boolVal, err := strconv.ParseBool(value); err == nil {
					fieldValue.SetBool(boolVal)
				}
			}
		}
	}

	return result, nil
}
