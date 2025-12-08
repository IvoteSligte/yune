package cpp

import (
	"fmt"
)

type Variable string

type FunctionCall struct {
	Function  Expression
	Arguments []Expression
}

func (c FunctionCall) String() string {
	return fmt.Sprintf("%s(%s)", c.Function, separatedBy(c.Arguments, ", "))
}

type Tuple struct {
	Elements []Expression
}

func (t Tuple) String() string {
	return fmt.Sprintf("std::make_tuple(%s)", separatedBy(t.Elements, ", "))
}

type UnaryExpression struct {
	Op         UnaryOp
	Expression Expression
}

func (u UnaryExpression) String() string {
	return fmt.Sprintf("%s %s", u.Op, u.Expression)
}

type UnaryOp string

type BinaryExpression struct {
	Op    BinaryOp
	Left  Expression
	Right Expression
}

func (b BinaryExpression) String() string {
	return fmt.Sprintf("%s %s %s", b.Left, b.Op, b.Right)
}

type BinaryOp string

type Integer int64
type Float float64

type Expression fmt.Stringer
