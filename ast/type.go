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

type TupleType struct {
	Span
	Elements []Type
}

func (t TupleType) GetValueDependencies() []string {
	return util.FlatMap(t.Elements, Type.GetValueDependencies)
}

func (t *TupleType) Get() cpp.Type {
	return cpp.TupleType{
		Elements: util.Map(t.Elements, Type.Get),
	}
}

// Calculates the true type without aliases that this type represents.
func (t *TupleType) Calc(deps DeclarationTable) (errors Errors) {
	for i := range t.Elements {
		errors = append(errors, t.Elements[i].Calc(deps)...)
	}
	return
}

func (t TupleType) Lower() cpp.Type {
	return t.Get()
}

func (t TupleType) String() string {
	if len(t.Elements) == 1 {
		return fmt.Sprintf("(%s,)", t.Elements[0])
	} else {
		return fmt.Sprintf("(%s)", util.Join(t.Elements, ", "))
	}
}

type FunctionType struct {
	Span
	Argument   Type
	ReturnType Type
}

func (t FunctionType) GetValueDependencies() []string {
	return append(t.Argument.GetValueDependencies(), t.ReturnType.GetValueDependencies()...)
}

func (t *FunctionType) Get() cpp.Type {
	return cpp.FunctionType{
		Parameter:  t.Argument.Get(),
		ReturnType: t.ReturnType.Get(),
	}
}

// Calculates the true type without aliases that this type represents.
func (t *FunctionType) Calc(deps DeclarationTable) (errors Errors) {
	errors = append(t.Argument.Calc(deps), t.ReturnType.Calc(deps)...)
	return
}

func (t FunctionType) Lower() cpp.Type {
	return t.Get()
}

func (t FunctionType) String() string {
	_, isTupleType := t.Argument.(*TupleType)
	if isTupleType {
		return fmt.Sprintf("fn%s: %s", t.Argument, t.ReturnType)
	} else {
		return fmt.Sprintf("fn(%s): %s", t.Argument, t.ReturnType)
	}
}

var _ Type = &NamedType{}
var _ Type = &TupleType{}
var _ Type = &FunctionType{}
