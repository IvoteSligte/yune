package util

import (
	"fmt"
	"strings"
)

func Map[T, V any](slice []T, function func(T) V) []V {
	result := make([]V, len(slice))
	for i, t := range slice {
		result[i] = function(t)
	}
	return result
}

func Map2[T, V1, V2 any](slice []T, function func(T) (V1, V2)) ([]V1, []V2) {
	result1 := make([]V1, len(slice))
	result2 := make([]V2, len(slice))
	for i, t := range slice {
		result1[i], result2[i] = function(t)
	}
	return result1, result2
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

func SeparatedBy[T fmt.Stringer](array []T, separator string) string {
	return strings.Join(Map(array, T.String), separator)
}

func SliceEqual[T comparable](left []T, right []T) bool {
	if len(left) != len(right) {
		return false
	}
	for i, l := range left {
		if l != right[i] {
			return false
		}
	}
	return true
}
