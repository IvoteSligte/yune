package ast

import (
	"fmt"
	"log"
	"strings"

	fj "github.com/valyala/fastjson"
)

func fjGetFields(object *fj.Object) (fields map[string]*fj.Value) {
	fields = map[string]*fj.Value{}
	object.Visit(func(byteKey []byte, v *fj.Value) {
		fields[string(byteKey)] = v
	})
	if object.Len() != len(fields) {
		log.Panicf("Duplicate field in JSON object: %s", object)
	}
	return fields
}

func fjUnmarshalStruct(object *fj.Object) (key string, value *fj.Value) {
	fields := fjGetFields(object)
	if len(fields) != 1 {
		log.Panicf("Found %d fields when deserializing JSON as struct. Expected 1. JSON: %s", object.Len(), object)
		return
	}
	// iterate over single value
	for key, value := range fields {
		return key, value
	}
	return
}

func UnmarshalItem[T any](data *fj.Value, f func(*fj.Value) (T, error), keys ...string) T {
	value := data.Get(keys...)
	if value == nil {
		panic(fmt.Sprintf("Key path '[%s]' does not exist in data: %s", strings.Join(keys, ", "), data))
	}
	item, err := f(value)
	if err != nil {
		panic(fmt.Sprintf("Item is not of the desired type. Error: %s. Value: %s", err, value))
	}
	return item
}

func UnmarshalInt(data *fj.Value, keys ...string) int64 {
	return UnmarshalItem(data, (*fj.Value).Int64)
}

func UnmarshalFloat(data *fj.Value, keys ...string) float64 {
	return UnmarshalItem(data, (*fj.Value).Float64)
}

func UnmarshalBool(data *fj.Value, keys ...string) bool {
	return UnmarshalItem(data, (*fj.Value).Bool)
}

func UnmarshalString(data *fj.Value, keys ...string) string {
	return string(UnmarshalItem(data, (*fj.Value).StringBytes, keys...))
}

func UnmarshalNonEmptyString(data *fj.Value, keys ...string) string {
	s := UnmarshalString(data, keys...)
	if s == "" {
		panic(fmt.Sprintf("Unmarshalled empty string from data: %s", data))
	}
	return s
}

func UnmarshalArray(data *fj.Value, keys ...string) []*fj.Value {
	return UnmarshalItem(data, (*fj.Value).Array, keys...)
}

func UnmarshalTuple(data *fj.Value) []*fj.Value {
	key, v := fjUnmarshalStruct(data.GetObject())
	if key != "Tuple" {
		log.Panicf("Expected Tuple, but found JSON variant: %s", data)
	}
	return UnmarshalArray(v, "elements")
}

func UnmarshalSpan(data *fj.Value, in *Macro) Span {
	v := data.Get("location")
	if v == nil {
		log.Panicf("JSON does not have the expected 'location' field. JSON: %s\n", data)
	}
	location, err := v.Int()
	if err != nil {
		log.Panicf("Failed to extract location from JSON. Error: `%s`. JSON: %s\n", err, data)
	}
	macroText := in.GetText()
	if location >= len(macroText) {
		return in.Lines[0].Span
	}
	// within the macro
	lineNumber := strings.Count(macroText[:location], "\n")
	lineStart := strings.LastIndex(macroText[:location], "\n")
	column := location - lineStart

	// in the whole file
	column = in.Lines[lineNumber].Span.Column + column
	lineNumber = in.Span.Line + lineNumber

	return Span{
		File:   in.Span.File,
		Source: in.Span.Source,
		Line:   lineNumber,
		Column: column,
		Length: 0,
	}
}
