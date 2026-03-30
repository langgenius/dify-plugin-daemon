package parser

import "fmt"

func MarshalPluginID(author string, name string, version string) string {
	if author == "" {
		return fmt.Sprintf("%s:%s", name, version)
	}
	return fmt.Sprintf("%s/%s:%s", author, name, version)
}
