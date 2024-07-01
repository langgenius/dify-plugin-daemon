package parser

import "encoding/json"

func UnmarshalJson[T any](text string) (T, error) {
	var result T
	err := json.Unmarshal([]byte(text), &result)
	return result, err
}

func MarshalJson[T any](data T) string {
	b, _ := json.Marshal(data)
	return string(b)
}
