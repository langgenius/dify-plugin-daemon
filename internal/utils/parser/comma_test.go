package parser

import (
	"testing"
)

func TestCommaSeparatedValues(t *testing.T) {
	type testStruct struct {
		A        int     `comma:"a"`
		B        string  `comma:"b"`
		C        bool    `comma:"c"`
		D        float64 `comma:"d"`
		E        string  // no tag
		Endpoint string  `comma:"endpoint"`
		Name     string  `comma:"name"`
		ID       string  `comma:"id"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    testStruct
		wantErr bool
	}{
		{
			name:  "basic test",
			input: []byte("a=1,b=hello,c=true,d=3.14,E=world"),
			want: testStruct{
				A: 1,
				B: "hello",
				C: true,
				D: 3.14,
				E: "world",
			},
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   []byte(""),
			want:    testStruct{},
			wantErr: false,
		},
		{
			name:  "partial fields",
			input: []byte("a=42,c=false"),
			want: testStruct{
				A: 42,
				C: false,
			},
			wantErr: false,
		},
		{
			name:  "extra fields ignored",
			input: []byte("a=1,b=test,extra=ignored"),
			want: testStruct{
				A: 1,
				B: "test",
			},
			wantErr: false,
		},
		{
			name:    "invalid format - missing value",
			input:   []byte("a=1,b"),
			wantErr: true,
		},
		{
			name:    "invalid format - missing equals",
			input:   []byte("a=1,b:2"),
			wantErr: true,
		},
		{
			name:  "whitespace handling",
			input: []byte("a = 1, b = hello "),
			want: testStruct{
				A: 1,
				B: "hello",
			},
			wantErr: false,
		},
		{
			name:  "type conversion - invalid int",
			input: []byte("a=notanumber,b=test"),
			want: testStruct{
				B: "test",
			},
			wantErr: false,
		},
		{
			name:  "type conversion - invalid bool",
			input: []byte("c=notabool,b=test"),
			want: testStruct{
				B: "test",
			},
			wantErr: false,
		},
		{
			name:  "type conversion - invalid float",
			input: []byte("d=notafloat,b=test"),
			want: testStruct{
				B: "test",
			},
			wantErr: false,
		},
		{
			name:  "actual example",
			input: []byte("endpoint=test-plugin.dify.uwu,name=c31a98df2ef139d6532d8da8caa2bb63,id=c31a98df2ef139d6532d8da8caa2bb63"),
			want: testStruct{
				Endpoint: "test-plugin.dify.uwu",
				Name:     "c31a98df2ef139d6532d8da8caa2bb63",
				ID:       "c31a98df2ef139d6532d8da8caa2bb63",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParserCommaSeparatedValues[testStruct](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParserCommaSeparatedValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParserCommaSeparatedValues() = %v, want %v", got, tt.want)
			}
		})
	}

}
