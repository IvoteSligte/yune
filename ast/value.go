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

func fjUnmarshalStruct(object *fj.Object) (key string, value *fj.Value, generic *fj.Value) {
	fields := fjGetFields(object)
	if object.Len() == 2 {
		var hasGeneric bool
		generic, hasGeneric = fields["generic_"]

		if !hasGeneric {
			log.Panicf("Found 2 fields when deserializing JSON as struct, but no 'generic_' parameter. JSON: %s", object)
		}
		delete(fields, "generic_")
	} else if len(fields) != 1 {
		log.Panicf("Found %d fields when deserializing JSON as struct. Expected 1 and an optional 'generic_' parameter. JSON: %s", object.Len(), object)
		return
	}
	// iterate over single value
	for key, value := range fields {
		return key, value, nil
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

func UnmarshalList(data *fj.Value, keys ...string) []*fj.Value {
	// ignoring generic because the users of this function ignore it
	key, value, _ := fjUnmarshalStruct(UnmarshalItem(data, (*fj.Value).Object, keys...))
	if key != "List" {
		panic(fmt.Sprintf("Unmarshalled struct that is not a list from data: %s", data))
	}
	return UnmarshalArray(value)
}
