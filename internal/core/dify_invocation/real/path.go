package real

func (r *RealBackwardsInvocation) difyPath(path ...string) string {
	path = append([]string{"inner", "api"}, path...)
	return r.baseurl.JoinPath(path...).String()
}
