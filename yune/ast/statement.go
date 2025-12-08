package ast

type VariableDeclaration struct {
	Span
	Name string
	Type Type
	Body []Statement
}

// GetSpan implements Declaration.
func (d VariableDeclaration) GetSpan() Span {
	panic("unimplemented")
}

func (d VariableDeclaration) GetName() string {
	return d.Name
}

func (d VariableDeclaration) GetDeclarationType() Type {
	return d.Type
}

type Assignment struct {
	Span
	Target string
	Op     AssignmentOp
	Body   []Statement
}

type AssignmentOp string

const (
	Assign         AssignmentOp = "="
	AddAssign      AssignmentOp = "+="
	SubtractAssign AssignmentOp = "-="
	MultiplyAssign AssignmentOp = "*="
	DivideAssign   AssignmentOp = "/="
)

// Always the last statement in a list, since the remaining
// statements in a block are is in its .Else field.
type BranchStatement struct {
	Span
	Condition Expression
	Then      []Statement
	Else      []Statement
}

type Statement any
