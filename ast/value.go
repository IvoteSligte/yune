package ast

import (
	"encoding/json"
	"log"

	fj "github.com/valyala/fastjson"
)

func fjUnmarshal[T any](fjValue *fj.Value, dest T) T {
	if fjValue == nil {
		return dest
	}
	if err := json.Unmarshal(fjValue.MarshalTo(nil), &dest); err != nil {
		log.Fatalf("Failed to unmarshal JSON: '%s'", err)
	}
	return dest
}

func fjUnmarshalUnion(object *fj.Object) (key string, value *fj.Value) {
	if object.Len() != 1 {
		log.Fatalf("Found %d keys when deserializing JSON union: '%v'. Expected 1.", object.Len(), object)
		return
	}
	object.Visit(func(byteKey []byte, v *fj.Value) {
		key = string(byteKey)
		value = v
	})
	return
}
