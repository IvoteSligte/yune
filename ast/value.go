package ast

import (
	"encoding/json"
	"log"
	"yune/util"

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

type Value interface {
	value()
}

func UnmarshalValue(data *fj.Value) (v Value) {
	if string(data.MarshalTo(nil)) == `"<no_value>"` {
		return
	}
	if v = UnmarshalType(data); v != nil {
		return
	}
	if v = UnmarshalExpression(data); v != nil {
		return
	}
	key, _ := fjUnmarshalUnion(data.GetObject())
	log.Fatalf("Unknown key for JSON Value: '%s'.", key)
	return
}

func Unmarshal(jsonBytes []byte) (values []Value) {
	data := fj.MustParseBytes(jsonBytes)
	return util.Map(data.GetArray(), UnmarshalValue)
}

type Destination interface {
	SetValue(v Value)
}

type SetType struct {
	Type *TypeValue
}

func (s SetType) SetValue(v Value) {
	*s.Type = v.(TypeValue)
}
