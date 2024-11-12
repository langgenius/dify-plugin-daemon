package real

func (r *RealBackwardsInvocation) difyPath(path ...string) string {
	path = append([]string{"inner", "api"}, path...)
	return r.difyInnerApiBaseurl.JoinPath(path...).String()
}
