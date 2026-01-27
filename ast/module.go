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
	stageNodes := mapset.NewThreadUnsafeSet[*evalNode]()
	declarationToNode := map[string]*evalNode{}

	// Add builtin declarations to the module's list of  declarations.
	m.Declarations = append(BuiltinDeclarations, m.Declarations...)

	// get unique mapping of name -> declaration
	for _, decl := range m.Declarations {
		name := decl.GetName()
		other, exists := declarationToNode[name.String]

		if exists {
			errors = append(errors, DuplicateDeclaration{
				First:  other.Declaration,
				Second: decl,
			})
		} else {
			_, isRawBuiltin := decl.(BuiltinRawDeclaration)
			node := &evalNode{
				Query:         Query{},
				Declaration:   decl,
				After:         mapset.NewThreadUnsafeSet[*evalNode](), // filled later
				Requires:      mapset.NewThreadUnsafeSet[*evalNode](), // filled later
				IsPrecomputed: isRawBuiltin,
			}
			declarationToNode[name.String] = node
			stageNodes.Add(node)
		}
	}
	if len(errors) > 0 {
		return
	}

	for name, node := range declarationToNode {
		decl := node.Declaration

		// for _, macro := range decl.GetMacros() {
		// }
		for _, query := range decl.GetTypeDependencies() {
			depNames := query.Expression.GetValueDependencies()
			requires := mapset.NewThreadUnsafeSet[*evalNode]()
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
			depNode := &evalNode{
				Query:       query,
				Declaration: nil,
				After:       mapset.NewThreadUnsafeSet[*evalNode](),
				Requires:    requires,
			}
			node.After.Add(depNode)
			stageNodes.Add(depNode)
		}
		for _, depName := range decl.GetValueDependencies() {
			if len(depName.String) == 0 {
				log.Printf("WARN: Empty string name of value dependency of declaration '%s'.", name)
			}
			depNode, exists := declarationToNode[depName.String]
			if !exists {
				errors = append(errors, UndefinedVariable(depName))
			}
			node.Requires.Add(depNode)
		}
		if len(errors) > 0 {
			return
		}
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
	evaluated := mapset.NewThreadUnsafeSet[*evalNode]()
	unevaluated := stageNodes

	// handle precomputed nodes
	for node := range unevaluated.Iter() {
		if node.IsPrecomputed {
			unevaluated.Remove(node)
			evaluated.Add(node)
			lowered.Declarations = append(lowered.Declarations, node.Declaration.Lower())
		}
	}
	// iteratively evaluate nodes
	for unevaluated.Cardinality() > 0 {
		evalNodes := extractEvaluatableNodes(unevaluated, evaluated)

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
		if unevaluated.Cardinality() == 0 {
			for _, node := range evalNodes {
				if node.Query.Expression != nil {
					// should be unreachable
					log.Fatalln("Unreachable: Last compilation stage (runtime) has expression queued. Expression:", node.Query.Expression)
				}
			}
			return
		}
		// TODO: make sure the main() function is always in the last stage
		values := cpp.Evaluate(lowered, util.Map(evalNodes, func(node *evalNode) cpp.Expression {
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
		evaluated.Append(evalNodes...)
	}
	return
}

func CheckCyclicType(stageNodes evalSet) (errors Errors) {
	for node := range stageNodes.Iter() {
		queue := node.After.ToSlice()
		visited := mapset.NewThreadUnsafeSet[*evalNode]()

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
func CheckCyclicConstant(stageNodes evalSet) (errors Errors) {
	for node := range stageNodes.Iter() {
		if !isConstantDeclaration(node.Declaration) {
			continue
		}
		queue := node.Requires.ToSlice()
		visited := mapset.NewThreadUnsafeSet[*evalNode]()

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
