package plugin_entities

import "strings"

type PluginUniqueIdentifier string

func (p PluginUniqueIdentifier) PluginID() string {
	// try find @
	split := strings.Split(p.String(), "@")
	if len(split) == 2 {
		return split[0]
	}
	return p.String()
}

func (p PluginUniqueIdentifier) String() string {
	return string(p)
}
