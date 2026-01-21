package ast

import (
	"fmt"
	"log"
	"yune/cpp"
	"yune/util"
)

type Type interface {
	Node
	fmt.Stringer
	Get() cpp.Type
	GetValueDependencies() []string
	Calc(deps DeclarationTable) (errors Errors)
	Lower() cpp.Type
}

type NamedType struct {
	Span
	// unique can reliably be compared with for equality, whereas Type
	// may contain aliases that would report false inequalities.
	unique cpp.Type
	Name   Name
}

func (t NamedType) GetValueDependencies() []string {
	return []string{t.Name.GetName()}
}

func (t *NamedType) Get() cpp.Type {
	if t.unique == nil {
		log.Printf("Get() called on unresolved type '%s'.", t.Name)
	}
	return t.unique
}

// Calculates the true type without aliases that this type represents.
func (t *NamedType) Calc(deps DeclarationTable) (errors Errors) {
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

func (t NamedType) Lower() cpp.Type {
	return cpp.NamedType{Name: t.GetName()}
}

func (t NamedType) GetName() string {
	return t.Name.GetName()
}

func (t NamedType) String() string {
	return t.GetName()
}

type FunctionType struct {
	Span
	Arguments  []Type
	ReturnType Type
}

func (t FunctionType) GetValueDependencies() []string {
	return append(util.FlatMap(t.Arguments, Type.GetValueDependencies), t.ReturnType.GetValueDependencies()...)
}

func (t *FunctionType) Get() cpp.Type {
	return cpp.FunctionType{
		Parameter: cpp.TupleType{
			Elements: util.Map(t.Arguments, Type.Get),
		},
		ReturnType: t.ReturnType.Get(),
	}
}

// Calculates the true type without aliases that this type represents.
func (t *FunctionType) Calc(deps DeclarationTable) (errors Errors) {
	for i := range t.Arguments {
		errors = append(errors, t.Arguments[i].Calc(deps)...)
	}
	errors = append(errors, t.ReturnType.Calc(deps)...)
	return
}

func (t FunctionType) Lower() cpp.Type {
	return t.Get()
}

func (t FunctionType) String() string {
	return fmt.Sprintf("fn(%s): %s", util.Join(t.Arguments, ", "), t.ReturnType)
}

var _ Type = &NamedType{}
var _ Type = &FunctionType{}
