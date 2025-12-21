package ast

import (
	"yune/cpp"

	mapset "github.com/deckarep/golang-set/v2"
)

type stageNode struct {
	// Names of nodes that must be in an earlier stage than this node.
	priors mapset.Set[string]
	// Names of nodes that must be in the same or an earlier stage than this node.
	simuls mapset.Set[string]
}

type stage map[string]stageNode

func tryMove(currentStage stage, priorStage stage, name string) {
	node, ok := currentStage[name]
	if !ok {
		// does not exist, no need to move
		return
	}
	// move node
	priorStage[name] = node
	delete(currentStage, name)

	// move all dependencies of node
	for priorName := range node.priors.Iter() {
		tryMove(currentStage, priorStage, priorName)
	}
	for simulName := range node.simuls.Iter() {
		tryMove(currentStage, priorStage, simulName)
	}
}

// Produces a topological ordering by splitting the current stage into
// multiple sub-stages as needed to satisfy the existing nodes' constraints.
func stagedOrdering(currentStage stage) []stage {
	priorStage := stage{}

	for _, node := range currentStage {
		// if any priors are in the current stage,
		// then move them to priorStage
		for priorName := range node.priors.Iter() {
			tryMove(currentStage, priorStage, priorName)
		}
	}
	if len(priorStage) == 0 {
		// no elements were moved, so the current stage is the first stage
		return []stage{currentStage}
	} else {
		// elements were moved and the prior stage still needs to be ordered
		return append(stagedOrdering(priorStage), currentStage)
	}
}

func (s stage) getPrefix(topLevel map[string]TopLevelDeclaration) (declarations []cpp.TopLevelDeclaration) {
	for _, decl := range BuiltinDeclarations {
		declarations = append(declarations, decl.(TopLevelDeclaration).Lower())
	}
	typeDeps := mapset.NewSet[string]()
	valueDeps := mapset.NewSet[string]()
	for _, deps := range s {
		typeDeps.Append(deps.priors.ToSlice()...)
		valueDeps.Append(deps.simuls.ToSlice()...)
	}
	// ensure dependencies in the current stage are not added again
	for name := range s {
		valueDeps.Remove(name)
	}
	// add type dependencies
	for typeDep := range typeDeps.Iter() {
		declarations = append(declarations, topLevel[typeDep].Lower())
	}
	// add value dependencies
	for valueDep := range valueDeps.Iter() {
		declarations = append(declarations, topLevel[valueDep].Lower())
	}
	return
}
