package parser

func MarshalPluginID(author string, name string, version string) string {
	if author == "" {
		return name + ":" + version
	}
	return author + "/" + name + ":" + version
}
