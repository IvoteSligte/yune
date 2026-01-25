package ast

import (
	"encoding/json"
	"log"
	"yune/value"

	mapset "github.com/deckarep/golang-set/v2"
)

type eval struct {
	Expression Expression
	After      mapset.Set[*eval]
	Requires   []TopLevelDeclaration
}

type stageNode struct {
	// Expression to be evaluated. May be nil.
	expression Expression
	// Destination to write the expression's evaluated value to. Required if expression is non-nil.
	// TODO: non-types
	destination *value.Type
	// The associated top-level declaration. May be nil.
	declaration TopLevelDeclaration
	// Names of nodes that must be in an earlier stage than this node.
	after mapset.Set[*stageNode]
	// Names of nodes that must be in the same or an earlier stage than this node.
	requires mapset.Set[*stageNode]
}

type stage = mapset.Set[*stageNode]

func tryMove(currentStage stage, afterStage stage, node *stageNode) {
	if !currentStage.Contains(node) {
		// does not exist, no need to move
		return
	}
	// move node
	afterStage.Add(node)
	currentStage.Remove(node)

	// move all dependencies of node
	for afterName := range node.after.Iter() {
		tryMove(currentStage, afterStage, afterName)
	}
	for simulName := range node.requires.Iter() {
		tryMove(currentStage, afterStage, simulName)
	}
}

// Produces a topological ordering by splitting the current stage into
// multiple sub-stages as needed to satisfy the existing nodes' constraints.
func stagedOrdering(currentStage stage) []stage {
	afterStage := mapset.NewSet[*stageNode]()

	for node := range currentStage.Iter() {
		// if any afters are in the current stage,
		// then move them to afterStage
		for afterName := range node.after.Iter() {
			tryMove(currentStage, afterStage, afterName)
		}
	}
	if afterStage.Cardinality() == 0 {
		// no elements were moved, so the current stage is the first stage
		return []stage{currentStage}
	} else {
		// elements were moved and the after stage still needs to be ordered
		return append(stagedOrdering(afterStage), currentStage)
	}
}

// Ensures dependencies within the stage are properly ordered to prevent C++ compiler errors.
// Consumes the stage.
func extractSortedNames(s stage) (nodes []*stageNode) {
	prevLen := s.Cardinality()
	for s.Cardinality() > 0 {
		// ToSlice() used to prevent a deadlock when calling s.Remove(node)
		// which is caused by simultaneous iteration and removal
		for _, node := range s.ToSlice() {
			if node.requires.IsSubset(mapset.NewSet(nodes...)) {
				nodes = append(nodes, node)
				s.Remove(node)
			}
		}
		if s.Cardinality() == prevLen {
			// The compiler should have errored before reaching this point.
			jsonStr, err := json.Marshal(s)
			if err != nil {
				log.Fatalln("Infinite loop in extractSortedNames and JSON serialization error:", err)
			}
			log.Fatalf("Infinite loop in extractSortedNames with remaining keys %s.", jsonStr)
		}
		prevLen = s.Cardinality()
	}
	return
}
