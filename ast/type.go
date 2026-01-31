package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"
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

type TypeValue interface {
	Value
	typeValue()
	Lower() cpp.Type
}

type TypeType struct{}

func (t TypeType) typeValue()    {}
func (t TypeType) value()        {}
func (TypeType) Lower() cpp.Type { return "Type" }

type IntType struct{}

func (i IntType) typeValue()    {}
func (i IntType) value()        {}
func (IntType) Lower() cpp.Type { return "int" }

type FloatType struct{}

func (f FloatType) typeValue()    {}
func (f FloatType) value()        {}
func (FloatType) Lower() cpp.Type { return "float" }

type BoolType struct{}

func (b BoolType) typeValue()    {}
func (b BoolType) value()        {}
func (BoolType) Lower() cpp.Type { return "bool" }

type StringType struct{}

func (s StringType) typeValue()    {}
func (s StringType) value()        {}
func (StringType) Lower() cpp.Type { return "std::string" }

type NilType struct{}

func (n NilType) typeValue()    {}
func (n NilType) value()        {}
func (NilType) Lower() cpp.Type { return "void" }

type TupleType struct {
	Elements []TypeValue
}

func (t TupleType) typeValue() {}
func (t TupleType) value()     {}
func (t TupleType) Lower() cpp.Type {
	return cpp.Type(util.JoinFunction(t.Elements, ", ", func(v TypeValue) string {
		return v.Lower().String()
	}))
}

type ListType struct {
	Element TypeValue
}

func (l ListType) typeValue() {}
func (l ListType) value()     {}
func (l ListType) Lower() cpp.Type {
	return cpp.Type("std::vector<" + l.Element.Lower() + ">")
}

type FnType struct {
	Argument TypeValue
	Return   TypeValue
}

func (f FnType) typeValue() {}
func (f FnType) value()     {}
func (f FnType) Lower() cpp.Type {
	_return := f.Return.Lower()
	argumentTuple, argumentIsTuple := f.Argument.(TupleType)
	var arguments string
	if argumentIsTuple {
		arguments = util.JoinFunction(argumentTuple.Elements, ", ", func(v TypeValue) string {
			return v.Lower().String()
		})
	} else {
		arguments = f.Argument.Lower().String()
	}
	return cpp.Type(fmt.Sprintf("std::function<%s(%s)>", _return, arguments))
}

type StructType struct {
	Name string
}

func (s StructType) typeValue() {}
func (s StructType) value()     {}
func (s StructType) Lower() cpp.Type {
	// TODO: register struct type if newly defined
	return cpp.Type("ty::" + s.Name)
}

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
