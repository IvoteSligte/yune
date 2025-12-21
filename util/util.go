package util

import (
	"fmt"
	"iter"
	"slices"
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

func Identity[T any](value T) T {
	return value
}

func Any[T any](f func(T) bool, slice ...T) bool {
	return slices.ContainsFunc(slice, f)
}

func All[T any](f func(T) bool, slice ...T) bool {
	for _, elem := range slice {
		if !f(elem) {
			return false
		}
	}
	return true
}

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](values ...T) Set[T] {
	set := Set[T]{}

	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func Contains[Map ~map[K]V, K comparable, V any](m Map, keys ...K) bool {
	for _, key := range keys {
		_, ok := m[key]
		if !ok {
			return false
		}
	}
	return true
}

func Remove[Map ~map[K]V, K comparable, V any](m Map, keys ...K) {
	for _, key := range keys {
		delete(m, key)
	}
}

func TakeExisting[Map ~map[K]V, K comparable, V any](m Map, keys ...K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, key := range keys {
			value, ok := m[key]
			if !ok {
				continue
			}
			if !yield(key, value) {
				return
			}
		}
	}
}
