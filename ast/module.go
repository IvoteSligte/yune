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
		// TODO: cyclic dependency detection for non-function constants (using .simuls)
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
				log.Printf("WARN: Empty type dependency of declaration '%s'.", name)
			}
		}
		// TEMP (should be a unit test)
		for _, t := range valueDeps {
			if len(t) == 0 {
				log.Printf("WARN: Empty value dependency of declaration '%s'.", name)
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
			errors = append(errors, decl.CalcType(table)...)
			if len(errors) > 0 {
				return
			}
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
