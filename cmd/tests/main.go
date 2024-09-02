package main

import (
	"encoding/json"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

const data = `name: John
age: 30
a:
  b: 2
`

type Test struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
	A    json.RawMessage
}

func main() {
	ret, err := parser.UnmarshalYamlBytes[Test]([]byte(data))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ret)
}
