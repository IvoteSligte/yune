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

func (m *Module) Lower() (lowered cpp.Module, errors Errors) {
	// thread unsafe set used so that iteration and removal can be done simultaneously
	// (deadlocks otherwise)
	stageNodes := mapset.NewThreadUnsafeSet[*stageNode]()
	declarationToNode := map[string]*stageNode{}

	for _, builtin := range BuiltinDeclarations {
		node := &stageNode{
			Expression:  nil,
			Destination: nil,
			Declaration: builtin,
			After:       mapset.NewSet[*stageNode](),
			Requires:    mapset.NewSet[*stageNode](),
		}
		stageNodes.Add(node)
		declarationToNode[builtin.GetName().String] = node
	}

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
			node := &stageNode{
				Expression:  nil,
				Destination: nil,
				Declaration: m.Declarations[i],
				After:       nil, // set later
				Requires:    nil, // set later
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
		typeDependencies := mapset.NewSet[*stageNode]()
		valueDependencies := mapset.NewSet[*stageNode]()

		for _, typeExpression := range m.Declarations[i].GetTypeDependencies() {
			depNames := typeExpression.Expression.GetGlobalDependencies()
			requires := mapset.NewSet[*stageNode]()
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
				Expression:  typeExpression.Expression,
				Destination: &typeExpression.value,
				Declaration: nil,
				After:       mapset.NewSet[*stageNode](),
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
	ordering := stagedOrdering(stageNodes)
	for i, stage := range ordering {
		evalNodes := extractSortedNames(stage)

		// type check all expressions and declarations
		for _, node := range evalNodes {
			if node.Expression != nil {
				// TODO: allow other types for expressions as well
				errors = append(errors, node.Expression.InferType(TypeType, table)...)
				if len(errors) == 0 && !node.Expression.GetType().Eq(TypeType) {
					errors = append(errors, NotAType{
						Found: node.Expression.GetType(),
						At:    node.Expression.GetSpan(),
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
				if node.Expression != nil {
					// should be unreachable
					log.Fatalln("Unreachable: Last compilation stage (runtime) has expression queued. Expression:", node.Expression)
				}
			}
			return
		}
		values := cpp.Evaluate(lowered, util.Map(evalNodes, func(node *stageNode) cpp.Expression {
			if node.Expression != nil {
				return node.Expression.Lower()
			} else {
				return nil
			}
		}))
		for i, v := range values {
			if evalNodes[i].Expression == nil {
				if v != value.Value("") {
					log.Fatalf("Passed nil expression to the C++ evaluator, but received non-empty string '%s'.", v)
				}
			} else {
				*evalNodes[i].Destination = value.Type(string(v))
			}
		}
	}
	return
}

func CheckCyclicType(stageNodes stage) (errors Errors) {
	for node := range stageNodes.Iter() {
		queue := node.After.ToSlice()
		visited := mapset.NewSet[*stageNode]()

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
		visited := mapset.NewSet[*stageNode]()

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
