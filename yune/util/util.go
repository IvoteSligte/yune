package util

func Map[T, V any](slice []T, function func(T) V) []V {
	result := make([]V, len(slice))
	for i, t := range slice {
		result[i] = function(t)
	}
	return result
}

func FlatMap[T, V any](slice []T, function func(T) []V) []V {
	result := make([]V, len(slice))
	for _, t := range slice {
		result = append(result, function(t)...)
	}
	return result
}

func Prepend[T any](element T, slice []T) []T {
	return append([]T{element}, slice...)
}
