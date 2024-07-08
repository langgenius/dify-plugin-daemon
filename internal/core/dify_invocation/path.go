package dify_invocation

func difyPath(path ...string) string {
	path = append([]string{"inner", "api"}, path...)
	return baseurl.JoinPath(path...).String()
}
