package real

func (r *RealBackwardsInvocation) difyPath(path ...string) string {
	path = append([]string{"inner", "api"}, path...)
	return r.dify_inner_api_baseurl.JoinPath(path...).String()
}
