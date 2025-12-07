package ast

type VariableDeclaration struct {
	Name string
	Type Type
	Body []Statement
}

func (d VariableDeclaration) GetName() string {
	return d.Name
}

func (d VariableDeclaration) GetDeclarationType() Type {
	return d.Type
}

type Assignment struct {
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
	Condition Expression
	Then      []Statement
	Else      []Statement
}

type Statement any
