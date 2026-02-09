package ast

import (
	"fmt"
	"yune/cpp"
)

type Span struct {
	Line   int
	Column int
}

func (s Span) GetSpan() Span {
	return s
}

func (s Span) String() string {
	return fmt.Sprintf("%d:%d", s.Line, s.Column)
}

func (s Span) Lower() cpp.Expression {
	return fmt.Sprintf(`Span(%d, %d)`, s.Line, s.Column)
}

type Name struct {
	Span
	String string
}

// Lowers a name, renaming in case of naming conflicts with reserved identifiers.
func (n Name) Lower() string {
	switch n.String {
	case "main":
		return "main_"
	default:
		return n.String
	}
}

type Node interface {
	GetSpan() Span
}

type IName interface {
	GetName() string
	GetSpan() Span
}

type Errors = []error

type Query struct {
	Expression
	Destination
}
