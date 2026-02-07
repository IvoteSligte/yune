package ast

import (
	"yune/cpp"
	"yune/util"
)

// Type checks a possibly unnamed function. `declaration == nil` for unnamed functions.
func typeCheckFunction(declaration *FunctionDeclaration, parameters []FunctionParameter, returnType Type, body Block, deps DeclarationTable) (errors Errors) {
	// check for duplicate parameters
	paramNames := map[string]*FunctionParameter{}
	for i := range parameters {
		param := &parameters[i]
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
	deps.declarations = map[string]Declaration{}
	if declaration != nil { // allow recursion by registering the function
		deps.declarations[declaration.GetName().String] = declaration
	}
	for i := range parameters {
		param := &parameters[i]
		deps.declarations[param.GetName().String] = param
	}
	errors = append(errors, body.InferType(deps.NewScope())...)
	if len(errors) > 0 {
		return
	}
	_returnType := returnType.Get()
	bodyType := body.GetType()

	if !typeEqual(_returnType, bodyType) {
		errors = append(errors, ReturnTypeMismatch{
			Expected: _returnType,
			Found:    bodyType,
			At:       body.Statements[len(body.Statements)-1].GetSpan(),
		})
	}
	return
}

func getFunctionType(parameters []FunctionParameter, returnType Type) FnType {
	params := util.Map(parameters, func(p FunctionParameter) TypeValue {
		return p.GetDeclaredType()
	})
	var argument TypeValue
	if len(parameters) == 1 {
		argument = params[0]
	} else {
		argument = NewTupleType(params...)
	}
	return FnType{Argument: argument, Return: returnType.Get()}
}

func getFunctionTypeDependencies(parameters []FunctionParameter, returnType *Type, body Block) (deps []Query) {
	deps = util.FlatMapPtr(parameters, (*FunctionParameter).GetTypeDependencies)
	returnType.Expression.SetType(TypeType{})
	deps = append(deps, Query{
		Expression:  returnType.Expression,
		Destination: SetType{Type: &returnType.value},
	})
	deps = append(deps, body.GetTypeDependencies()...)
	return
}

type FunctionDeclaration struct {
	Span       Span
	Name       Name
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
}

// GetMacros implements Declaration.
func (d *FunctionDeclaration) GetMacros() (macros []*Macro) {
	// NOTE: currently assuming that types do not have macros that also have type dependencies
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
	errors = typeCheckFunction(d, d.Parameters, d.ReturnType, d.Body, deps)
	if len(errors) > 0 {
		return
	}
	if d.GetName().String == "main" && !typeEqual(d.GetDeclaredType(), MainType) {
		errors = append(errors, InvalidMainSignature{
			Found: d.GetDeclaredType(),
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
	return getFunctionTypeDependencies(d.Parameters, &d.ReturnType, d.Body)
}

// Lower implements Declaration.
func (d FunctionDeclaration) Lower() cpp.Declaration {
	return cpp.FunctionDeclaration(
		d.Name.String,
		util.Map(d.Parameters, FunctionParameter.Lower),
		d.ReturnType.Lower(),
		cpp.Block(d.Body.lowerStatements()),
	)
}

func (d FunctionDeclaration) GetName() Name {
	return d.Name
}

func (d FunctionDeclaration) GetDeclaredType() TypeValue {
	return getFunctionType(d.Parameters, d.ReturnType)
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
	return d.Type.Lower() + " " + d.Name.String
}

// GetName implements Declaration
func (d FunctionParameter) GetName() Name {
	return d.Name
}

// GetDeclaredType implements Declaration
func (d FunctionParameter) GetDeclaredType() TypeValue {
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
	d.Type.Expression.SetType(TypeType{})
	deps = append(deps, Query{
		Expression:  d.Type.Expression,
		Destination: SetType{&d.Type.value},
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
	errors = d.Body.InferType(deps.NewScope())
	if len(errors) > 0 {
		return
	}
	declType := d.Type.Get()
	bodyType := d.Body.GetType()

	if !typeEqual(declType, bodyType) {
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
	d.Type.Expression.SetType(TypeType{})
	deps = append(deps, Query{
		Expression:  d.Type.Expression,
		Destination: SetType{Type: &d.Type.value},
	})
	deps = append(deps, d.Body.GetTypeDependencies()...)
	return
}

// GetValueDependencies implements Declaration.
func (d ConstantDeclaration) GetValueDependencies() []Name {
	return d.Body.GetValueDependencies()
}

// Lower implements Declaration.
func (d ConstantDeclaration) Lower() cpp.Declaration {
	return cpp.ConstantDeclaration(
		d.Name.String,
		d.Type.Lower(),
		cpp.LambdaBlock(d.Body.lowerStatements()),
	)
}

// GetType implements Declaration.
func (d ConstantDeclaration) GetDeclaredType() TypeValue {
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
	Lower() cpp.Declaration
}

// TODO: when types and type aliases can be created, make sure that
// values are cached and aliases are properly resolved.

var _ TopLevelDeclaration = &FunctionDeclaration{}
var _ TopLevelDeclaration = &ConstantDeclaration{}
