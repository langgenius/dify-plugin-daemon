package mapping

import "fmt"

// ConvertAnyMap converts a map[any]any to a map[string]any
// please make sure i is a map[any]any
func ConvertAnyMap(i any) any {
	switch v := i.(type) {
	case map[any]any:
		m2 := make(map[string]any)
		for k, val := range v {
			keyStr := fmt.Sprintf("%v", k)
			m2[keyStr] = ConvertAnyMap(val)
		}
		return m2
	case map[string]any:
		m2 := make(map[string]any)
		for k, val := range v {
			m2[k] = ConvertAnyMap(val)
		}
		return m2
	case []any:
		m2 := make([]any, len(v))
		for i, val := range v {
			m2[i] = ConvertAnyMap(val)
		}
		return m2
	default:
		return v
	}
}
