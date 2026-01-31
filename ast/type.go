package ast

import (
	"yune/cpp"
)

// TODO: macros in types

type Type struct {
	// Evaluated expression
	value      TypeValue
	Expression Expression
}

// TODO: rename to GetValue
func (t Type) Get() TypeValue {
	return t.value
}

func (t Type) Lower() cpp.Type {
	return t.value.Lower()
}

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

type TypeValue interface {
	Value
	typeValue()
}

type TypeType struct{}

func (t TypeType) typeValue() {}
func (t TypeType) value()     {}

type IntType struct{}

func (i IntType) typeValue() {}
func (i IntType) value()     {}

type FloatType struct{}

func (f FloatType) typeValue() {}
func (f FloatType) value()     {}

type BoolType struct{}

func (b BoolType) typeValue() {}
func (b BoolType) value()     {}

type StringType struct{}

func (s StringType) typeValue() {}
func (s StringType) value()     {}

type NilType struct{}

func (n NilType) typeValue() {}
func (n NilType) value()     {}

type TupleType struct {
	Elements []TypeValue
}

func (t TupleType) typeValue() {}
func (t TupleType) value()     {}

type ListType struct {
	Element TypeValue
}

func (l ListType) typeValue() {}
func (l ListType) value()     {}

type FnType struct {
	Argument TypeValue
	Return   TypeValue
}

func (f FnType) typeValue() {}
func (f FnType) value()     {}

type StructType struct {
	Name string
}

func (s StructType) typeValue() {}
func (s StructType) value()     {}

var _ TypeValue = TypeType{}
var _ TypeValue = IntType{}
var _ TypeValue = FloatType{}
var _ TypeValue = BoolType{}
var _ TypeValue = StringType{}
var _ TypeValue = NilType{}
var _ TypeValue = TupleType{}
var _ TypeValue = ListType{}
var _ TypeValue = FnType{}
var _ TypeValue = StructType{}
