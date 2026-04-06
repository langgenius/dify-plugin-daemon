package parser

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/system"
)

func MarshalPluginID(author string, name string, version string) string {
	if author == "" {
		return fmt.Sprintf("%s%s%s", name, system.DelimiterFLag, version)
	}
	return fmt.Sprintf("%s/%s%s%s", author, name, system.DelimiterFLag, version)
}
