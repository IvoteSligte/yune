package cpp

import (
	"fmt"
	"yune/util"
)

type VariableDeclaration struct {
	Name  string
	Type  Type
	Value Expression
}

// String implements Statement.
func (v VariableDeclaration) String() string {
	return fmt.Sprintf("%s %s = %s;", v.Type, v.Name, v.Value)
}

type Assignment struct {
	Target string
	Op     AssignmentOp
	Value  Expression
}

func (a Assignment) String() string {
	return fmt.Sprintf("%s %s = %s;", a.Target, a.Op, a.Value)
}

type AssignmentOp string

// Always the last statement in a list, since the remaining
// statements in a block are is in its .Else field.
type BranchStatement struct {
	Condition Expression
	Then      Block
	Else      Block
}

func (b BranchStatement) String() string {
	return fmt.Sprintf("if (%s) %s else %s", b.Condition, b.Then, b.Else)
}

type ReturnStatement struct {
	Expression Expression
}

func (r ReturnStatement) String() string {
	return fmt.Sprintf("return %s;", r.Expression)
}

type Block []Statement

func (b Block) String() string {
	return "{\n" + util.Join(b, "\n") + "\n}"
}

type ExpressionStatement struct {
	Expression
}

func (e ExpressionStatement) String() string {
	return e.Expression.String() + ";"
}

type Statement fmt.Stringer
