package ast

import (
	"log"
	"maps"

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

// FIXME: needs forward function declarations for recursion between two functions

// Ensures dependencies within the stage are properly ordered to prevent C++ compiler errors.
// Consumes the stage.
func (s stage) extractSortedNames() (names []string) {
	prevLen := len(s)
	for len(s) > 0 {
		for name, node := range s {
			if node.simuls.IsSubset(mapset.NewSet(names...)) {
				names = append(names, name)
				delete(s, name)
			}
		}
		if len(s) == prevLen {
			log.Fatalln("Infinite loop in extractSortedNames with remaining keys", maps.Keys(s))
		}
		prevLen = len(s)
	}
	return
}
