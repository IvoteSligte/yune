package ast

import (
	"log"
	"maps"
	"slices"
	"yune/cpp"
	"yune/util"

	mapset "github.com/deckarep/golang-set/v2"
)

type Module struct {
	Declarations []TopLevelDeclaration
}

func (m *Module) Lower() (lowered cpp.Module, errors Errors) {
	declarations := map[string]TopLevelDeclaration{}

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
	if len(errors) > 0 {
		return
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
		valueDeps := m.Declarations[i].GetValueDependencies()
		// TEMP (should be a unit test)
		for _, t := range deps {
			if len(t) == 0 {
				log.Fatalf("Empty type dependency of declaration '%s'", name)
			}
		}
		// TEMP (should be a unit test)
		for _, t := range valueDeps {
			if len(t) == 0 {
				log.Fatalf("Empty value dependency of declaration '%s'", name)
			}
		}
		graph[name] = stageNode{
			priors: mapset.NewSet(deps...),
			simuls: mapset.NewSet(valueDeps...),
		}
	}
	table := DeclarationTable{
		parent: &DeclarationTable{declarations: BuiltinDeclarations},
		declarations: util.MapMap(declarations, func(name string, decl TopLevelDeclaration) (string, Declaration) {
			return name, decl
		}),
	}
	// remove links to builtins to prevent them from being calculated
	for _, deps := range graph {
		deps.priors.RemoveAll(slices.Collect(maps.Keys(BuiltinDeclarations))...)
		deps.simuls.RemoveAll(slices.Collect(maps.Keys(BuiltinDeclarations))...)
	}
	ordering := stagedOrdering(graph)

	if len(ordering) > 1 {
		log.Fatalln("Multiple compilation stages are currently not supported.")
	}
	for i, stage := range ordering {
		lowered = cpp.Module{
			Declarations: stage.getPrefix(declarations),
		}
		// add the actual code
		for name := range stage {
			decl := declarations[name]
			decl.CalcType(table)
			errors = append(errors, decl.TypeCheckBody(table)...)
			if len(errors) > 0 {
				return
			}
			// TODO: cache the serialized value instead of the raw cpp code so that it's only run once
			cppDeclaration := decl.Lower()
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

// TypeCheckBody implements Declaration.
func (d *FunctionDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	deps = deps.NewScope()
	deps.declarations = map[string]Declaration{d.GetName(): d}
	for i := range d.Parameters {
		param := &d.Parameters[i]
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

func (d FunctionDeclaration) GetType() cpp.Type {
	return cpp.Type{
		Name:     "Fn",
		Generics: append(util.Map(d.Parameters, FunctionParameter.GetType), d.ReturnType.Get()),
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
func (d FunctionParameter) GetType() cpp.Type {
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
	Name Name
	Type Type
	Body Block
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
func (d ConstantDeclaration) GetType() cpp.Type {
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

// TODO: when types and type aliases can be created, make sure that
// values are cached and aliases are properly resolved.

var _ TopLevelDeclaration = &FunctionDeclaration{}
var _ TopLevelDeclaration = &ConstantDeclaration{}
