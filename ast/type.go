package ast

import (
	"yune/cpp"
	"yune/value"
)

// TODO: macros in types

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
	return t.value.Lower()
}
