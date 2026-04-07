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

func literalType(name string, _type TypeValue) TypeValue {
	return &StructType{
		Name: name,
		Fields: []StructTypeField{
			{Name: "value", Type: _type},
		},
	}
}

// kinds of Expression types
var IntegerLiteralType = literalType("IntegerLiteral", &IntType{})
var FloatLiteralType = literalType("FloatLiteral", &FloatType{})
var BoolLiteralType = literalType("BoolLiteral", &BoolType{})
var StringLiteralType = literalType("StringLiteral", &StringType{})

// technically not even a StructType, but this works
var ExpressionType = &StructType{Name: "Expression", Fields: []StructTypeField{}}

var MacroFunctionType = &FnType{
	// (text String, getType Fn(String, Type))
	Argument: &TupleType{Elements: []TypeValue{
		&StringType{},
		&FnType{Argument: &StringType{}, Return: &TypeType{}},
	}},
	// (error String, result Expression)
	Return: &UnionType{Variants: []TypeValue{&StringType{}, ExpressionType}},
}

type TypeValue interface {
	fmt.Stringer
	typeValue()
	LowerType() cpp.Type
	LowerValue() cpp.Value
	Eq(other TypeValue) bool
}

func IsSubType(sub TypeValue, super TypeValue) bool {
	superUnion, superIsUnion := super.(*UnionType)
	if !superIsUnion {
		return sub.Eq(super) || sub.Eq(&UnionType{})
	}
	subUnion, subIsUnion := sub.(*UnionType)
	if !subIsUnion {
		return superUnion.HasVariant(sub)
	}
	return subUnion.IsSubUnion(superUnion)
}

type DefaultTypeValue struct{}

func (DefaultTypeValue) typeValue() {}

type TypeType struct{ DefaultTypeValue }

func (TypeType) String() string { return "Type" }

func (t *TypeType) Eq(other TypeValue) bool {
	_, ok := other.(*TypeType)
	return ok
}

func (TypeType) LowerType() cpp.Type   { return "Type_t" }
func (TypeType) LowerValue() cpp.Value { return "TypeType_t{}" }

type IntType struct{ DefaultTypeValue }

func (IntType) String() string { return "Int" }

func (i *IntType) Eq(other TypeValue) bool {
	_, ok := other.(*IntType)
	return ok
}

func (IntType) LowerType() cpp.Type   { return "int" }
func (IntType) LowerValue() cpp.Value { return "IntType_t{}" }

type FloatType struct{ DefaultTypeValue }

func (FloatType) String() string { return "Float" }

func (f *FloatType) Eq(other TypeValue) bool {
	_, ok := other.(*FloatType)
	return ok
}

func (FloatType) LowerType() cpp.Type   { return "float" }
func (FloatType) LowerValue() cpp.Value { return "FloatType_t{}" }

type BoolType struct{ DefaultTypeValue }

func (BoolType) String() string { return "Bool" }

func (b *BoolType) Eq(other TypeValue) bool {
	_, ok := other.(*BoolType)
	return ok
}

func (BoolType) LowerType() cpp.Type   { return "bool" }
func (BoolType) LowerValue() cpp.Value { return "BoolType_t{}" }

type StringType struct{ DefaultTypeValue }

func (StringType) String() string { return "String" }

func (s *StringType) Eq(other TypeValue) bool {
	_, ok := other.(*StringType)
	return ok
}

func (StringType) LowerType() cpp.Type   { return "String_t" }
func (StringType) LowerValue() cpp.Value { return "StringType_t{}" }

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
	return "std::tuple<" + util.JoinFunc(t.Elements, ", ", TypeValue.LowerType) + ">"
}
func (t TupleType) LowerValue() cpp.Value {
	return "box(TupleType_t{ .elements = { " + util.JoinFunc(t.Elements, ", ", TypeValue.LowerValue) + " } })"
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
	return "List_t<" + l.Element.LowerType() + ">"
}
func (l ListType) LowerValue() cpp.Value {
	return "box(ListType_t{ .element = " + l.Element.LowerValue() + " })"
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
		return fmt.Sprintf("Fn_t<%s, %s>", _return, f.Argument.LowerType())
	}
	if len(argumentTuple.Elements) == 0 {
		return fmt.Sprintf("Fn_t<%s>", _return)
	}
	arguments := util.JoinFunc(argumentTuple.Elements, ", ", TypeValue.LowerType)
	return fmt.Sprintf("Fn_t<%s, %s>", _return, arguments)
}

func (f FnType) LowerValue() cpp.Value {
	return "box(FnType_t{ .argument = " + f.Argument.LowerValue() + ", .returnType = " + f.Return.LowerValue() + " })"
}

type StructTypeField struct {
	Name string
	Type TypeValue
}

func (s StructTypeField) String() string {
	return fmt.Sprintf(`%s: %s`, s.Name, s.Type)
}

func (s StructTypeField) LowerValue() cpp.Type {
	return fmt.Sprintf(`{%s, %s}`, s.Name, s.Type.LowerValue())
}

type StructType struct {
	DefaultTypeValue
	Name   string
	Fields []StructTypeField
}

func (s StructType) String() string {
	return fmt.Sprintf(`%s { %s }`, s.Name, util.Join(s.Fields, ", "))
}

func (s *StructType) Eq(other TypeValue) bool {
	// NOTE: this does not compare fields as it assumes name uniqueness
	otherStruct, ok := other.(*StructType)
	return ok && s.Name == otherStruct.Name
}
func (s StructType) LowerType() cpp.Type {
	return s.Name + "_t"
}
func (s StructType) LowerValue() cpp.Type {
	return fmt.Sprintf(
		`box(StructType_t{ .name = %s, .fields = { %s }  })`,
		s.Name, util.JoinFunc(s.Fields, ", ", StructTypeField.LowerValue),
	)
}

type UnionType struct {
	DefaultTypeValue
	Variants []TypeValue
}

func (u UnionType) String() string {
	return fmt.Sprintf("Union(%s)", util.Join(u.Variants, ", "))
}

func (u *UnionType) Eq(other TypeValue) bool {
	otherUnion, ok := other.(*UnionType)
	if !ok || len(u.Variants) != len(otherUnion.Variants) {
		return false
	}
	// unions are unordered
	for _, element := range u.Variants {
		return util.Any(otherUnion.Variants, func(otherElement TypeValue) bool {
			return element.Eq(otherElement)
		})
	}
	return true
}
func (u UnionType) LowerType() cpp.Type {
	return "Union_t<" + util.JoinFunc(u.Variants, ", ", TypeValue.LowerType) + ">"
}
func (u UnionType) LowerValue() cpp.Type {
	return "box(TupleType_t{ .variants = { " + util.JoinFunc(u.Variants, ", ", TypeValue.LowerValue) + " } })"
}

func (u UnionType) HasVariant(variant TypeValue) bool {
	for _, v := range u.Variants {
		if v.Eq(variant) {
			return true
		}
	}
	return false
}

func (u UnionType) IsSubUnion(super *UnionType) bool {
	for _, v := range u.Variants {
		if !super.HasVariant(v) {
			return false
		}
	}
	return true
}

// Creates a union with the same logic as in pb.hpp
func NewUnionType(variants ...TypeValue) TypeValue {
	// Flattens variants non-recursively.
	flatVariants := []TypeValue{}
	for _, variant := range variants {
		variantUnion, variantIsUnion := variant.(*UnionType)
		if variantIsUnion {
			flatVariants = append(flatVariants, variantUnion.Variants...)
		} else {
			flatVariants = append(flatVariants, variant)
		}
	}
	// Remove duplicate variants.
	uniqueVariants := []TypeValue{}
	for _, variant := range flatVariants {
		if !util.Any(uniqueVariants, func(other TypeValue) bool {
			return other.Eq(variant)
		}) {
			uniqueVariants = append(uniqueVariants, variant)
		}
	}
	if len(uniqueVariants) == 1 {
		return uniqueVariants[0]
	}
	return &UnionType{Variants: uniqueVariants}
}

// Tries to unmarshal a TypeValue, returning nil if the union key does not match an Expression.
func (state *State) UnmarshalTypeValue(data *fj.Value) (t TypeValue) {
	key, v, _ := fjUnmarshalStruct(data.GetObject())
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
			Elements: util.Map(UnmarshalList(v, "elements"), state.UnmarshalTypeValue),
		}
	case "ListType":
		t = &ListType{
			Element: state.UnmarshalTypeValue(v.Get("element")),
		}
	case "FnType":
		t = &FnType{
			Argument: state.UnmarshalTypeValue(v.Get("argument")),
			Return:   state.UnmarshalTypeValue(v.Get("return")),
		}
	case "StructType":
		t = &StructType{
			Name: UnmarshalNonEmptyString(v, "name"),
			Fields: util.Map(UnmarshalList(v, "fields"), func(v *fj.Value) StructTypeField {
				return StructTypeField{
					Name: UnmarshalNonEmptyString(v, "name"),
					Type: state.UnmarshalTypeValue(v.Get("type")),
				}
			}),
		}
	case "UnionType":
		t = &UnionType{
			Variants: util.Map(v.Get("variants").GetArray(), state.UnmarshalTypeValue),
		}
	case "TypeId":
		id := UnmarshalNonEmptyString(v)
		t = state.registeredTypeValues[id]
	case "Box":
		return state.UnmarshalTypeValue(v)
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
var _ TypeValue = (*UnionType)(nil)
