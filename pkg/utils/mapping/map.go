package mapping

func MapArray[T any, R any](arr []T, mapFunc func(T) R) []R {
	result := make([]R, len(arr))
	for i, v := range arr {
		result[i] = mapFunc(v)
	}
	return result
}
