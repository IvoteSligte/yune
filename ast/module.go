package ast

import (
	"log"
	"yune/cpp"
	"yune/util"
	"yune/value"

	mapset "github.com/deckarep/golang-set/v2"
)

func newQueryEvalNode(query Query, declarationToNode map[string]*evalNode, stageNodes evalSet) (node *evalNode, errors Errors) {
	if len(query.Expression.GetMacros()) != 0 {
		panic("Query expressions are assumed to not contain macros.")
	}
	if len(query.Expression.GetTypeDependencies()) != 0 {
		panic("Query expression type dependencies are assumed to not exist.")
	}
	depNames := query.Expression.GetValueDependencies()
	requires := mapset.NewThreadUnsafeSet[*evalNode]()
	for _, depName := range depNames {
		if len(depName.String) == 0 {
			log.Printf("WARN: Empty string name in dependency of query.")
		}
		requiredNode, exists := declarationToNode[depName.String]
		if !exists {
			errors = append(errors, UndefinedType(depName))
		}
		requires.Add(requiredNode)
	}
	node = &evalNode{
		Query:    query,
		After:    mapset.NewThreadUnsafeSet[*evalNode](),
		Requires: requires,
	}
	stageNodes.Add(node)
	return
}

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

		for _, macro := range decl.GetMacros() {
			call := macro.AsFunctionCall()
			query := Query{
				Expression:   &call,
				Destination:  macro,
				ExpectedType: MacroReturnType,
			}
			macroNode, _errors := newQueryEvalNode(query, declarationToNode, stageNodes)
			macroNode.UpdateHook = node
			errors = append(errors, _errors...)
			// another node for delay because the macro needs to be executed 2 stages
			// before the declaration itself (as it can add type dependencies to the declaration)
			depNode := &evalNode{
				After:    mapset.NewThreadUnsafeSet[*evalNode](macroNode),
				Requires: mapset.NewThreadUnsafeSet[*evalNode](),
			}
			node.After.Add(depNode)
			stageNodes.Add(depNode)
		}
		for _, query := range decl.GetTypeDependencies() {
			depNode, _errors := newQueryEvalNode(query, declarationToNode, stageNodes)
			errors = append(errors, _errors...)
			node.After.Add(depNode)
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
	}
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

	// add precomputed nodes to 'evaluated'
	for node := range unevaluated.Iter() {
		if node.IsPrecomputed {
			unevaluated.Remove(node)
			evaluated.Add(node)
		}
	}
	// sort precomputed nodes
	lowered.Declarations = util.Map(
		sortedEvaluatableNodes(evaluated.Clone(), mapset.NewThreadUnsafeSet[*evalNode]()),
		func(node *evalNode) cpp.TopLevelDeclaration {
			return node.Declaration.Lower()
		},
	)
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
				continue
			}
			evalNodes[i].Query.SetValue(string(v))

			// Update node that depends on the result of this query.
			node := evalNodes[i].UpdateHook
			decl := node.Declaration
			if decl == nil {
				// NOTE: can non-declarations also contain macros?
				continue
			}
			for _, query := range decl.GetMacroTypeDependencies() {
				depNode, _errors := newQueryEvalNode(query, declarationToNode, stageNodes)
				errors = append(errors, _errors...)
				node.After.Add(depNode)
			}
			for _, depName := range decl.GetMacroValueDependencies() {
				if len(depName.String) == 0 {
					log.Printf("WARN: Empty string name of value dependency of declaration '%s'.", decl.GetName().String)
				}
				depNode, exists := declarationToNode[depName.String]
				if !exists {
					errors = append(errors, UndefinedVariable(depName))
				}
				node.Requires.Add(depNode)
			}
		}
		if len(errors) > 0 {
			return
		}
		evaluated.Append(evalNodes...)
	}
	return
}
