package slim

import "encoding/json"

var extractCleanupKeys = map[string]struct{}{
	"background":        {},
	"created_at":        {},
	"description":       {},
	"help":              {},
	"human_description": {},
	"icon":              {},
	"icon_dark":         {},
	"icon_large":        {},
	"icon_large_dark":   {},
	"icon_small":        {},
	"icon_small_dark":   {},
	"label":             {},
	"llm_description":   {},
	"placeholder":       {},
	"privacy":           {},
	"repo":              {},
	"tags":              {},
	"verified":          {},
	"verification":      {},
}

func BuildExtractData(result ExtractResult) map[string]any {
	var raw map[string]any
	b, err := json.Marshal(result)
	if err != nil {
		return map[string]any{}
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return map[string]any{}
	}
	cleanupExtractValue(raw)
	return raw
}

func cleanupExtractValue(v any) any {
	switch value := v.(type) {
	case map[string]any:
		for key, child := range value {
			if _, ok := extractCleanupKeys[key]; ok {
				delete(value, key)
				continue
			}
			cleaned := cleanupExtractValue(child)
			if isEmptyExtractValue(cleaned) {
				delete(value, key)
				continue
			}
			value[key] = cleaned
		}
		return value
	case []any:
		out := value[:0]
		for _, item := range value {
			cleaned := cleanupExtractValue(item)
			if !isEmptyExtractValue(cleaned) {
				out = append(out, cleaned)
			}
		}
		return out
	default:
		return value
	}
}

func isEmptyExtractValue(v any) bool {
	switch value := v.(type) {
	case nil:
		return true
	case string:
		return value == ""
	case []any:
		return len(value) == 0
	case map[string]any:
		return len(value) == 0
	default:
		return false
	}
}
