package parser

import "fmt"

func MarshalPluginIdentity(name string, version string) string {
	return fmt.Sprintf("%s:%s", name, version)
}
