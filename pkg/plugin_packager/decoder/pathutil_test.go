package decoder

import "testing"

func TestNormalizeLogicalPath(t *testing.T) {
	cases := []struct {
		in      string
		out     string
		wantErr bool
	}{
		{"", "", false},
		{"   ", "", false},
		{"models\\llm\\*.yaml", "models/llm/*.yaml", false},
		{"models//llm/../llm/a.yaml", "models/llm/a.yaml", false},
		{"./models\\llm\\..\\b.yaml", "models/b.yaml", false},
		{"/absolute/path/file.yaml", "", true},
		{"C:\\models\\x.yaml", "", true},
		{"C:models\\x.yaml", "", true},
		{".\\dir\\file", "dir/file", false},
		{"../outside/file.yaml", "", true},
		{"a/..", "", true},
		{"a/./b", "a/b", false},
		{"a//b///c", "a/b/c", false},
	}

	for i, c := range cases {
		got, err := normalizeLogicalPath(c.in)
		if c.wantErr {
			if err == nil {
				t.Fatalf("case %d: expected error for %q, got nil with %q", i, c.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("case %d: unexpected error for %q: %v", i, c.in, err)
		}
		if got != c.out {
			t.Fatalf("case %d: normalizeLogicalPath(%q) = %q, want %q", i, c.in, got, c.out)
		}
	}
}
