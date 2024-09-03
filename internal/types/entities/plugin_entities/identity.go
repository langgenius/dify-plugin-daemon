package plugin_entities

import "strings"

type PluginIdentity string

func (p PluginIdentity) PluginID() string {
	// try find @
	split := strings.Split(p.String(), "@")
	if len(split) == 2 {
		return split[0]
	}
	return p.String()
}

func (p PluginIdentity) String() string {
	return string(p)
}
