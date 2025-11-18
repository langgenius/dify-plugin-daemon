package parser

func Map[From any, To any](f func(From) To, arr []From) []To {
	var result []To
	for _, v := range arr {
		result = append(result, f(v))
	}
	return result
}
