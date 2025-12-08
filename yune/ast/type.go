package ast

type Type struct {
	Span
	Name     string
	Generics []Type
}
