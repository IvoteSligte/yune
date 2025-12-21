package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"
)

type Type struct {
	Span
	// unique can reliably be compared with for equality, whereas Type
	// may contain aliases that would report false inequalities.
	unique   cpp.Type
	Name     Name
	Generics []Type
}

func (t Type) GetValueDependencies() []string {
	return append(util.FlatMap(t.Generics, Type.GetValueDependencies), t.Name.GetName())
}

func (t *Type) Get() cpp.Type {
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
			Found: cpp.Type{},
			At:    Span{},
		})
	}
	// TODO: use Generics to compute the final type
	// and use the computed value when aliases exist
	t.unique = cpp.Type{
		Name: decl.Lower().(cpp.TypeAlias).Alias,
	}
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
