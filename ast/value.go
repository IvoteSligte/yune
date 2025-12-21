package ast

import "yune/cpp"

// The result of a runtime computation or a C++ syntax tree node.
type Value interface {
	// Marks that this is a value.
	value()
}

type CppASTValue struct {
	cpp.Node
}

// value implements Value.
func (c CppASTValue) value() {
}

var _ Value = InferredType{}
var _ Value = CppASTValue{}
