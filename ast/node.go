package ast

import (
	"fmt"
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

type Name struct {
	Span
	String string
}

type Node interface {
	GetSpan() Span
}

type IName interface {
	GetName() string
	GetSpan() Span
}

type Query = Name
type Queries = []Query
type Errors = []error
