package mapping

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestConvertAnyMap(t *testing.T) {
	src := map[any]any{
		"a": 1,
		"b": 2,
		"c": map[any]any{
			"d": 3,
			"e": 4,
			"f": []any{
				1,
				2,
				3,
			},
		},
	}

	dst := ConvertAnyMap(src).(map[string]any)

	// use reflect to check key is string
	keys := reflect.ValueOf(dst).MapKeys()
	for _, key := range keys {
		if key.Kind() != reflect.String {
			t.Errorf("key is not string: %v", key)
		}
	}

	c := dst["c"]
	keys = reflect.ValueOf(c).MapKeys()
	for _, key := range keys {
		if key.Kind() != reflect.String {
			t.Errorf("key is not string: %v", key)
		}
	}

	f := dst["c"].(map[string]any)["f"]
	if _, ok := f.([]any); !ok {
		t.Errorf("f is not []any: %v", f)
	}

	if len(f.([]any)) != 3 {
		t.Errorf("f is not 3: %v", f)
	}

	if f.([]any)[0] != 1 {
		t.Errorf("f[0] is not 1: %v", f)
	}
}

func TestConvertAnyMap_JsonMarshal(t *testing.T) {
	src := map[any]any{
		"a": 1,
		"b": 2,
		"c": map[any]any{
			"d": 3,
			"e": 4,
			"f": []any{
				1,
				2,
				3,
			},
		},
	}

	dst := ConvertAnyMap(src).(map[string]any)

	_, err := json.Marshal(dst)
	if err != nil {
		t.Errorf("json.Marshal error: %v", err)
	}
}
