package ast

import (
	"yune/cpp"
	"yune/util"
)

type VariableDeclaration struct {
	Span
	Name Name
	Type Type
	Body Block
}

// GetValueDependencies implements Statement.
func (d VariableDeclaration) GetValueDependencies(locals DeclarationTable) []string {
	return append(d.Type.GetValueDependencies(), d.Body.GetValueDependencies(locals.NewScope())...)
}

// InferType implements Statement.
func (d *VariableDeclaration) InferType(deps DeclarationTable) Errors {
	return d.Type.InferType(deps)
}

// Lower implements Statement.
func (d VariableDeclaration) Lower() cpp.Statement {
	return cpp.VariableDeclaration{
		Name:  d.Name.String,
		Type:  d.Type.Lower(),
		Value: nil, // TODO
	}
}

func (d VariableDeclaration) GetName() string {
	return d.Name.String
}

func (d VariableDeclaration) GetType() InferredType {
	return d.Type.InferredType
}

type Assignment struct {
	Span
	Target Variable
	Op     AssignmentOp
	Body   Block
}

// GetValueDependencies implements Statement.
func (a *Assignment) GetValueDependencies(locals DeclarationTable) []string {
	return append(a.Target.GetGlobalDependencies(locals), a.Body.GetValueDependencies(locals.NewScope())...)
}

// InferType implements Statement.
func (a *Assignment) InferType(deps DeclarationTable) (errors Errors) {
	errors = append(a.Target.InferType(deps), a.Body.InferType(deps.NewScope())...)
	if len(errors) > 0 {
		return
	}
	targetType := a.Target.GetType()
	bodyType := a.Body.GetType()
	if !targetType.Eq(bodyType) {
		errors = append(errors, TypeMismatch{
			Expected: targetType,
			Found:    bodyType,
			At:       a.Body.GetSpan(),
		})
		return
	}
	return
}

// Lower implements Statement.
func (a *Assignment) Lower() cpp.Statement {
	return cpp.Assignment{
		Target: a.Target.GetName(),
		Op:     cpp.AssignmentOp(a.Op),
		Value:  nil, // TODO
	}
}

func (a Assignment) GetType() InferredType {
	return NilType
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
	InferredType
	Condition Expression
	Then      Block
	Else      Block
}

// GetValueDependencies implements Statement.
func (b *BranchStatement) GetValueDependencies(locals DeclarationTable) (deps []string) {
	deps = b.Condition.GetGlobalDependencies(locals)
	deps = append(deps, b.Then.GetValueDependencies(locals.NewScope())...)
	deps = append(deps, b.Else.GetValueDependencies(locals.NewScope())...)
	return
}

// InferType implements Statement.
func (b *BranchStatement) InferType(deps DeclarationTable) (errors Errors) {
	errors = b.Condition.InferType(deps)
	errors = append(errors, b.Then.InferType(deps.NewScope())...)
	errors = append(errors, b.Else.InferType(deps.NewScope())...)
	if len(errors) > 0 {
		return
	}
	conditionType := b.Condition.GetType()
	thenType := b.Then.GetType()
	elseType := b.Else.GetType()

	if !conditionType.Eq(BoolType) {
		errors = append(errors, TypeMismatch{
			Expected: BoolType,
			Found:    conditionType,
			At:       b.Condition.GetSpan(),
		})
	}
	if !thenType.Eq(elseType) {
		errors = append(errors, BranchTypeNotEqual{
			Then:   thenType,
			ThenAt: b.Then.GetSpan(),
			Else:   elseType,
			ElseAt: b.Else.GetSpan(),
		})
	}
	return
}

// Lower implements Statement.
func (b *BranchStatement) Lower() cpp.Statement {
	return cpp.BranchStatement{
		Condition: b.Condition.Lower(),
		Then:      b.Then.Lower(),
		Else:      b.Else.Lower(),
	}
}

type Block struct {
	Span
	InferredType
	Statements []Statement
}

// GetValueDependencies implements Node.
func (b *Block) GetValueDependencies(locals DeclarationTable) (deps []string) {
	for _, stmt := range b.Statements {
		decl, ok := stmt.(Declaration)
		if ok {
			locals.Add(decl)
		}
		deps = append(deps, stmt.GetValueDependencies(locals)...)
	}
	return
}

// InferType implements Node.
func (b *Block) InferType(deps DeclarationTable) (errors Errors) {
	for i := range b.Statements {
		// FIXME: does not take into account variable declarations
		errors = append(errors, b.Statements[i].InferType(deps)...)
		if len(errors) > 0 {
			return
		}
		decl, ok := b.Statements[i].(Declaration)
		if ok {
			TODO
		}
	}
	b.InferredType = b.Statements[len(b.Statements)-1].GetType()
	return
}

func (b *Block) Lower() cpp.Block {
	// TODO: function body vs variable body
	return util.Map(b.Statements, Statement.Lower)
}

var _ Node = &Block{}

type ExpressionStatement struct {
	Expression
}

// GetValueDependencies implements Statement.
func (e *ExpressionStatement) GetValueDependencies(locals DeclarationTable) (deps []string) {
	return e.Expression.GetGlobalDependencies(locals)
}

// Lower implements Statement.
func (e *ExpressionStatement) Lower() cpp.Statement {
	return e.Expression.Lower()
}

type Types = []*Type

type Statement interface {
	Node
	GetType() InferredType
	// TODO: rename to GetGlobalDependencies
	GetValueDependencies(locals DeclarationTable) (deps []string)
	Lower() cpp.Statement
}

var _ Statement = &VariableDeclaration{}
var _ Statement = &Assignment{}
var _ Statement = &BranchStatement{}
var _ Statement = &ExpressionStatement{}
