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
	stageNodes := mapset.NewSet[*stageNode]()
	declarationToNode := map[string]*stageNode{}

	for name, builtin := range BuiltinDeclarations {
		node := &stageNode{
			expression:  nil,
			destination: nil,
			declaration: builtin,
			after:       nil,
			requires:    nil,
		}
		stageNodes.Add(node)
		declarationToNode[name] = node
	}

	// get unique mapping of name -> declaration
	for i := range m.Declarations {
		name := m.Declarations[i].GetName()
		other, exists := declarationToNode[name]

		if exists {
			errors = append(errors, DuplicateDeclaration{
				First:  other.declaration,
				Second: m.Declarations[i],
			})
		} else {
			node := &stageNode{
				expression:  nil,
				destination: nil,
				declaration: m.Declarations[i],
				after:       nil, // set later
				requires:    nil, // set later
			}
			declarationToNode[name] = node
			stageNodes.Add(node)
		}
	}
	if len(errors) > 0 {
		return
	}
	getDeclarationNode := func(name string) *stageNode {
		node, exists := declarationToNode[name]
		if !exists {
			errors = append(errors, UndefinedType{
				Span:   Span{}, // TODO: span
				String: name,
			})
		}
		return node
	}

	// detect dependency cycles
	for i := range m.Declarations {
		name := m.Declarations[i].GetName()
		typeDependencies := mapset.NewSet[*stageNode]()
		valueDependencies := mapset.NewSet[*stageNode]()

		for _, typeExpression := range m.Declarations[i].GetTypeDependencies() {
			depNames := typeExpression.expression.GetGlobalDependencies()
			requires := mapset.NewSet[*stageNode]()
			for _, depName := range depNames {
				if len(depName) == 0 {
					log.Printf("WARN: Empty string name of type dependency of declaration '%s'.", name)
				}
				requires.Add(getDeclarationNode(depName))
			}
			typeDependencies.Add(&stageNode{
				expression:  typeExpression.expression,
				destination: &typeExpression.value,
				declaration: nil,
				after:       nil,
				requires:    requires,
			})
		}
		for _, d := range m.Declarations[i].GetValueDependencies() {
			if len(d) == 0 {
				log.Printf("WARN: Empty string name of value dependency of declaration '%s'.", name)
			}
			valueDependencies.Add(getDeclarationNode(d))
		}
		if len(errors) > 0 {
			return
		}
		node := declarationToNode[name]
		node.after = typeDependencies
		node.requires = valueDependencies
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
		table.declarations[name] = node.declaration
	}
	ordering := stagedOrdering(stageNodes)
	for i, stage := range ordering {
		evalNodes := extractSortedNames(stage)
		// type check all expressions and declarations
		for _, node := range evalNodes {
			if node.expression != nil {
				errors = append(errors, node.expression.InferType(table)...)
				// TODO: allow other types as well
				if len(errors) == 0 && !node.expression.GetType().Eq(TypeType) {
					errors = append(errors, NotAType{
						Found: node.expression.GetType(),
						At:    node.expression.GetSpan(),
					})
				}
			}
			if node.declaration != nil {
				errors = append(errors, node.declaration.TypeCheckBody(table)...)
			}
		}
		if len(errors) > 0 {
			return
		}
		// add the actual code
		for _, node := range evalNodes {
			if node.declaration != nil {
				lowered.Declarations = append(lowered.Declarations, node.declaration.Lower())
			}
		}
		// the last lowered stage is simply the runtime code
		if i+1 == len(ordering) {
			for _, node := range evalNodes {
				if node.expression != nil {
					// should be unreachable
					log.Fatalln("Unreachable: Last compilation stage (runtime) has expression queued. Expression:", node.expression)
				}
			}
			return
		}
		values := cpp.Evaluate(lowered, util.Map(evalNodes, func(node *stageNode) cpp.Expression {
			return node.expression.Lower()
		}))
		for i, v := range values {
			if evalNodes[i].expression == nil {
				if v != value.Value("") {
					log.Fatalf("Passed nil expression to the C++ evaluator, but received non-empty string '%s'.", v)
				}
			} else {
				*evalNodes[i].destination = value.Type(string(v))
			}
		}
	}
	return
}

func mapContains[K comparable, V any, M map[K]V](m M, key K) bool {
	_, exists := m[key]
	return exists
}

func CheckCyclicType(stageNodes stage) (errors Errors) {
	for node := range stageNodes.Iter() {
		queue := node.after.ToSlice()
		visited := mapset.NewSet[*stageNode]()

		for len(queue) > 0 {
			dep := queue[0]
			queue = queue[1:]
			if visited.Contains(dep) {
				continue
			}
			visited.Add(dep)
			queue = append(queue, dep.after.ToSlice()...)
		}
		if visited.Contains(node) {
			errors = append(errors, CyclicTypeDependency{
				In: node.declaration,
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
		if !isConstantDeclaration(node.declaration) {
			continue
		}
		queue := node.requires.ToSlice()
		visited := mapset.NewSet[*stageNode]()

		for len(queue) > 0 {
			dep := queue[0]
			queue = queue[1:]
			if !isConstantDeclaration(dep.declaration) {
				continue
			}
			if visited.Contains(dep) {
				continue
			}
			visited.Add(dep)
			queue = append(queue, dep.requires.ToSlice()...)
		}
		if visited.Contains(node) {
			errors = append(errors, CyclicConstantDependency{
				In: node.declaration,
			})
		}
	}
	return
}
