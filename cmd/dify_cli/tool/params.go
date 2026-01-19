package tool

import (
	"fmt"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
)

func ParseArgs(tool *types.DifyToolDeclaration, args []string) map[string]any {
	params := make(map[string]any)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			continue
		}

		name := strings.TrimPrefix(arg, "--")
		if i+1 >= len(args) {
			continue
		}

		value := args[i+1]
		i++

		for _, p := range tool.Parameters {
			if p.Name == name {
				switch p.Type {
				case types.ToolParameterTypeNumber:
					var num float64
					fmt.Sscanf(value, "%f", &num)
					params[name] = num
				case types.ToolParameterTypeBoolean:
					params[name] = value == "true" || value == "1"
				default:
					params[name] = value
				}
				break
			}
		}

		if _, exists := params[name]; !exists {
			params[name] = value
		}
	}

	return params
}
