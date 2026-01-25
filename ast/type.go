package ast

import (
	"yune/cpp"
	"yune/value"
)

var unknownType = value.Type("")

type Type struct {
	// Evaluated expression
	value      value.Type
	Expression Expression
}

// TODO: rename to GetValue
func (t Type) Get() value.Type {
	return t.value
}

func (t Type) Lower() cpp.Type {
	return cpp.Type(string(t.value))
}
