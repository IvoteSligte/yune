package ast

import (
	"fmt"
	"slices"
	"yune/cpp"
	"yune/util"
)

type Type struct {
	Span
	// unique can reliably be compared with for equality, whereas Type
	// may contain aliases that would report false inequalities.
	unique   InferredType
	Name     Name
	Generics []Type
}

func (t Type) GetValueDependencies() []string {
	return append(util.FlatMap(t.Generics, Type.GetValueDependencies), t.Name.GetName())
}

func (t *Type) Get() InferredType {
	return t.unique
}

// Calculates the true type without aliases that this type represents.
func (t *Type) Calc(deps DeclarationTable) (errors Errors) {
	for i := range t.Generics {
		errors = append(errors, t.Generics[i].Calc(deps)...)
	}
	if len(errors) > 0 {
		return
	}
	decl, ok := deps.GetTopLevel(t.Name.String)
	if !ok {
		errors = append(errors, UndefinedType{
			String: t.Name.String,
			Span:   t.Name.GetSpan(),
		})
		return
	}
	if !decl.GetType().Eq(TypeType) {
		errors = append(errors, NotAType{
			Found: InferredType{},
			At:    Span{},
		})
	}
	// TODO: use Generics to compute the final type
	t.unique = decl.GetValue().(InferredType)
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

// value implements Value.
func (t InferredType) value() {
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
