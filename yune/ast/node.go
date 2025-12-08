package ast

import "fmt"

type Span struct {
	Line   int
	Column int
}

func (s Span) String() string {
	return fmt.Sprintf("%s:%s", s.Line, s.Column)
}
