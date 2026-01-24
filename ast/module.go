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
		typeExprs := m.Declarations[i].GetTypeDependencies()
		valueDeps := m.Declarations[i].GetValueDependencies()

		// TEMP (should be a unit test)
		for _, t := range typeDeps {
			if len(t) == 0 {
				log.Printf("WARN: Empty string name of type dependency of declaration '%s'.", name)
			}
		}
		// TEMP (should be a unit test)
		for _, t := range valueDeps {
			if len(t) == 0 {
				log.Printf("WARN: Empty string name of value dependency of declaration '%s'.", name)
			}
		}
		graph[name] = stageNode{
			priors: mapset.NewSet(typeDeps...),
			simuls: mapset.NewSet(valueDeps...),
		}
	}
	errors = append(errors, CheckUndefinedDependencies(declarations, graph)...)
	if len(errors) > 0 {
		return
	}
	errors = append(errors, CheckCyclicType(declarations, graph)...)
	errors = append(errors, CheckCyclicConstant(declarations, graph)...)
	if len(errors) > 0 {
		return
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
	priorDeclarations := util.Map(
		slices.Collect(maps.Values(BuiltinDeclarations)),
		func(d Declaration) cpp.TopLevelDeclaration { return d.(TopLevelDeclaration).Lower() },
	)
	for i, stage := range ordering {
		lowered = cpp.Module{
			Declarations: priorDeclarations,
		}
		declarationNames := stage.extractSortedNames()
		// compute signatures
		for _, name := range declarationNames {
			decl := declarations[name]
			errors = append(errors, decl.CalcType(table)...)
		}
		if len(errors) > 0 {
			return
		}
		// type check bodies
		for _, name := range declarationNames {
			decl := declarations[name]
			errors = append(errors, decl.TypeCheckBody(table)...)
		}
		if len(errors) > 0 {
			return
		}
		// add the actual code
		for _, name := range declarationNames {
			decl := declarations[name]
			// NOTE: declarations are added in a random order because of the map
			cppDeclaration := decl.Lower()
			lowered.Declarations = append(lowered.Declarations, cppDeclaration)
		}
		// the last lowered stage is simply the runtime code
		if i+1 == len(ordering) {
			return
		}
		priorDeclarations = append(priorDeclarations, lowered.Declarations...)
	}
	return
}

func mapContains[K comparable, V any, M map[K]V](m M, key K) bool {
	_, exists := m[key]
	return exists
}

func CheckUndefinedDependencies(declarations map[string]TopLevelDeclaration, graph map[string]stageNode) (errors Errors) {
	for _, node := range graph {
		for _, prior := range node.priors.ToSlice() {
			if !mapContains(declarations, prior) && !mapContains(BuiltinDeclarations, prior) {
				errors = append(errors, UndefinedVariable{
					// TODO: span
					String: prior,
				})
			}
		}
		for _, simul := range node.simuls.ToSlice() {
			if !mapContains(declarations, simul) && !mapContains(BuiltinDeclarations, simul) {
				errors = append(errors, UndefinedVariable{
					// TODO: span
					String: simul,
				})
			}
		}
	}
	return
}

func CheckCyclicType(declarations map[string]TopLevelDeclaration, graph map[string]stageNode) (errors Errors) {
	for name, node := range graph {
		queue := node.priors.ToSlice()
		visited := mapset.NewSet[string]()

		for len(queue) > 0 {
			dep := queue[0]
			queue = queue[1:]
			_, isBuiltin := BuiltinDeclarations[dep]
			if isBuiltin {
				continue
			}
			if visited.Contains(dep) {
				continue
			}
			visited.Add(dep)
			queue = append(queue, graph[dep].priors.ToSlice()...)
		}
		if visited.Contains(name) {
			errors = append(errors, CyclicTypeDependency{
				In: declarations[name],
			})
		}
	}
	return
}

// NOTE: an uncommon edge case that is currently not handled is when a constant depends on another constant
// through a function call and that forms a cycle
// ```
// f(): Int = A
// A: Int = B
// B: Int = f()
// ```
func CheckCyclicConstant(declarations map[string]TopLevelDeclaration, graph map[string]stageNode) (errors Errors) {
	for name, node := range graph {
		if !isConstantDeclaration(declarations[name]) {
			continue
		}
		queue := node.simuls.ToSlice()
		visited := mapset.NewSet[string]()

		for len(queue) > 0 {
			dep := queue[0]
			queue = queue[1:]
			_, isBuiltin := BuiltinDeclarations[dep]
			if isBuiltin {
				continue
			}
			if !isConstantDeclaration(declarations[dep]) {
				continue
			}
			if visited.Contains(dep) {
				continue
			}
			visited.Add(dep)
			queue = append(queue, graph[dep].simuls.ToSlice()...)
		}
		if visited.Contains(name) {
			errors = append(errors, CyclicConstantDependency{
				In: declarations[name],
			})
		}
	}
	return
}
