package ast

import (
	"encoding/json"
	"log"

	mapset "github.com/deckarep/golang-set/v2"
)

type stageNode struct {
	// Query to be evaluated. May be empty.
	Query Query
	// The associated top-level Declaration. May be nil.
	Declaration TopLevelDeclaration
	// Names of nodes that must be in an earlier stage than this node.
	After mapset.Set[*stageNode]
	// Names of nodes that must be in the same or an earlier stage than this node.
	Requires mapset.Set[*stageNode]
	// Forces this node to execute in the earliest stage.
	ExecuteFirst bool
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
	for after := range node.After.Iter() {
		tryMove(currentStage, afterStage, after)
	}
	for required := range node.Requires.Iter() {
		tryMove(currentStage, afterStage, required)
	}
}

// Produces a topological ordering by splitting the current stage into
// multiple sub-stages as needed to satisfy the existing nodes' constraints.
//
// Expects currentStage to be a mapset thread unsafe set (i.e. map[*stageNode]struct{}).
func stagedOrdering(currentStage stage) []stage {
	firstStage := mapset.NewThreadUnsafeSet[*stageNode]()

	for node := range currentStage.Iter() {
		if node.ExecuteFirst {
			firstStage.Add(node)
			currentStage.Remove(node)
		}
	}
	return append([]stage{firstStage}, partialStagedOrdering(currentStage)...)
}

// The recursive part of stagedOrdering
func partialStagedOrdering(currentStage stage) []stage {
	// thread unsafe set used so that iteration and removal can be done simultaneously
	// (deadlocks otherwise)
	afterStage := mapset.NewThreadUnsafeSet[*stageNode]()

	for node := range currentStage.Iter() {
		// if any 'after's are in the current stage, then move them to afterStage
		for after := range node.After.Iter() {
			tryMove(currentStage, afterStage, after)
		}
	}
	if afterStage.Cardinality() == 0 {
		// no elements were moved, so the current stage is the first stage
		return []stage{currentStage}
	}
	// elements were moved and the after stage still needs to be ordered
	return append(partialStagedOrdering(afterStage), currentStage)
}

// Ensures dependencies within the stage are properly ordered to prevent C++ compiler errors.
// Consumes the stage.
// Requires that the set arguments are a mapset thread unsafe sets.
func extractSortedNames(s stage, evaluated mapset.Set[*stageNode]) (nodes []*stageNode) {
	prevLen := s.Cardinality()
	for s.Cardinality() > 0 {
		for node := range s.Iter() {
			existing := mapset.NewThreadUnsafeSet(nodes...).Union(evaluated)
			if node.Requires.IsSubset(existing) {
				nodes = append(nodes, node)
				s.Remove(node)
			}
		}
		if s.Cardinality() == prevLen {
			// The compiler should have errored before reaching this point.
			jsonStr, err := json.MarshalIndent(s, "", "    ")
			if err != nil {
				log.Fatalln("Infinite loop in extractSortedNames and JSON serialization error:", err)
			}
			log.Fatalf("Infinite loop in extractSortedNames with remaining keys %s.", jsonStr)
		}
		prevLen = s.Cardinality()
	}
	return
}
