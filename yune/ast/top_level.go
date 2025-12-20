package ast

import (
	"yune/cpp"
	"yune/util"

	mapset "github.com/deckarep/golang-set/v2"
)

type Module struct {
	Declarations []TopLevelDeclaration
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
	cache := map[string]cpp.TopLevelDeclaration{}

	// remove links to builtins to prevent them from being calculated
	for _, deps := range graph {
		deps.priors.RemoveAll(BuiltinNames...)
		deps.simuls.RemoveAll(BuiltinNames...)
	}
	ordering := stagedOrdering(graph)

	for i, stage := range ordering {
		lowered = cpp.Module{
			Declarations: stage.getPrefix(cache),
		}
		// add the actual code
		for name := range stage {
			decl := declarations[name]
			errors = append(errors, decl.CalcType(table)...)
			if len(errors) > 0 {
				return
			}
			errors = append(errors, decl.TypeCheck(table)...)
			if len(errors) > 0 {
				return
			}
			// TODO: cache the serialized value instead of the raw cpp code so that it's only run once
			cppDeclaration := decl.Lower()
			cache[name] = cppDeclaration
			lowered.Declarations = append(lowered.Declarations, cppDeclaration)
		}
		// the last lowered stage is simply the runtime code
		if i+1 == len(ordering) {
			return
		}
	}
	return
}

func (m *Module) InferType(deps DeclarationTable) (errors Errors) {
	for i := range m.Declarations {
		errors = append(errors, m.Declarations[i].InferType(deps)...)
	}
	return
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
	errors = d.Body.InferType(deps.NewScope())
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
