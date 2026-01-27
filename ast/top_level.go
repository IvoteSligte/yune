package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"
	"yune/value"
)

type FunctionDeclaration struct {
	Span       Span
	Name       Name
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
	Value      cpp.FunctionDeclaration
}

// GetMacros implements Declaration.
func (d *FunctionDeclaration) GetMacros() (macros []*Macro) {
	macros = util.FlatMap(d.Parameters, FunctionParameter.GetMacros)
	macros = append(macros, d.Body.GetMacros()...)
	return
}

// GetSpan implements TopLevelDeclaration.
func (d *FunctionDeclaration) GetSpan() Span {
	return d.Name.GetSpan()
}

// TypeCheckBody implements Declaration.
func (d *FunctionDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	// check for duplicate parameters
	paramNames := map[string]*FunctionParameter{}
	for i := range d.Parameters {
		param := &d.Parameters[i]
		prev, exists := paramNames[param.GetName().String]
		if exists {
			errors = append(errors, DuplicateDeclaration{
				First:  prev,
				Second: param,
			})
		}
	}
	if len(errors) > 0 {
		return
	}
	deps = deps.NewScope()
	deps.declarations = map[string]Declaration{d.GetName().String: d}
	for i := range d.Parameters {
		param := &d.Parameters[i]
		deps.declarations[param.GetName().String] = param
	}
	errors = append(errors, d.Body.InferType(deps)...)
	if len(errors) > 0 {
		return
	}
	returnType := d.ReturnType.Get()
	bodyType := d.Body.GetType()

	if !returnType.Eq(bodyType) {
		errors = append(errors, ReturnTypeMismatch{
			Expected: returnType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
		})
	}
	if d.GetName().String == "main" && !d.GetType().Eq(MainType) {
		errors = append(errors, InvalidMainSignature{
			Found: d.GetType(),
			At:    d.Name.GetSpan(),
		})
	}
	return
}

// GetMacroValueDependencies implements Declaration.
func (d FunctionDeclaration) GetMacroValueDependencies() (deps []Name) {
	for _, depName := range d.Body.GetMacroValueDependencies() {
		equals := func(param FunctionParameter) bool {
			return depName.String == param.GetName().String
		}
		if depName.String != d.GetName().String && !util.Any(equals, d.Parameters...) {
			deps = append(deps, depName)
		}
	}
	return
}

// GetMacroTypeDependencies implements Declaration.
func (d *FunctionDeclaration) GetMacroTypeDependencies() (deps []Query) {
	deps = util.FlatMapPtr(d.Parameters, (*FunctionParameter).GetMacroTypeDependencies)
	deps = append(deps, d.Body.GetMacroTypeDependencies()...)
	return
}

// GetValueDependencies implements Declaration.
func (d FunctionDeclaration) GetValueDependencies() (deps []Name) {
	for _, depName := range d.Body.GetValueDependencies() {
		equals := func(param FunctionParameter) bool {
			return depName.String == param.GetName().String
		}
		if depName.String != d.GetName().String && !util.Any(equals, d.Parameters...) {
			deps = append(deps, depName)
		}
	}
	return
}

// GetTypeDependencies implements Declaration.
func (d *FunctionDeclaration) GetTypeDependencies() (deps []Query) {
	deps = util.FlatMapPtr(d.Parameters, (*FunctionParameter).GetTypeDependencies)
	deps = append(deps, Query{
		Expression:   d.ReturnType.Expression,
		Destination:  &d.ReturnType.value,
		ExpectedType: TypeType,
	})
	deps = append(deps, d.Body.GetTypeDependencies()...)
	return
}

// Lower implements Declaration.
func (d FunctionDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.FunctionDeclaration{
		Name:       d.Name.String,
		Parameters: util.Map(d.Parameters, FunctionParameter.Lower),
		ReturnType: d.ReturnType.Lower(),
		Body:       d.Body.LowerFunctionBody(),
	}
}

func (d FunctionDeclaration) GetName() Name {
	return d.Name
}

func (d FunctionDeclaration) GetType() value.Type {
	params := util.Join(util.Map(d.Parameters, FunctionParameter.GetType), ", ")
	return value.Type(fmt.Sprintf("std::function<%s(%s)>", d.ReturnType.Get(), params))
}

type FunctionParameter struct {
	Span
	Name Name
	Type Type
}

// GetMacros implements Declaration.
func (d FunctionParameter) GetMacros() []*Macro {
	return []*Macro{}
}

// TypeCheckBody implements Declaration.
func (d FunctionParameter) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	return
}

func (d FunctionParameter) Lower() cpp.FunctionParameter {
	return cpp.FunctionParameter{
		Name: d.Name.String,
		Type: d.Type.Lower(),
	}
}

// GetName implements Declaration
func (d FunctionParameter) GetName() Name {
	return d.Name
}

// GetType implements Declaration
func (d FunctionParameter) GetType() value.Type {
	return d.Type.Get()
}

// GetMacroTypeDependencies implements Declaration
func (d *FunctionParameter) GetMacroTypeDependencies() (deps []Query) {
	return
}

// GetMacroValueDependencies implements Declaration
func (d FunctionParameter) GetMacroValueDependencies() (deps []Name) {
	return
}

// GetTypeDependencies implements Declaration
func (d *FunctionParameter) GetTypeDependencies() (deps []Query) {
	deps = append(deps, Query{
		Expression:   d.Type.Expression,
		Destination:  &d.Type.value,
		ExpectedType: TypeType,
	})
	return
}

// GetValueDependencies implements Declaration
func (d FunctionParameter) GetValueDependencies() (deps []Name) {
	return
}

type ConstantDeclaration struct {
	Span Span
	Name Name
	Type Type
	Body Block
}

// GetMacros implements Declaration.
func (d *ConstantDeclaration) GetMacros() []*Macro {
	return d.Body.GetMacros()
}

// GetSpan implements TopLevelDeclaration.
func (d *ConstantDeclaration) GetSpan() Span {
	return d.Name.GetSpan()
}

// TypeCheckBody implements TopLevelDeclaration.
func (d *ConstantDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	errors = d.Body.InferType(deps)
	if len(errors) > 0 {
		return
	}
	declType := d.Type.Get()
	bodyType := d.Body.GetType()

	if !declType.Eq(bodyType) {
		errors = append(errors, ConstantTypeMismatch{
			Expected: declType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
		})
	}
	return
}

// GetMacroTypeDependencies implements Declaration.
func (d *ConstantDeclaration) GetMacroTypeDependencies() (deps []Query) {
	deps = append(deps, d.Body.GetMacroTypeDependencies()...)
	return
}

// GetMacroValueDependencies implements Declaration.
func (d ConstantDeclaration) GetMacroValueDependencies() []Name {
	return d.Body.GetMacroValueDependencies()
}

// GetTypeDependencies implements Declaration.
func (d *ConstantDeclaration) GetTypeDependencies() (deps []Query) {
	deps = append(deps, Query{
		Expression:   d.Type.Expression,
		Destination:  &d.Type.value,
		ExpectedType: TypeType,
	})
	deps = append(deps, d.Body.GetTypeDependencies()...)
	return
}

// GetValueDependencies implements Declaration.
func (d ConstantDeclaration) GetValueDependencies() []Name {
	return d.Body.GetValueDependencies()
}

// Lower implements Declaration.
func (d ConstantDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.ConstantDeclaration{
		Name:  d.Name.String,
		Type:  d.Type.Lower(),
		Value: d.Body.LowerVariableBody(),
	}
}

// GetType implements Declaration.
func (d ConstantDeclaration) GetType() value.Type {
	return d.Type.Get()
}

func (d ConstantDeclaration) GetName() Name {
	return d.Name
}

func (d ConstantDeclaration) GetDeclarationType() Type {
	return d.Type
}

type TopLevelDeclaration interface {
	Declaration
	// Lowers the declaration to executable C++ code.
	// Assumes type checking has been performed.
	//
	// NOTE: when the value has been computed, this function should
	// lower to a more efficient representation instead of forcing
	// the same code to run.
	Lower() cpp.TopLevelDeclaration
}

func isConstantDeclaration(decl TopLevelDeclaration) bool {
	_, isConstant := decl.(*ConstantDeclaration)
	return isConstant
}

// TODO: when types and type aliases can be created, make sure that
// values are cached and aliases are properly resolved.

var _ TopLevelDeclaration = &FunctionDeclaration{}
var _ TopLevelDeclaration = &ConstantDeclaration{}
