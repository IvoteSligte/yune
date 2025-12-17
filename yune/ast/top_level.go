package ast

import (
	"yune/cpp"
	"yune/util"
)

type Module struct {
	Declarations []TopLevelDeclaration
}

func (m *Module) ResolveDependencies() {
	for i := range m.Declarations {
		m.Declarations[i].GetTypeDependencies()
		m.Declarations[i].GetValueDependencies()
	}
}

func (m *Module) InferType(deps DeclarationTable) (errors Errors) {
	for i := range m.Declarations {
		errors = append(errors, m.Declarations[i].InferType(deps)...)
	}
	return
}

func (m Module) Lower() cpp.Module {
	return cpp.Module{
		Declarations: util.Map(m.Declarations, TopLevelDeclaration.Lower),
	}
}

type FunctionDeclaration struct {
	Span
	Name       Name
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
}

// GetValueDependencies implements TopLevelDeclaration.
func (d FunctionDeclaration) GetValueDependencies() []string {
	locals := DeclarationTable{}
	for _, param := range d.Parameters {
		locals.Add(param)
	}
	return append(d.GetTypeDependencies(), d.Body.GetValueDependencies(locals)...)
}

// CalcValue implements TopLevelDeclaration.
func (d FunctionDeclaration) CalcValue(deps DeclarationTable) (result Value, errors Errors) {
	panic("unimplemented")
}

// GetTypeDependencies implements TopLevelDeclaration.
func (d FunctionDeclaration) GetTypeDependencies() (deps []string) {
	deps = util.FlatMap(d.Parameters, FunctionParameter.GetTypeDependencies)
	deps = append(deps, d.ReturnType.GetValueDependencies()...)
	return
}

// InferType implements TopLevelDeclaration.
func (d *FunctionDeclaration) InferType(deps DeclarationTable) (errors Errors) {
	for i := range d.Parameters {
		errors = append(errors, d.Parameters[i].InferType(deps)...)
	}
	errors = append(errors, d.ReturnType.InferType(deps)...)
	return
}

// Lower implements TopLevelDeclaration.
func (d FunctionDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.FunctionDeclaration{
		Name:       d.Name.String,
		Parameters: util.Map(d.Parameters, FunctionParameter.Lower),
		ReturnType: d.ReturnType.Lower(),
		Body:       d.Body.Lower(),
	}
}

func (FunctionDeclaration) topLevelDeclaration() {}

func (d FunctionDeclaration) GetName() string {
	return d.Name.String
}

func (d FunctionDeclaration) GetType() InferredType {
	return InferredType{
		name:     "Fn",
		generics: append(util.Map(d.Parameters, FunctionParameter.GetType), d.ReturnType.InferredType),
	}
}

type FunctionParameter struct {
	Span
	Name Name
	Type Type
}

func (d FunctionParameter) Lower() cpp.FunctionParameter {
	return cpp.FunctionParameter{
		Name: d.Name.String,
		Type: d.Type.Lower(),
	}
}

func (d FunctionParameter) GetName() string {
	return d.Name.String
}

func (d FunctionParameter) GetType() InferredType {
	return d.Type.InferredType
}

func (d FunctionParameter) GetTypeDependencies() []string {
	return d.Type.GetValueDependencies()
}

func (d FunctionParameter) InferType(deps DeclarationTable) (errors Errors) {
	return d.Type.InferType(deps)
}

type ConstantDeclaration struct {
	Span
	Name Name
	Type Type
	Body Block
}

// CalcValue implements TopLevelDeclaration.
func (d *ConstantDeclaration) CalcValue(deps DeclarationTable) (result Value, errors Errors) {
	errors = d.Body.InferType(deps)
	if len(errors) > 0 {
		return
	}
	// evaluate body with dependencies
	// cppBlock := d.Body.Lower()
	panic("unimplemented")
}

// GetTypeDependencies implements TopLevelDeclaration.
func (d ConstantDeclaration) GetTypeDependencies() []string {
	return d.Type.GetValueDependencies()
}

// GetValueDependencies implements TopLevelDeclaration.
func (d ConstantDeclaration) GetValueDependencies() []string {
	return append(d.GetTypeDependencies(), d.Body.GetValueDependencies(DeclarationTable{})...)
}

// InferType implements TopLevelDeclaration.
func (d *ConstantDeclaration) InferType(deps DeclarationTable) (errors Errors) {
	return d.Type.InferType(deps)
}

// Lower implements TopLevelDeclaration.
func (d ConstantDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.ConstantDeclaration{
		Name:  d.Name.String,
		Type:  d.Type.Lower(),
		Value: d.Body.Lower(),
	}
}

// GetType implements Declaration.
func (d ConstantDeclaration) GetType() InferredType {
	return d.Type.InferredType
}

func (ConstantDeclaration) topLevelDeclaration() {}

func (d ConstantDeclaration) GetName() string {
	return d.Name.String
}

func (d ConstantDeclaration) GetDeclarationType() Type {
	return d.Type
}

type TopLevelDeclaration interface {
	Declaration
	topLevelDeclaration()
	// A list of dependencies required to calculate the type of this declaration.
	// Subset of the result of GetValueDependencies().
	GetTypeDependencies() []string
	GetValueDependencies() []string
	// Calculates the value of the declaration, assuming InferType has been called.
	CalcValue(deps DeclarationTable) (Value, Errors)
	Lower() cpp.TopLevelDeclaration
}

var _ TopLevelDeclaration = &FunctionDeclaration{}
var _ TopLevelDeclaration = &ConstantDeclaration{}
