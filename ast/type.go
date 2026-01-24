package ast

import (
	"encoding/hex"
	"yune/cpp"
	"yune/value"
)

type Type struct {
	// Evaluated expression
	value      value.Type
	expression Expression
}

// TODO: rename to GetValue
func (t Type) Get() value.Type {
	return t.value
}

func (t Type) Lower() cpp.Type {
	return cpp.Type(hex.EncodeToString([]byte(string(t.value))))
}
