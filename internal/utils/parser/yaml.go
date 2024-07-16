package parser

import (
	"gopkg.in/yaml.v3"
)

func UnmarshalYaml[T any](text string) (T, error) {
	return UnmarshalYamlBytes[T]([]byte(text))
}

func UnmarshalYamlBytes[T any](data []byte) (T, error) {
	var result T
	err := yaml.Unmarshal(data, &result)
	return result, err
}

func MarshalYaml[T any](data T) string {
	return string(MarshalYamlBytes(data))
}

func MarshalYamlBytes[T any](data T) []byte {
	b, _ := yaml.Marshal(data)
	return b
}

func UnmarshalYaml2Map(yaml []byte) (map[string]any, error) {
	return UnmarshalYamlBytes[map[string]any](yaml)
}
