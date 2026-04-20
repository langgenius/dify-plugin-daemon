package parser

import (
	"reflect"
	"testing"
)

func TestSplitAndTrimCSV(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty string", "", nil},
		{"only separators", ",,,", []string{}},
		{"only whitespace between separators", " , , ,", []string{}},
		{"single value", "a", []string{"a"}},
		{"trims surrounding whitespace", " a , b ", []string{"a", "b"}},
		{"drops empty segments", "a, ,b", []string{"a", "b"}},
		{"preserves inner colons", "host1:6379, host2:6379", []string{"host1:6379", "host2:6379"}},
		{"trailing comma", "a,b,", []string{"a", "b"}},
		{"leading comma", ",a,b", []string{"a", "b"}},
		{"tabs and mixed whitespace", "\ta\t,\nb\n", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitAndTrimCSV(tt.input)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitAndTrimCSV(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}
