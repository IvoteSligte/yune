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

var MainType = FnType{
	Argument: NewTupleType(IntType{}),
	Return:   NilType{},
}
var ExpressionType = StructType{Name: "Expression"}
var MacroReturnType = NewTupleType(StringType{}, ExpressionType)

type TypeValue interface {
	Value
	typeValue()
	Lower() cpp.Type
	Eq(other TypeValue) bool
}

type TypeType struct{}

func (t TypeType) Eq(other TypeValue) bool {
	_, ok := other.(TypeType)
	return ok
}
func (t TypeType) typeValue()    {}
func (t TypeType) value()        {}
func (TypeType) Lower() cpp.Type { return "Type" }

type IntType struct{}

func (i IntType) Eq(other TypeValue) bool {
	_, ok := other.(IntType)
	return ok
}
func (i IntType) typeValue()    {}
func (i IntType) value()        {}
func (IntType) Lower() cpp.Type { return "int" }

type FloatType struct{}

func (f FloatType) Eq(other TypeValue) bool {
	_, ok := other.(FloatType)
	return ok
}
func (f FloatType) typeValue()    {}
func (f FloatType) value()        {}
func (FloatType) Lower() cpp.Type { return "float" }

type BoolType struct{}

func (b BoolType) Eq(other TypeValue) bool {
	_, ok := other.(BoolType)
	return ok
}
func (b BoolType) typeValue()    {}
func (b BoolType) value()        {}
func (BoolType) Lower() cpp.Type { return "bool" }

type StringType struct{}

func (s StringType) Eq(other TypeValue) bool {
	_, ok := other.(StringType)
	return ok
}
func (s StringType) typeValue()    {}
func (s StringType) value()        {}
func (StringType) Lower() cpp.Type { return "std::string" }

type NilType struct{}

func (n NilType) Eq(other TypeValue) bool {
	_, ok := other.(NilType)
	return ok
}
func (n NilType) typeValue()    {}
func (n NilType) value()        {}
func (NilType) Lower() cpp.Type { return "void" }

type TupleType struct {
	Elements []TypeValue
}

func (t TupleType) Eq(other TypeValue) bool {
	otherTuple, ok := other.(TupleType)
	if !ok {
		return false
	}
	for i, element := range t.Elements {
		if !element.Eq(otherTuple.Elements[i]) {
			return false
		}
	}
	return true
}
func (t TupleType) typeValue() {}
func (t TupleType) value()     {}
func (t TupleType) Lower() cpp.Type {
	return cpp.Type(util.JoinFunction(t.Elements, ", ", func(v TypeValue) string {
		return v.Lower().String()
	}))
}

func NewTupleType(elements ...TypeValue) TupleType {
	return TupleType{
		Elements: elements,
	}
}

type ListType struct {
	Element TypeValue
}

func (l ListType) Eq(other TypeValue) bool {
	otherList, ok := other.(ListType)
	return ok && l.Element.Eq(otherList.Element)
}
func (l ListType) typeValue() {}
func (l ListType) value()     {}
func (l ListType) Lower() cpp.Type {
	return cpp.Type("std::vector<" + l.Element.Lower() + ">")
}

type FnType struct {
	Argument TupleType
	Return   TypeValue
}

func (f FnType) Eq(other TypeValue) bool {
	otherFn, ok := other.(FnType)
	return ok && f.Argument.Eq(otherFn.Argument) && f.Return.Eq(otherFn.Return)
}
func (f FnType) typeValue() {}
func (f FnType) value()     {}
func (f FnType) Lower() cpp.Type {
	_return := f.Return.Lower()
	arguments := util.JoinFunction(f.Argument.Elements, ", ", func(v TypeValue) string {
		return v.Lower().String()
	})
	return cpp.Type(fmt.Sprintf("std::function<%s(%s)>", _return, arguments))
}

type StructType struct {
	Name string
}

func (s StructType) Eq(other TypeValue) bool {
	otherStruct, ok := other.(StructType)
	return ok && s.Name == otherStruct.Name
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
