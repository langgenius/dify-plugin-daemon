package parser

import "fmt"

func MarshalPluginID(name string, version string) string {
	return fmt.Sprintf("%s:%s", name, version)
}
