package ast

import (
	"fmt"
	"log"
	"yune/cpp"
	"yune/util"

	fj "github.com/valyala/fastjson"
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

// Default value for TypeValue that still allows method calls.
var noType = DefaultTypeValue{}

func UnmarshalType(data *fj.Value) (t TypeValue) {
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

// Wraps self in a TupleType, if self is not already a TupleType
func wrapTupleType(t TypeValue) TupleType {
	tupleType, ok := t.(TupleType)
	if ok {
		return tupleType
	}
	return NewTupleType(t)
}

// If self is a TupleType of one element, returns the first element. Otherwise returns self.
func unwrapTupleType(t TypeValue) TypeValue {
	tupleType, ok := t.(TupleType)
	if ok {
		return tupleType.Elements[0]
	}
	return t
}

type TypeValue interface {
	Value
	fmt.Stringer
	typeValue()
	Lower() cpp.Type
	Eq(other TypeValue) bool
}

type DefaultTypeValue struct{}

func (DefaultTypeValue) value()     {}
func (DefaultTypeValue) typeValue() {}
func (DefaultTypeValue) String() string {
	panic("DefaultTypeValue.String should be overridden")
}
func (DefaultTypeValue) Lower() cpp.Type {
	panic("DefaultTypeValue.Lower should be overridden")
}
func (DefaultTypeValue) Eq(other TypeValue) bool {
	return false // allows using DefaultTypeValue{} as default Expression value
}

var _ TypeValue = DefaultTypeValue{}

type TypeType struct{ DefaultTypeValue }

func (TypeType) String() string { return "Type" }
func (t TypeType) Eq(other TypeValue) bool {
	_, ok := other.(TypeType)
	return ok
}
func (TypeType) Lower() cpp.Type { return "ty::Type" }

type IntType struct{ DefaultTypeValue }

func (IntType) String() string { return "Int" }
func (i IntType) Eq(other TypeValue) bool {
	_, ok := other.(IntType)
	return ok
}
func (IntType) Lower() cpp.Type            { return "int" }
func (t IntType) WrapTupleType() TupleType { return NewTupleType(t) }

type FloatType struct{ DefaultTypeValue }

func (FloatType) String() string { return "Float" }
func (f FloatType) Eq(other TypeValue) bool {
	_, ok := other.(FloatType)
	return ok
}
func (FloatType) Lower() cpp.Type { return "float" }

type BoolType struct{ DefaultTypeValue }

func (BoolType) String() string { return "Bool" }
func (b BoolType) Eq(other TypeValue) bool {
	_, ok := other.(BoolType)
	return ok
}
func (BoolType) Lower() cpp.Type { return "bool" }

type StringType struct{ DefaultTypeValue }

func (StringType) String() string { return "String" }
func (s StringType) Eq(other TypeValue) bool {
	_, ok := other.(StringType)
	return ok
}
func (StringType) Lower() cpp.Type { return "std::string" }

type NilType struct{ DefaultTypeValue }

func (NilType) String() string { return "Nil" }
func (n NilType) Eq(other TypeValue) bool {
	_, ok := other.(NilType)
	return ok
}
func (NilType) Lower() cpp.Type { return "void" }

type TupleType struct {
	DefaultTypeValue
	Elements []TypeValue
}

func (t TupleType) String() string {
	return "(" + util.Join(t.Elements, ", ") + ")"
}

func (t TupleType) Eq(other TypeValue) bool {
	otherTuple, ok := other.(TupleType)
	if !ok || len(t.Elements) != len(otherTuple.Elements) {
		return false
	}
	for i, element := range t.Elements {
		if !element.Eq(otherTuple.Elements[i]) {
			return false
		}
	}
	return true
}
func (t TupleType) Lower() cpp.Type {
	return cpp.Type("std::tuple<" + util.JoinFunction(t.Elements, ", ", func(v TypeValue) string {
		return v.Lower().String()
	}) + ">")
}

func NewTupleType(elements ...TypeValue) TupleType {
	return TupleType{
		Elements: elements,
	}
}

type ListType struct {
	DefaultTypeValue
	Element TypeValue
}

func (l ListType) String() string {
	return "List(" + l.Element.String() + ")"
}

func (l ListType) Eq(other TypeValue) bool {
	otherList, ok := other.(ListType)
	return ok && l.Element.Eq(otherList.Element)
}
func (l ListType) Lower() cpp.Type {
	return cpp.Type("std::vector<" + l.Element.Lower() + ">")
}

type FnType struct {
	DefaultTypeValue
	Argument TypeValue
	Return   TypeValue
}

func (f FnType) String() string {
	return "Fn(" + f.Argument.String() + ", " + f.Return.String() + ")"
}

func (f FnType) Eq(other TypeValue) bool {
	otherFn, ok := other.(FnType)
	return ok && f.Argument.Eq(otherFn.Argument) && f.Return.Eq(otherFn.Return)
}
func (f FnType) Lower() cpp.Type {
	_return := f.Return.Lower()
	arguments := util.JoinFunction(wrapTupleType(f.Argument).Elements, ", ", func(v TypeValue) string {
		return v.Lower().String()
	})
	return cpp.Type(fmt.Sprintf("std::function<%s(%s)>", _return, arguments))
}

type StructType struct {
	DefaultTypeValue
	Name string
}

func (s StructType) String() string {
	return s.Name
}

func (s StructType) Eq(other TypeValue) bool {
	otherStruct, ok := other.(StructType)
	return ok && s.Name == otherStruct.Name
}
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
