package parser

import "encoding/json"

func UnmarshalJson[T any](text string) (T, error) {
	return UnmarshalJsonBytes[T]([]byte(text))
}

func UnmarshalJsonBytes[T any](data []byte) (T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	return result, err
}

func MarshalJson[T any](data T) string {
	b, _ := json.Marshal(data)
	return string(b)
}

func MarshalJsonBytes[T any](data T) []byte {
	b, _ := json.Marshal(data)
	return b
}
