package ast

import (
	"github.com/dhoelle/oneof"
	"github.com/go-json-experiment/json" // JSON V2
)

var valueOptions = map[string]Value{
	"Type":       TypeType{},
	"IntType":    IntType{},
	"FloatType":  FloatType{},
	"BoolType":   BoolType{},
	"StringType": StringType{},
	"NilType":    NilType{},
	"TupleType":  TupleType{},
	"ListType":   ListType{},
	"FnType":     FnType{},
	"StructType": StructType{},
}

type Value interface {
	value()
}

func Deserialize(jsonBytes []byte) (v Value) {
	unmarshaler := oneof.UnmarshalFunc(valueOptions, nil)
	err := json.Unmarshal(jsonBytes, &v, json.WithUnmarshalers(unmarshaler))
	if err != nil {
		panic("Failed to deserialize JSON: " + err.Error())
	}
	return
}

type Destination interface {
	SetValue(v Value)
}
