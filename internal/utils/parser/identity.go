package parser

import "fmt"

func MarshalPluginUniqueIdentifier(name string, version string) string {
	return fmt.Sprintf("%s:%s", name, version)
}
