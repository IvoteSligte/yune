package ast

import (
	"fmt"
	"slices"
	"yune/cpp"
	"yune/util"
)

type Type struct {
	Span
	// InferredType can reliably be compared with for equality, whereas Type
	// may contain aliases that would report false inequalities.
	InferredType
	Name     Variable
	Generics []Type
}

// GetValueDependencies implements Node.
func (t Type) GetValueDependencies() []string {
	return append(util.FlatMap(t.Generics, Type.GetValueDependencies), t.Name.GetName())
}

// InferType implements Node.
func (t *Type) InferType(deps DeclarationTable) (errors Errors) {
	for i := range t.Generics {
		errors = append(errors, t.Generics[i].InferType(deps)...)
	}
	errors = append(errors, t.Name.InferType(deps)...)
	if len(errors) > 0 {
		return
	}
	// TODO: use Generics
	t.InferredType = t.Name.GetType()
	return
}

func (t Type) Lower() cpp.Type {
	return cpp.Type{
		Name:     t.GetName(),
		Generics: util.Map(t.Generics, Type.Lower),
	}
}

func (t Type) GetName() string {
	return t.Name.GetName()
}

func (t Type) String() string {
	if len(t.Generics) == 0 {
		return t.GetName()
	} else {
		return fmt.Sprintf("%s<%s>", t.Name, util.SeparatedBy(t.Generics, ", "))
	}
}

var _ Node = &Type{}

type InferredType struct {
	name     string
	generics []InferredType
}

func (t InferredType) GetType() InferredType {
	return t
}

// Function type has generics[0] as argument type and generics[1] as return type.
func (t InferredType) IsFunction() bool {
	return t.name == "Fn"
}

func (t InferredType) GetParameterType() (parameterType InferredType, isFunction bool) {
	isFunction = t.IsFunction()
	if !isFunction {
		return
	}
	parameterType = t.generics[0]
	return
}

func (t InferredType) GetReturnType() (returnType InferredType, isFunction bool) {
	isFunction = t.IsFunction()
	if !isFunction {
		return
	}
	returnType = t.generics[1]
	return
}

func (t InferredType) IsTuple() bool {
	return t.name == "Tuple"
}

func (t InferredType) GetGeneric(i int) InferredType {
	return t.generics[i]
}

func (t InferredType) String() string {
	if len(t.generics) == 0 {
		return t.name
	} else {
		return fmt.Sprintf("%s<%s>", t.name, util.SeparatedBy(t.generics, ", "))
	}
}

func (left InferredType) Eq(right InferredType) bool {
	return left.name == right.name &&
		slices.EqualFunc(left.generics, right.generics, InferredType.Eq)
}

var IntType InferredType = InferredType{name: "Int"}
var FloatType InferredType = InferredType{name: "Float"}
var BoolType InferredType = InferredType{name: "Bool"}
var NilType InferredType = InferredType{name: "Nil"}
