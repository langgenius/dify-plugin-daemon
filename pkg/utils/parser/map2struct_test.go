package parser

import "testing"

func TestMapToStruct(t *testing.T) {
	m := map[string]any{
		"result": "result",
		"inherit": map[string]any{
			"inherit_result": "result",
		},
		"object": map[string]any{
			"a": 1,
		},
	}

	type p struct {
		Inherit struct {
			InheritResult string `json:"inherit_result"`
		}
	}

	type s struct {
		p

		Result string `json:"result"`
		Object struct {
			A int `json:"a"`
		} `json:"object"`
	}

	result, err := MapToStruct[s](m)
	if err != nil {
		t.Error(err)
	}

	if result.Result != "result" {
		t.Error("result should be result")
	}

	if result.Inherit.InheritResult != "result" {
		t.Error("inherit_result should be result")
	}

	if result.Object.A != 1 {
		t.Error("a should be 1")
	}

}
