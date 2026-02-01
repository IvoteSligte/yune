package ast

import (
	"github.com/go-json-experiment/json"
	"yune/oneof"
)

var valueOptions = []oneof.Option{
	TypeType{},
	IntType{},
	FloatType{},
	BoolType{},
	StringType{},
	NilType{},
	TupleType{},
	ListType{},
	FnType{},
	StructType{},
}

type Value interface {
	oneof.Option
	value()
}

func Deserialize(jsonBytes []byte) (vs []Value) {
	unmarshaler := oneof.UnmarshalFunc[Value](valueOptions)
	err := json.Unmarshal(jsonBytes, &vs, json.WithUnmarshalers(unmarshaler))
	if err != nil {
		panic("Failed to deserialize JSON: " + err.Error())
	}
	return
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
