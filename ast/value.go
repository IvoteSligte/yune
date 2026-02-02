package ast

import (
	"encoding/json"
	"log"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

func fjUnmarshal[T any](fjValue *fj.Value, dest T) T {
	if err := json.Unmarshal(fjValue.MarshalTo(nil), &dest); err != nil {
		log.Fatalf("Failed to unmarshal JSON: '%s'", err)
	}
	return dest
}

func fjUnmarshalUnion(data *fj.Value) (key string, value *fj.Value) {
	object := data.GetObject()
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

func UnmarshalValue(data *fj.Value) Value {
	key, v := fjUnmarshalUnion(data)
	switch key {
	case "Type":
		return UnmarshalType(v)
	case "Expression":
		return UnmarshalExpression(v)
	default:
		log.Fatalf("Unknown key for JSON Value: '%s'.", key)
	}
	panic("unreachable")
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
