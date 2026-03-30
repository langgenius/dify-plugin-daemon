package parser

import (
	"bytes"
	"encoding/gob"
)

func MarshalGob(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func UnmarshalGob[T any](data []byte) (T, error) {
	var v T
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}
