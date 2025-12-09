package parser

import (
	"reflect"

	"github.com/fxamacker/cbor/v2"
)

var (
	_defaultCBORMapType        = reflect.TypeOf(map[string]any{})
	_defaultCBORDecoderOptions = cbor.DecOptions{
		DefaultMapType: _defaultCBORMapType,
	}
	_defaultCBORDecoder cbor.DecMode
)

func init() {
	var err error
	_defaultCBORDecoder, err = _defaultCBORDecoderOptions.DecMode()
	if err != nil {
		panic(err)
	}
}

func MarshalCBOR[T any](v T) ([]byte, error) {
	return cbor.Marshal(v)
}

func UnmarshalCBOR[T any](data []byte) (T, error) {
	var v T
	err := _defaultCBORDecoder.Unmarshal(data, &v)
	return v, err
}
