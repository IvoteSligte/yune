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
		prev, exists := paramNames[param.GetName()]
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
	deps.declarations = map[string]Declaration{d.GetName(): d}
	for i := range d.Parameters {
		param := &d.Parameters[i]
		deps.declarations[param.GetName()] = param
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
	if d.GetName() == "main" && !d.GetType().Eq(MainType) {
		errors = append(errors, InvalidMainSignature{
			Found: d.GetType(),
			At:    d.Name.GetSpan(),
		})
	}
	return
}

// GetValueDependencies implements Declaration.
func (d FunctionDeclaration) GetValueDependencies() (deps []string) {
	for _, dep := range d.Body.GetValueDependencies() {
		equals := func(param FunctionParameter) bool {
			return dep == param.GetName()
		}
		if dep != d.GetName() && !util.Any(equals, d.Parameters...) {
			deps = append(deps, dep)
		}
	}
	return
}

// GetTypeDependencies implements Declaration.
func (d *FunctionDeclaration) GetTypeDependencies() (deps []*Type) {
	deps = util.FlatMapPtr(d.Parameters, (*FunctionParameter).GetTypeDependencies)
	deps = append(deps, &d.ReturnType)
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

func (d FunctionDeclaration) GetName() string {
	return d.Name.String
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
func (d FunctionParameter) GetName() string {
	return d.Name.String
}

// GetType implements Declaration
func (d FunctionParameter) GetType() value.Type {
	return d.Type.Get()
}

// GetTypeDependencies implements Declaration
func (d *FunctionParameter) GetTypeDependencies() []*Type {
	return []*Type{&d.Type}
}

// GetValueDependencies implements Declaration
func (d FunctionParameter) GetValueDependencies() (deps []string) {
	return
}

type ConstantDeclaration struct {
	Span Span
	Name Name
	Type Type
	Body Block
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

// GetTypeDependencies implements Declaration.
func (d *ConstantDeclaration) GetTypeDependencies() []*Type {
	return append([]*Type{&d.Type}, d.Body.GetTypeDependencies()...)
}

// GetValueDependencies implements Declaration.
func (d ConstantDeclaration) GetValueDependencies() []string {
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

func (d ConstantDeclaration) GetName() string {
	return d.Name.String
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
