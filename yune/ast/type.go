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
	Name     Name
	Generics []Type
}

func (t Type) Lower() cpp.Type {
	return cpp.Type{
		Name:     t.Name.String,
		Generics: util.Map(t.Generics, Type.Lower),
	}
}

func (t *Type) Analyze() (queries Queries, _finalizer Finalizer) {
	finalizers := make([]Finalizer, len(t.Generics))
	for i := range t.Generics {
		_queries, _finalizer := t.Generics[i].Analyze()
		queries = append(queries, _queries...)
		finalizers = append(finalizers, _finalizer)
	}
	queries = append(queries, t.Name)
	_finalizer = func(env Env) (errors Errors) {
		for _, fin := range finalizers {
			errors = append(errors, fin(env)...)
		}
		if len(errors) > 0 {
			return
		}
		if len(t.Generics) > 0 {
			panic("unimplemented: generics")
		} else {
			t.InferredType = env.Get(t.Name.String).GetType()
		}
		return
	}
	return
}

func (t Type) String() string {
	if len(t.Generics) == 0 {
		return t.Name.String
	} else {
		return fmt.Sprintf("%s<%s>", t.Name, util.SeparatedBy(t.Generics, ", "))
	}
}

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
