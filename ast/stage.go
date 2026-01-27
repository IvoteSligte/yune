package ast

import (
	"log"
	"yune/util"

	mapset "github.com/deckarep/golang-set/v2"
)

type evalNode struct {
	// Query to be evaluated. May be empty.
	Query Query
	// The associated top-level Declaration. May be nil.
	Declaration TopLevelDeclaration
	// Names of nodes that must be in an earlier stage than this node.
	After mapset.Set[*evalNode]
	// Names of nodes that must be in the same or an earlier stage than this node.
	Requires mapset.Set[*evalNode]
	// Node that depends on the evaluation of this expression.
	// Query is expected to be non-empty if this is non-nil.
	// This is used for macros to update the declaration containing them.
	UpdateHook *evalNode
	// Precomputed nodes are not executed again.
	// Declaration should be non-nil if this is true.
	IsPrecomputed bool
}

// Converts the node to a string for debugging purposes.
func (e *evalNode) String() string {
	if e.Declaration != nil {
		return e.Declaration.GetName().String
	}
	if e.Query.Expression != nil {
		return "<expression at " + e.Query.Expression.GetSpan().String() + ">"
	}
	return "<empty>"

}

type evalSet = mapset.Set[*evalNode]

// Returns a sorted list of the 'unsorted' set.
// Clears the 'unsorted' set.
func sortedEvaluatableNodes(unsorted evalSet, evaluated evalSet) (sorted []*evalNode) {
	// TODO: allow mutual recursion for functions
	existing := evaluated.Clone()
	for unsorted.Cardinality() > 0 {
		anyChange := false
		for node := range unsorted.Iter() {
			if node.Requires.IsSubset(existing) {
				unsorted.Remove(node)
				sorted = append(sorted, node)
				existing.Add(node)
				anyChange = true
			}
		}
		if !anyChange {
			// there must be a loop with "requires" relations
			// e.g. A requires B, but B requires A, which prevents a proper ordering
			panic("Cannot sort evaluation nodes due to 'required' dependency loop") // TODO: return proper error
		}
	}
	return
}

func extractEvaluatableNodes(unevaluated evalSet, evaluated evalSet) []*evalNode {
	// determine nodes to execute based only on whether their 'after's are evaluated
	candidates := mapset.NewThreadUnsafeSet[*evalNode]()
	for node := range unevaluated.Iter() {
		if node.After.IsSubset(evaluated) {
			unevaluated.Remove(node)
			candidates.Add(node)
		}
	}
	available := evaluated.Union(candidates)
	// retain only nodes that can execute based on the availability of their 'required' nodes
	for candidates.Cardinality() > 0 {
		anyChange := false
		for node := range candidates.Iter() {
			if !node.Requires.IsSubset(available) {
				anyChange = true
				candidates.Remove(node)
				available.Remove(node)
				unevaluated.Add(node)
			}
		}
		if !anyChange {
			break // successfully removed all invalid nodes
		}
	}
	// check for errors
	if unevaluated.Cardinality() > 0 && candidates.Cardinality() == 0 {
		log.Fatalf("Dependency loop in evaluation with nodes: %v. After (unevaluated): %v. Requires (unevaluated): %v.", unevaluated.ToSlice(),
			util.Map(unevaluated.ToSlice(), func(node *evalNode) []*evalNode {
				return node.After.Difference(unevaluated).ToSlice()
			}),
			util.Map(unevaluated.ToSlice(), func(node *evalNode) []*evalNode {
				return node.Requires.Difference(available).ToSlice()
			}),
		)

		// there must be a loop with "after" relations
		// e.g. A executes after B, but B executes after A as well
		panic("'after-after' or 'after-requires' loop") // TODO: return proper error

		// // there must be a loop with "after" and "requires" relations
		// // e.g. A executes after B, but B requires A to execute
		// missing := node.Requires.Difference(available).ToSlice()
		// // TODO: return proper error
		// missingAfterUnevaluated := missing[0].After.Intersect(unevaluated).ToSlice()
		// log.Fatalf(
		// 	"'after' and 'requires' loop: %s misses required dependencies %v of which the first executes after unevaluated %v",
		// 	node, missing, missingAfterUnevaluated,
		// )
	}

	// sort nodes to execute
	return sortedEvaluatableNodes(candidates, evaluated)
}
