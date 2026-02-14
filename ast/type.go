package ast

import (
	"fmt"
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

func UnmarshalType(data *fj.Value) Type {
	return Type{
		Expression: UnmarshalExpression(data.Get("Type")),
	}
}

func (t *Type) Analyze(anal Analyzer) TypeValue {
	println("start analyzing type")
	// FIXME: t.Expression.Analyze can access local variables right now
	expressionType := t.Expression.Analyze(TypeType{}, anal)
	// TODO: check if expressionType is part of the union TypeType rather than equal
	// (is this necessary?)
	if expressionType != nil && !expressionType.Eq(TypeType{}) {
		anal.PushError(UnexpectedType{
			Expected: TypeType{},
			Found:    t.value,
			At:       t.Expression.GetSpan(),
		})
	}
	json := anal.Evaluate(t.Expression)
	t.value = UnmarshalTypeValue(fj.MustParse(json))
	println("finish analyzing type")
	return t.value
}

var MainType = FnType{
	Argument: TupleType{},
	Return:   TupleType{},
}

// kinds of Expression types
var IntegerLiteralType = StructType{Name: "IntegerLiteral"}
var FloatLiteralType = StructType{Name: "FloatLiteral"}
var BoolLiteralType = StructType{Name: "BoolLiteral"}
var StringLiteralType = StructType{Name: "StringLiteral"}
var ExpressionType = StructType{Name: "Expression"}

var MacroFunctionType = FnType{
	Argument: StringType{},
	Return:   NewTupleType(StringType{}, ExpressionType),
}

// Tries to unmarshal a TypeValue, returning nil if the union key does not match an Expression.
func UnmarshalTypeValue(data *fj.Value) (t TypeValue) {
	key, v := fjUnmarshalUnion(data.GetObject())
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
	case "TupleType":
		t = TupleType{
			Elements: util.Map(v.Get("elements").GetArray(), UnmarshalTypeValue),
		}
	case "ListType":
		t = ListType{
			Element: UnmarshalTypeValue(v.Get("element")),
		}
	case "FnType":
		t = FnType{
			Argument: UnmarshalTypeValue(v.Get("argument")),
			Return:   UnmarshalTypeValue(v.Get("return")),
		}
	case "StructType":
		t = StructType{
			Name: string(v.GetStringBytes("name")),
		}
	default:
		// t = nil
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

type TypeValue interface {
	fmt.Stringer
	typeValue()
	Lower() cpp.Type
	Eq(other TypeValue) bool
}

type DefaultTypeValue struct{}

func (DefaultTypeValue) typeValue() {}
func (DefaultTypeValue) String() string {
	panic("DefaultTypeValue.String should be overridden")
}
func (DefaultTypeValue) Lower() cpp.Type {
	panic("DefaultTypeValue.Lower should be overridden")
}
func (DefaultTypeValue) Eq(other TypeValue) bool {
	panic("DefaultTypeValue.Eq should be overridden")
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
	return "std::tuple<" + util.JoinFunction(t.Elements, ", ", TypeValue.Lower) + ">"
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
	return "std::vector<" + l.Element.Lower() + ">"
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
	argumentTuple := wrapTupleType(f.Argument)
	if len(argumentTuple.Elements) == 0 {
		return fmt.Sprintf("ty::Function<%s>", _return)
	} else {
		arguments := util.JoinFunction(argumentTuple.Elements, ", ", TypeValue.Lower)
		return fmt.Sprintf("ty::Function<%s, %s>", _return, arguments)
	}
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
	return "ty::" + s.Name
}

var _ TypeValue = TypeType{}
var _ TypeValue = IntType{}
var _ TypeValue = FloatType{}
var _ TypeValue = BoolType{}
var _ TypeValue = StringType{}
var _ TypeValue = TupleType{}
var _ TypeValue = ListType{}
var _ TypeValue = FnType{}
var _ TypeValue = StructType{}
