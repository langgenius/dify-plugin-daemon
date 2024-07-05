package dify_invocation

func difyPath(path ...string) string {
	return baseurl.JoinPath(path...).String()
}
