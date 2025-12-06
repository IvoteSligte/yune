package ast

type Module struct {
	Declarations []TopLevelDeclaration
}

type TopLevelDeclaration any

type FunctionDeclaration struct {
	Name       string
	Parameters []FunctionParameter
	ReturnType Type
	Body       []Statement
}

type FunctionParameter struct {
	Name string
	Type Type
}

type ConstantDeclaration struct {
	Name string
	Type Type
	Body []Statement
}

type Type string

type Statement any

type VariableDeclaration struct {
	ConstantDeclaration
}

type Assignment struct {
	Target string
	Op     AssignmentOp
	Body   []Statement
}

type AssignmentOp int

const (
	Assign         AssignmentOp = iota // =
	AddAssign                          // +=
	SubtractAssign                     // -=
	MultiplyAssign                     // *=
	DivideAssign                       // /=
)

type FunctionCall struct {
	Function string
	Argument Expression
}

type Tuple []Expression

type Macro struct {
	Language string
	Text     string
}

type UnaryExpression struct {
	Op         UnaryOp
	Expression Expression
}

type UnaryOp int

const (
	Negate UnaryOp = iota // -
)

type BinaryExpression struct {
	Op    BinaryOp
	Left  Expression
	Right Expression
}

type BinaryOp int

const (
	Add          BinaryOp = iota // +
	Subtract                     // -
	Multiply                     // *
	Divide                       // /
	Less                         // <
	Greater                      // >
	Equal                        // ==
	NotEqual                     // !=
	LessEqual                    // <=
	GreaterEqual                 // >=
)

type Expression any

// Always the last statement in a list, since the remaining
// statements in a block are is in its .Else field.
type BranchStatement struct {
	Condition Expression
	Then      []Statement
	Else      []Statement
}
