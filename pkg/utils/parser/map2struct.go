package parser

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
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

	if err := validators.GlobalEntitiesValidator.Struct(s); err != nil {
		return nil, fmt.Errorf("error validating struct: %s", err)
	}

	return &s, nil
}
