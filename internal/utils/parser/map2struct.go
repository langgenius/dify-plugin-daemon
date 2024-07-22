package parser

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

func MapToStruct[T any](m map[string]any) (*T, error) {
	var s T
	decoder := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &s,
		TagName:  "json",
		Squash:   true,
	}

	d, err := mapstructure.NewDecoder(decoder)
	if err != nil {
		return nil, fmt.Errorf("error creating decoder: %s", err)
	}

	err = d.Decode(m)
	if err != nil {
		return nil, fmt.Errorf("error decoding map: %s", err)
	}

	return &s, nil
}
