package ast

import (
	"fmt"
	"log"
	"yune/util"

	"github.com/go-json-experiment/json"
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

func UnmarshalType(data *fj.Value) (t TypeValue) {
	fmt.Printf("JSON: %s\n", data.String())
	key, v := fjUnmarshalUnion(data)
	switch key {
	case "TypeType":
		t = TypeType{}
	case "IntType":
		t = IntType{}
	case "FloatType":
		t = FloatType{}
	case "BoolType":
		t = BoolType{}
	case "StringType":
		t = StringType{}
	case "NilType":
		t = NilType{}
	case "TupleType":
		t = TupleType{
			Elements: util.Map(v.Get("elements").GetArray(), UnmarshalType),
		}
	case "ListType":
		t = ListType{
			Element: UnmarshalType(v.Get("element")),
		}
	case "FnType":
		t = FnType{
			Argument: UnmarshalType(v.Get("argument")),
			Return:   UnmarshalType(v.Get("return")),
		}
	case "StructType":
		t = StructType{
			Name: string(v.GetStringBytes("name")),
		}
	default:
		log.Fatalf("Unknown key for JSON Type: '%s'.", key)
	}
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
