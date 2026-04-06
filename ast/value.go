package ast

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	fj "github.com/valyala/fastjson"
)

func fjUnmarshal[T any](v *fj.Value, dest T) T {
	if v == nil {
		return dest
	}
	if err := json.Unmarshal(v.MarshalTo(nil), &dest); err != nil {
		log.Panicf("Failed to unmarshal JSON: '%s'", err)
	}
	return dest
}

type fjField = struct {
	key   string
	value *fj.Value
}

func fjGetFields(object *fj.Object) (fields []fjField) {
	object.Visit(func(byteKey []byte, v *fj.Value) {
		fields = append(fields, fjField{
			key:   string(byteKey),
			value: v,
		})
	})
	return fields
}

func fjUnmarshalUnion(object *fj.Object) (key string, value *fj.Value) {
	if object.Len() != 1 {
		log.Panicf("Found %d keys when deserializing JSON union: '%v'. Expected 1.", object.Len(), object)
		return
	}
	fields := fjGetFields(object)
	key = fields[0].key
	value = fields[0].value
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
