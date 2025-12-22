package cpp

import (
	"fmt"
	"yune/util"
)

type Variable string

func (v Variable) String() string {
	return string(v)
}

type FunctionCall struct {
	Function  Expression
	Arguments []Expression
}

func (c FunctionCall) String() string {
	return fmt.Sprintf("%s(%s)", c.Function, util.Join(c.Arguments, ", "))
}

type Tuple struct {
	Elements []Expression
}

func (t Tuple) String() string {
	return fmt.Sprintf("std::make_tuple(%s)", util.Join(t.Elements, ", "))
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

type LambdaBlock []Statement

func (b LambdaBlock) String() string {
	return "[](){" + util.Join(b, "") + "}()"
}

type Integer int64

// String implements Expression.
func (i Integer) String() string {
	return fmt.Sprint(int64(i))
}

type Float float64

// String implements Expression.
func (f Float) String() string {
	return fmt.Sprint(float64(f))
}

type Expression fmt.Stringer
