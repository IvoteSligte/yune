package ast

import (
	"log"
	"yune/cpp"
	"yune/util"
	"yune/value"

	mapset "github.com/deckarep/golang-set/v2"
)

type Module struct {
	Declarations []TopLevelDeclaration
}

func (m Module) Lower() (lowered cpp.Module, errors Errors) {
	// thread unsafe set used so that iteration and removal can be done simultaneously
	// (deadlocks otherwise)
	stageNodes := mapset.NewThreadUnsafeSet[*stageNode]()
	declarationToNode := map[string]*stageNode{}

	// Add builtin declarations to the module's list of  declarations.
	m.Declarations = append(BuiltinDeclarations, m.Declarations...)

	// get unique mapping of name -> declaration
	for i := range m.Declarations {
		name := m.Declarations[i].GetName()
		other, exists := declarationToNode[name.String]

		if exists {
			errors = append(errors, DuplicateDeclaration{
				First:  other.Declaration,
				Second: m.Declarations[i],
			})
		} else {
			_, isRawBuiltin := m.Declarations[i].(BuiltinRawDeclaration)
			node := &stageNode{
				Query:       Query{},
				Declaration: m.Declarations[i],
				After:       nil, // set later
				Requires:    nil, // set later
				// Raw nodes are precalculated and expected to be available by other nodes.
				ExecuteFirst: isRawBuiltin,
			}
			declarationToNode[name.String] = node
			stageNodes.Add(node)
		}
	}
	if len(errors) > 0 {
		return
	}

	// detect dependency cycles
	for i := range m.Declarations {
		name := m.Declarations[i].GetName()
		typeDependencies := mapset.NewThreadUnsafeSet[*stageNode]()
		valueDependencies := mapset.NewThreadUnsafeSet[*stageNode]()

		for _, query := range m.Declarations[i].GetTypeDependencies() {
			depNames := query.Expression.GetValueDependencies()
			requires := mapset.NewThreadUnsafeSet[*stageNode]()
			for _, depName := range depNames {
				if len(depName.String) == 0 {
					log.Printf("WARN: Empty string name of type dependency of declaration '%s'.", name)
				}
				requiredNode, exists := declarationToNode[depName.String]
				if !exists {
					errors = append(errors, UndefinedType(depName))
				}
				requires.Add(requiredNode)
			}
			node := &stageNode{
				Query:       query,
				Declaration: nil,
				After:       mapset.NewThreadUnsafeSet[*stageNode](),
				Requires:    requires,
			}
			typeDependencies.Add(node)
			stageNodes.Add(node)
		}
		for _, depName := range m.Declarations[i].GetValueDependencies() {
			if len(depName.String) == 0 {
				log.Printf("WARN: Empty string name of value dependency of declaration '%s'.", name)
			}
			depNode, exists := declarationToNode[depName.String]
			if !exists {
				errors = append(errors, UndefinedVariable(depName))
			}
			valueDependencies.Add(depNode)
		}
		if len(errors) > 0 {
			return
		}
		node := declarationToNode[name.String]
		node.After = typeDependencies
		node.Requires = valueDependencies
	}
	errors = append(errors, CheckCyclicType(stageNodes)...)
	errors = append(errors, CheckCyclicConstant(stageNodes)...)
	if len(errors) > 0 {
		return
	}
	table := DeclarationTable{
		parent:       nil,
		declarations: map[string]Declaration{},
	}
	for name, node := range declarationToNode {
		table.declarations[name] = node.Declaration
	}
	evaluatedNodes := mapset.NewThreadUnsafeSet[*stageNode]()
	ordering := stagedOrdering(stageNodes)
	for i, stage := range ordering {
		evalNodes := extractSortedNames(stage, evaluatedNodes)
		if i == 0 {
			util.PrettyPrint(evalNodes)
		}

		// type check all expressions and declarations
		for _, node := range evalNodes {
			query := node.Query
			if query.Expression != nil {
				errors = append(errors, query.Expression.InferType(query.ExpectedType, table)...)
				if len(errors) == 0 && !query.GetType().Eq(query.ExpectedType) {
					errors = append(errors, UnexpectedType{
						Expected: query.ExpectedType,
						Found:    query.GetType(),
						At:       query.GetSpan(),
					})
				}
			}
			if node.Declaration != nil {
				errors = append(errors, node.Declaration.TypeCheckBody(table)...)
			}
		}
		if len(errors) > 0 {
			return
		}
		// add the actual code
		for _, node := range evalNodes {
			if node.Declaration != nil {
				lowered.Declarations = append(lowered.Declarations, node.Declaration.Lower())
			}
		}
		// the last lowered stage is simply the runtime code
		if i+1 == len(ordering) {
			for _, node := range evalNodes {
				if node.Query.Expression != nil {
					// should be unreachable
					log.Fatalln("Unreachable: Last compilation stage (runtime) has expression queued. Expression:", node.Query.Expression)
				}
			}
			return
		}
		values := cpp.Evaluate(lowered, util.Map(evalNodes, func(node *stageNode) cpp.Expression {
			if node.Query.Expression != nil {
				return node.Query.Expression.Lower()
			} else {
				return nil
			}
		}))
		for i, v := range values {
			if evalNodes[i].Query.Expression == nil {
				if v != value.Value("") {
					log.Fatalf("Passed nil expression to the C++ evaluator, but received non-empty string '%s'.", v)
				}
			} else {
				evalNodes[i].Query.SetValue(string(v))
			}
		}
		evaluatedNodes.Append(evalNodes...)
	}
	return
}

func CheckCyclicType(stageNodes stage) (errors Errors) {
	for node := range stageNodes.Iter() {
		queue := node.After.ToSlice()
		visited := mapset.NewThreadUnsafeSet[*stageNode]()

		for len(queue) > 0 {
			dep := queue[0]
			queue = queue[1:]
			if visited.Contains(dep) {
				continue
			}
			visited.Add(dep)
			queue = append(queue, dep.After.ToSlice()...)
		}
		if visited.Contains(node) {
			errors = append(errors, CyclicTypeDependency{
				In: node.Declaration,
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
func CheckCyclicConstant(stageNodes stage) (errors Errors) {
	for node := range stageNodes.Iter() {
		if !isConstantDeclaration(node.Declaration) {
			continue
		}
		queue := node.Requires.ToSlice()
		visited := mapset.NewThreadUnsafeSet[*stageNode]()

		for len(queue) > 0 {
			dep := queue[0]
			queue = queue[1:]
			if !isConstantDeclaration(dep.Declaration) {
				continue
			}
			if visited.Contains(dep) {
				continue
			}
			visited.Add(dep)
			queue = append(queue, dep.Requires.ToSlice()...)
		}
		if visited.Contains(node) {
			errors = append(errors, CyclicConstantDependency{
				In: node.Declaration,
			})
		}
	}
	return
}
