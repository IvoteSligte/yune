package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

var MainType = &FnType{
	Argument: &TupleType{},
	Return:   &TupleType{},
}

// kinds of Expression types
var IntegerLiteralType = &StructType{Name: "IntegerLiteral"}
var FloatLiteralType = &StructType{Name: "FloatLiteral"}
var BoolLiteralType = &StructType{Name: "BoolLiteral"}
var StringLiteralType = &StructType{Name: "StringLiteral"}
var ExpressionType = &StructType{Name: "Expression"}

var MacroFunctionType = &FnType{
	// (text String, getType Fn(String, Type))
	Argument: &TupleType{Elements: []TypeValue{
		&StringType{},
		&FnType{Argument: &StringType{}, Return: &TypeType{}},
	}},
	// (error String, result Expression)
	Return: &TupleType{Elements: []TypeValue{&StringType{}, ExpressionType}},
}

type TypeValue interface {
	fmt.Stringer
	typeValue()
	LowerType() cpp.Type
	LowerValue() cpp.Value
	Eq(other TypeValue) bool
}

type DefaultTypeValue struct{}

func (DefaultTypeValue) typeValue() {}

type TypeType struct{ DefaultTypeValue }

func (TypeType) String() string { return "Type" }

func (t *TypeType) Eq(other TypeValue) bool {
	_, ok := other.(*TypeType)
	return ok
}

func (TypeType) LowerType() cpp.Type   { return "ty::Type" }
func (TypeType) LowerValue() cpp.Value { return "ty::TypeType{}" }

type IntType struct{ DefaultTypeValue }

func (IntType) String() string { return "Int" }

func (i *IntType) Eq(other TypeValue) bool {
	_, ok := other.(*IntType)
	return ok
}

func (IntType) LowerType() cpp.Type   { return "int" }
func (IntType) LowerValue() cpp.Value { return "ty::IntType{}" }

type FloatType struct{ DefaultTypeValue }

func (FloatType) String() string { return "Float" }

func (f *FloatType) Eq(other TypeValue) bool {
	_, ok := other.(*FloatType)
	return ok
}

func (FloatType) LowerType() cpp.Type   { return "float" }
func (FloatType) LowerValue() cpp.Value { return "ty::FloatType{}" }

type BoolType struct{ DefaultTypeValue }

func (BoolType) String() string { return "Bool" }

func (b *BoolType) Eq(other TypeValue) bool {
	_, ok := other.(*BoolType)
	return ok
}

func (BoolType) LowerType() cpp.Type   { return "bool" }
func (BoolType) LowerValue() cpp.Value { return "ty::BoolType{}" }

type StringType struct{ DefaultTypeValue }

func (StringType) String() string { return "String" }

func (s *StringType) Eq(other TypeValue) bool {
	_, ok := other.(*StringType)
	return ok
}

func (StringType) LowerType() cpp.Type   { return "std::string" }
func (StringType) LowerValue() cpp.Value { return "ty::StringType{}" }

type TupleType struct {
	DefaultTypeValue
	Elements []TypeValue
}

func (t TupleType) String() string {
	return "(" + util.Join(t.Elements, ", ") + ")"
}

func (t *TupleType) Eq(other TypeValue) bool {
	otherTuple, ok := other.(*TupleType)
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

func (t TupleType) LowerType() cpp.Type {
	return "std::tuple<" + util.JoinFunction(t.Elements, ", ", TypeValue.LowerType) + ">"
}
func (t TupleType) LowerValue() cpp.Value {
	return "box(ty::TupleType{ .elements = { " + util.JoinFunction(t.Elements, ", ", TypeValue.LowerValue) + " } })"
}

type ListType struct {
	DefaultTypeValue
	Element TypeValue
}

func (l ListType) String() string {
	return "List(" + l.Element.String() + ")"
}

func (l *ListType) Eq(other TypeValue) bool {
	otherList, ok := other.(*ListType)
	return ok && l.Element.Eq(otherList.Element)
}
func (l ListType) LowerType() cpp.Type {
	return "std::vector<" + l.Element.LowerType() + ">"
}
func (l ListType) LowerValue() cpp.Value {
	return "box(ty::ListType{ .element = " + l.Element.LowerValue() + " })"
}

type FnType struct {
	DefaultTypeValue
	Argument TypeValue
	Return   TypeValue
}

func (f FnType) String() string {
	return "Fn(" + f.Argument.String() + ", " + f.Return.String() + ")"
}

func (f *FnType) Eq(other TypeValue) bool {
	otherFn, ok := other.(*FnType)
	return ok && f.Argument.Eq(otherFn.Argument) && f.Return.Eq(otherFn.Return)
}

func (f FnType) LowerType() cpp.Type {
	_return := f.Return.LowerType()
	argumentTuple, argumentIsTuple := f.Argument.(*TupleType)
	if !argumentIsTuple {
		return fmt.Sprintf("ty::Function<%s, %s>", _return, f.Argument.LowerType())
	}
	if len(argumentTuple.Elements) == 0 {
		return fmt.Sprintf("ty::Function<%s>", _return)
	}
	arguments := util.JoinFunction(argumentTuple.Elements, ", ", TypeValue.LowerType)
	return fmt.Sprintf("ty::Function<%s, %s>", _return, arguments)
}

func (f FnType) LowerValue() cpp.Value {
	return "box(ty::FnType{ .argument = " + f.Argument.LowerValue() + ", .returnType = " + f.Return.LowerValue() + " })"
}

type StructType struct {
	DefaultTypeValue
	Name string
}

func (s StructType) String() string {
	return s.Name
}

func (s *StructType) Eq(other TypeValue) bool {
	otherStruct, ok := other.(*StructType)
	return ok && s.Name == otherStruct.Name
}
func (s StructType) LowerType() cpp.Type {
	return "ty::" + s.Name
}
func (s StructType) LowerValue() cpp.Type {
	return "box(ty::StructType{ .name = " + s.Name + " })"
}

// Tries to unmarshal a TypeValue, returning nil if the union key does not match an Expression.
func UnmarshalTypeValue(data *fj.Value) (t TypeValue) {
	key, v := fjUnmarshalUnion(data.GetObject())
	switch key {
	case "TypeType":
		t = &TypeType{}
	case "IntType":
		t = &IntType{}
	case "FloatType":
		t = &FloatType{}
	case "BoolType":
		t = &BoolType{}
	case "StringType":
		t = &StringType{}
	case "TupleType":
		t = &TupleType{
			Elements: util.Map(v.Get("elements").GetArray(), UnmarshalTypeValue),
		}
	case "ListType":
		t = &ListType{
			Element: UnmarshalTypeValue(v.Get("element")),
		}
	case "FnType":
		t = &FnType{
			Argument: UnmarshalTypeValue(v.Get("argument")),
			Return:   UnmarshalTypeValue(v.Get("return")),
		}
	case "StructType":
		t = &StructType{
			Name: string(v.GetStringBytes("name")),
		}
	case "TypeId":
		id := string(v.GetStringBytes())
		t = registeredTypeValues[id]
	case "Box":
		return UnmarshalTypeValue(v)
	default:
		panic(fmt.Sprintf("unexpected TypeValue key when unmarshalling: %s", key))
	}
	return
}

var _ TypeValue = (*TypeType)(nil)
var _ TypeValue = (*IntType)(nil)
var _ TypeValue = (*FloatType)(nil)
var _ TypeValue = (*BoolType)(nil)
var _ TypeValue = (*StringType)(nil)
var _ TypeValue = (*TupleType)(nil)
var _ TypeValue = (*ListType)(nil)
var _ TypeValue = (*FnType)(nil)
var _ TypeValue = (*StructType)(nil)
