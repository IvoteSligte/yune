package util

import (
	"encoding/json"
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

func MapMap[Map ~map[K1]V1, K1, K2 comparable, V1, V2 any](m Map, function func(K1, V1) (K2, V2)) map[K2]V2 {
	result := make(map[K2]V2, len(m))
	for k1, v1 := range m {
		k2, v2 := function(k1, v1)
		result[k2] = v2
	}
	return result
}

func FlatMap[T, V any](slice []T, function func(T) []V) []V {
	result := make([]V, 0, len(slice))
	for _, t := range slice {
		result = append(result, function(t)...)
	}
	return result
}

func FlatMapPtr[T, V any](slice []T, function func(*T) []V) []V {
	result := make([]V, 0, len(slice))
	for i := range slice {
		result = append(result, function(&slice[i])...)
	}
	return result
}

func Prepend[T any](element T, slice []T) []T {
	return append([]T{element}, slice...)
}

func JoinFunction[T any](array []T, separator string, function func(T) string) string {
	return strings.Join(Map(array, function), separator)
}

func Join[T fmt.Stringer](array []T, separator string) string {
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

func Any[T any](slice []T, f func(T) bool) bool {
	return slices.ContainsFunc(slice, f)
}

func All[T any](slice []T, f func(T) bool) bool {
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

func PrettyPrint(bs ...any) {
	for _, b := range bs {
		s, _ := json.MarshalIndent(b, "", "   ")
		fmt.Println(string(s))
	}
}

func FirstError(s ...error) error {
	for _, t := range s {
		if t != nil {
			return t
		}
	}
	return nil
}
