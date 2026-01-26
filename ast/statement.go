package ast

import (
	"yune/cpp"
	"yune/util"
	"yune/value"
)

type VariableDeclaration struct {
	Span
	Name Name
	Type Type
	Body Block
}

// TypeCheckBody implements Declaration.
func (d *VariableDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	panic("TypeCheckBody should not be called on VariableDeclaration (use InferType).")
}

// GetTypeDependencies implements Statement.
func (d *VariableDeclaration) GetTypeDependencies() []*Type {
	return append([]*Type{&d.Type}, d.Body.GetTypeDependencies()...)
}

// GetValueDependencies implements Statement.
func (d VariableDeclaration) GetValueDependencies() []Name {
	return d.Body.GetValueDependencies()
}

// InferType implements Statement.
func (d *VariableDeclaration) InferType(deps DeclarationTable) (errors Errors) {
	errors = d.Body.InferType(deps)
	if len(errors) > 0 {
		return
	}
	declType := d.Type.Get()
	bodyType := d.Body.GetType()
	if !declType.Eq(bodyType) {
		errors = append(errors, VariableTypeMismatch{
			Expected: declType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
		})
		return
	}
	return
}

func (d *VariableDeclaration) CalcType(deps DeclarationTable) Errors {
	panic("CalcType should not be called on VariableDeclaration (use InferType).")
}

// Lower implements Statement.
func (d VariableDeclaration) Lower() cpp.Statement {
	return cpp.VariableDeclaration{
		Name:  d.Name.String,
		Type:  d.Type.Lower(), // TODO: actually register the type too
		Value: d.Body.LowerVariableBody(),
	}
}

func (d VariableDeclaration) GetName() Name {
	return d.Name
}

func (d VariableDeclaration) GetType() value.Type {
	return NilType
}

type Assignment struct {
	Span
	Target Variable
	Op     AssignmentOp
	Body   Block
}

// GetTypeDependencies implements Statement.
func (a *Assignment) GetTypeDependencies() []*Type {
	return a.Body.GetTypeDependencies()
}

// GetValueDependencies implements Statement.
func (a *Assignment) GetValueDependencies() []Name {
	return append(a.Target.GetGlobalDependencies(), a.Body.GetValueDependencies()...)
}

// InferType implements Statement.
func (a *Assignment) InferType(deps DeclarationTable) (errors Errors) {
	errors = append(a.Target.InferType(unknownType, deps), a.Body.InferType(deps.NewScope())...)
	if len(errors) > 0 {
		return
	}
	targetType := a.Target.GetType()
	bodyType := a.Body.GetType()
	if !targetType.Eq(bodyType) {
		errors = append(errors, AssignmentTypeMismatch{
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
		Target: a.Target.Name.String,
		Op:     cpp.AssignmentOp(a.Op),
		Value:  a.Body.LowerVariableBody(),
	}
}

func (a Assignment) GetType() value.Type {
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
	value.Type
	Condition Expression
	Then      Block
	Else      Block
}

// GetType implements Statement.
func (b *BranchStatement) GetType() value.Type {
	return b.Type
}

// GetTypeDependencies implements Statement.
func (b *BranchStatement) GetTypeDependencies() (deps []*Type) {
	return append(b.Then.GetTypeDependencies(), b.Else.GetTypeDependencies()...)
}

// GetValueDependencies implements Statement.
func (b *BranchStatement) GetValueDependencies() (deps []Name) {
	deps = b.Condition.GetGlobalDependencies()
	deps = append(deps, b.Then.GetValueDependencies()...)
	deps = append(deps, b.Else.GetValueDependencies()...)
	return
}

// InferType implements Statement.
func (b *BranchStatement) InferType(deps DeclarationTable) (errors Errors) {
	errors = b.Condition.InferType(BoolType, deps)
	errors = append(errors, b.Then.InferType(deps.NewScope())...)
	errors = append(errors, b.Else.InferType(deps.NewScope())...)
	if len(errors) > 0 {
		return
	}
	conditionType := b.Condition.GetType()
	thenType := b.Then.GetType()
	elseType := b.Else.GetType()

	if !conditionType.Eq(BoolType) {
		errors = append(errors, InvalidConditionType{
			Found: conditionType,
			At:    b.Condition.GetSpan(),
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
		Then:      b.Then.LowerFunctionBody(),
		Else:      b.Else.LowerFunctionBody(), // FIXME: branches generate duplicate, unreachable code (not sure where the bug is)
	}
}

type Block struct {
	Span
	Statements []Statement
}

func (b Block) GetType() value.Type {
	return b.Statements[len(b.Statements)-1].GetType()
}

func (b *Block) GetValueDependencies() (deps []Name) {
	locals := map[string]Declaration{}
	for _, stmt := range b.Statements {
		for _, dep := range stmt.GetValueDependencies() {
			_, ok := locals[dep.String]
			if !ok {
				deps = append(deps, dep)
			}
		}
		// register local after getting dependencies to prevent cyclic definitions
		decl, ok := stmt.(Declaration)
		if ok {
			locals[decl.GetName().String] = decl
		}
	}
	return
}

func (b *Block) GetTypeDependencies() (deps []*Type) {
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
			err := deps.Add(decl)
			if err != nil {
				errors = append(errors, err)
			}
		}
		b.Statements[i].InferType(deps)
	}
	return
}

func (b *Block) lowerStatements() []cpp.Statement {
	statements := util.Map(b.Statements, Statement.Lower)

	if !b.Statements[len(b.Statements)-1].GetType().Eq(NilType) {
		// last expression is implicitly returned in Yune,
		// but needs to be explicitly returned in C++
		statements[len(statements)-1] = cpp.ReturnStatement{
			Expression: statements[len(statements)-1].(cpp.ExpressionStatement).Expression,
		}
	}
	return statements
}

func (b *Block) LowerFunctionBody() cpp.Block {
	return b.lowerStatements()
}

func (b *Block) LowerVariableBody() cpp.LambdaBlock {
	return b.lowerStatements()
}

var _ Node = &Block{}

type ExpressionStatement struct {
	Expression
}

// InferType implements Statement.
// Subtle: this method shadows the method (Expression).InferType of ExpressionStatement.Expression.
func (e *ExpressionStatement) InferType(deps DeclarationTable) []error {
	return e.Expression.InferType(unknownType, deps)
}

// GetTypeDependencies implements Statement.
func (e *ExpressionStatement) GetTypeDependencies() (deps []*Type) {
	return
}

// GetValueDependencies implements Statement.
func (e *ExpressionStatement) GetValueDependencies() (deps []Name) {
	return e.Expression.GetGlobalDependencies()
}

// Lower implements Statement.
func (e *ExpressionStatement) Lower() cpp.Statement {
	return cpp.ExpressionStatement{Expression: e.Expression.Lower()}
}

type Types = []*Type

type Statement interface {
	Node
	InferType(deps DeclarationTable) Errors
	GetType() value.Type
	GetTypeDependencies() (deps []*Type)
	GetValueDependencies() (deps []Name)
	Lower() cpp.Statement
}

var _ Statement = &VariableDeclaration{}
var _ Statement = &Assignment{}
var _ Statement = &BranchStatement{}
var _ Statement = &ExpressionStatement{}
