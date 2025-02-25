package parser

import "github.com/fxamacker/cbor/v2"

func MarshalCBOR[T any](v T) ([]byte, error) {
	return cbor.Marshal(v)
}

func UnmarshalCBOR[T any](data []byte) (T, error) {
	var v T
	err := cbor.Unmarshal(data, &v)
	return v, err
}
