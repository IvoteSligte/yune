package ast

import (
	"log"
	"yune/cpp"
)

type Type struct {
	Span
	// unique can reliably be compared with for equality, whereas Type
	// may contain aliases that would report false inequalities.
	unique cpp.Type
	Name   Name
}

func (t Type) GetValueDependencies() []string {
	return []string{t.Name.GetName()}
}

func (t *Type) Get() cpp.Type {
	if t.unique == nil {
		log.Printf("Get() called on unresolved type '%s'.", t.Name)
	}
	return t.unique
}

// Calculates the true type without aliases that this type represents.
func (t *Type) Calc(deps DeclarationTable) (errors Errors) {
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
			Found: decl.GetType(),
			At:    decl.GetSpan(),
		})
		return
	}
	// TODO: use the computed value once (non-builtin) aliases exist
	t.unique = decl.Lower().(cpp.TypeAlias).Get()
	return
}

func (t Type) Lower() cpp.Type {
	return cpp.NamedType{Name: t.GetName()}
}

func (t Type) GetName() string {
	return t.Name.GetName()
}

func (t Type) String() string {
	return t.GetName()
}

var _ Node = &Type{}
