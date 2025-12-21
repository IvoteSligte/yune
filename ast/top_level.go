package ast

import (
	"log"
	"yune/cpp"
	"yune/util"

	mapset "github.com/deckarep/golang-set/v2"
)

type Module struct {
	Declarations []Declaration
}

func (m *Module) Lower() (lowered cpp.Module, errors Errors) {
	declarations := map[string]Declaration{}

	// get unique mapping of name -> declaration
	for i := range m.Declarations {
		name := m.Declarations[i].GetName()
		other, exists := declarations[name]

		if exists {
			errors = append(errors, DuplicateDeclaration{
				First:  other,
				Second: m.Declarations[i],
			})
		} else {
			declarations[name] = m.Declarations[i]
		}
	}
	graph := map[string]stageNode{}

	// detect dependency cycles
	for i := range m.Declarations {
		name := m.Declarations[i].GetName()
		deps := m.Declarations[i].GetTypeDependencies()
		for _, dep := range deps {
			depDeps, ok := graph[dep]

			if ok && depDeps.priors.Contains(name) {
				errors = append(errors, CyclicDependency{
					First:  dep,
					Second: name,
				})
				break
			}
		}
		graph[name] = stageNode{
			priors: mapset.NewSet(deps...),
			simuls: mapset.NewSet(m.Declarations[i].GetValueDependencies()...),
		}
	}
	table := DeclarationTable{
		parent:       &DeclarationTable{declarations: BuiltinDeclarations},
		declarations: declarations,
	}
	// remove links to builtins to prevent them from being calculated
	for _, deps := range graph {
		deps.priors.RemoveAll(BuiltinNames...)
		deps.simuls.RemoveAll(BuiltinNames...)
	}
	ordering := stagedOrdering(graph)

	if len(ordering) > 1 {
		log.Fatalln("Multiple compilation stages are currently not supported.")
	}

	for i, stage := range ordering {
		lowered = cpp.Module{
			Declarations: stage.getPrefix(cache),
		}
		// add the actual code
		for name := range stage {
			decl := declarations[name].(TopLevelDeclaration)
			decl.CalcType(table)
			errors = append(errors, decl.TypeCheckBody(table)...)
			if len(errors) > 0 {
				return
			}
			// TODO: cache the serialized value instead of the raw cpp code so that it's only run once
			cppDeclaration := decl.Lower()
			decl.SetValue(cppDeclaration)
			lowered.Declarations = append(lowered.Declarations, cppDeclaration)
		}
		// the last lowered stage is simply the runtime code
		if i+1 == len(ordering) {
			return
		}
	}
	return
}

type FunctionDeclaration struct {
	Span
	Name       Name
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
	Value      cpp.FunctionDeclaration
}

// GetValue implements TopLevelDeclaration.
func (d *FunctionDeclaration) GetValue() Value {
	return CppASTValue{d.Value}
}

// TypeCheckBody implements Declaration.
func (d *FunctionDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	deps = deps.NewScope()
	deps.declarations = map[string]Declaration{d.GetName(): d}
	for i := range d.Parameters {
		param := &d.Parameters[i]
		println("Param", param.GetName(), ":", param.GetType().String())
		deps.declarations[param.GetName()] = param
	}
	errors = d.Body.InferType(deps)
	if len(errors) > 0 {
		return
	}
	returnType := d.ReturnType.Get()
	bodyType := d.Body.GetType()

	if !returnType.Eq(bodyType) {
		errors = append(errors, TypeMismatch{
			Expected: returnType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
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
func (d FunctionDeclaration) GetTypeDependencies() (deps []string) {
	deps = util.FlatMap(d.Parameters, FunctionParameter.GetTypeDependencies)
	deps = append(deps, d.ReturnType.GetValueDependencies()...)
	return
}

// CalcType implements Declaration.
func (d *FunctionDeclaration) CalcType(deps DeclarationTable) (errors Errors) {
	for i := range d.Parameters {
		errors = append(errors, d.Parameters[i].CalcType(deps)...)
	}
	errors = append(errors, d.ReturnType.Calc(deps)...)
	return
}

// Lower implements Declaration.
func (d FunctionDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.FunctionDeclaration{
		Name:       d.Name.String,
		Parameters: util.Map(d.Parameters, FunctionParameter.Lower),
		ReturnType: d.ReturnType.Lower(),
		Body:       d.Body.Lower(),
	}
}

func (d FunctionDeclaration) GetName() string {
	return d.Name.String
}

func (d FunctionDeclaration) GetType() InferredType {
	return InferredType{
		name:     "Fn",
		generics: append(util.Map(d.Parameters, FunctionParameter.GetType), d.ReturnType.Get()),
	}
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
func (d FunctionParameter) GetType() InferredType {
	return d.Type.Get()
}

// GetTypeDependencies implements Declaration
func (d FunctionParameter) GetTypeDependencies() []string {
	return d.Type.GetValueDependencies()
}

// GetValueDependencies implements Declaration
func (d FunctionParameter) GetValueDependencies() (deps []string) {
	return
}

// CalcType implements Declaration
func (d FunctionParameter) CalcType(deps DeclarationTable) Errors {
	return d.Type.Calc(deps)
}

type ConstantDeclaration struct {
	Span
	Name  Name
	Type  Type
	Body  Block
	Value Value
}

// GetValue implements TopLevelDeclaration.
func (d *ConstantDeclaration) GetValue() Value {
	return d.Value
}

// TypeCheckBody implements TopLevelDeclaration.
func (d *ConstantDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	errors = d.Body.InferType(deps)
	if len(errors) > 0 {
		return
	}
	returnType := d.Type.Get()
	bodyType := d.Body.GetType()

	if !returnType.Eq(bodyType) {
		errors = append(errors, TypeMismatch{
			Expected: returnType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
		})
	}
	return
}

// GetTypeDependencies implements Declaration.
func (d ConstantDeclaration) GetTypeDependencies() []string {
	return d.Type.GetValueDependencies()
}

// GetValueDependencies implements Declaration.
func (d ConstantDeclaration) GetValueDependencies() []string {
	return append(d.GetTypeDependencies(), d.Body.GetValueDependencies()...)
}

// InferType implements Declaration.
func (d *ConstantDeclaration) CalcType(deps DeclarationTable) Errors {
	return d.Type.Calc(deps)
}

// Lower implements Declaration.
func (d ConstantDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.ConstantDeclaration{
		Name:  d.Name.String,
		Type:  d.Type.Lower(),
		Value: d.Body.Lower(),
	}
}

// GetType implements Declaration.
func (d ConstantDeclaration) GetType() InferredType {
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
	Lower() cpp.TopLevelDeclaration
	// Returns the value computed by running the code in
	// the declaration's body if this is a constant.
	// If this is not a constant, this function returns
	// the C++ code of the declaration.
	GetValue() Value
}

var _ TopLevelDeclaration = &FunctionDeclaration{}
var _ TopLevelDeclaration = &ConstantDeclaration{}
