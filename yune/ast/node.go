package ast

import "fmt"

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

type Name struct {
	Span
	String string
}

func (n Name) GetName() string {
	return n.String
}

type Node interface {
	GetSpan() Span
	InferType(deps DeclarationTable) Errors
	GetType() InferredType
}

type IName interface {
	GetName() string
	GetSpan() Span
}

type Query = Name
type Queries = []Query
type Errors = []error
