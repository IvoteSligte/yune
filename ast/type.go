package ast

import (
	"yune/cpp"
	"yune/pb"
)

// TODO: macros in types

type Type struct {
	// Evaluated expression
	value      pb.Type
	Expression Expression
}

// TODO: rename to GetValue
func (t Type) Get() pb.Type {
	return t.value
}

func (t Type) Lower() cpp.Type {
	return pb.LowerType(t.value)
}
