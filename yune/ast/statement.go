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

// TypeCheckBody implements Declaration.
func (d *VariableDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	panic("TypeCheckBody should not be called on VariableDeclaration.")
}

// GetTypeDependencies implements Statement.
func (d *VariableDeclaration) GetTypeDependencies() []string {
	return append(d.Type.GetValueDependencies(), d.Body.GetTypeDependencies()...)
}

// GetValueDependencies implements Statement.
func (d VariableDeclaration) GetValueDependencies() []string {
	return append(d.Type.GetValueDependencies(), d.Body.GetValueDependencies()...)
}

// InferType implements Statement.
func (d *VariableDeclaration) InferType(deps DeclarationTable) (errors Errors) {
	d.Type.CalcType(deps)
	return
}

func (d *VariableDeclaration) CalcType(deps DeclarationTable) {
	d.Type.CalcType(deps)
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

// GetTypeDependencies implements Statement.
func (a *Assignment) GetTypeDependencies() []string {
	return a.Body.GetTypeDependencies()
}

// GetValueDependencies implements Statement.
func (a *Assignment) GetValueDependencies() []string {
	return append(a.Target.GetGlobalDependencies(), a.Body.GetValueDependencies()...)
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
		Value:  nil, // TODO: body -> expression using immediately invoked lambda
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

// GetTypeDependencies implements Statement.
func (b *BranchStatement) GetTypeDependencies() (deps []string) {
	return append(b.Then.GetTypeDependencies(), b.Else.GetTypeDependencies()...)
}

// GetValueDependencies implements Statement.
func (b *BranchStatement) GetValueDependencies() (deps []string) {
	deps = b.Condition.GetGlobalDependencies()
	deps = append(deps, b.Then.GetValueDependencies()...)
	deps = append(deps, b.Else.GetValueDependencies()...)
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

func (b *Block) GetValueDependencies() (deps []string) {
	locals := map[string]Declaration{}
	for _, stmt := range b.Statements {
		for _, dep := range stmt.GetValueDependencies() {
			_, ok := locals[dep]
			if !ok {
				deps = append(deps, dep)
			}
		}
		// register local after getting dependencies to prevent cyclic definitions
		decl, ok := stmt.(Declaration)
		if ok {
			locals[decl.GetName()] = decl
		}
	}
	return
}

func (b *Block) GetTypeDependencies() (deps []string) {
	for _, stmt := range b.Statements {
		decl, ok := stmt.(Declaration)
		if ok {
			deps = append(deps, decl.GetTypeDependencies()...)
		}
	}
	return
}

func (b *Block) InferType(deps DeclarationTable) (errors Errors) {
	for i := range b.Statements {
		errors = append(errors, b.Statements[i].InferType(deps)...)
		if len(errors) > 0 {
			return
		}
		decl, ok := b.Statements[i].(Declaration)
		if ok {
			deps.Add(decl)
		}
		b.Statements[i].InferType(deps)
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

// GetTypeDependencies implements Statement.
func (e *ExpressionStatement) GetTypeDependencies() (deps []string) {
	return
}

// GetValueDependencies implements Statement.
func (e *ExpressionStatement) GetValueDependencies() (deps []string) {
	return e.Expression.GetGlobalDependencies()
}

// Lower implements Statement.
func (e *ExpressionStatement) Lower() cpp.Statement {
	return e.Expression.Lower()
}

type Types = []*Type

type Statement interface {
	Node
	InferType(deps DeclarationTable) Errors
	GetType() InferredType
	GetTypeDependencies() (deps []string)
	GetValueDependencies() (deps []string)
	Lower() cpp.Statement
}

var _ Statement = &VariableDeclaration{}
var _ Statement = &Assignment{}
var _ Statement = &BranchStatement{}
var _ Statement = &ExpressionStatement{}
