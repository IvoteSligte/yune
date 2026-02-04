package ast

import (
	"yune/cpp"
	"yune/util"
)

type StatementBase interface {
	Node
	InferType(deps DeclarationTable) Errors
	GetType() TypeValue
	GetMacros() []*Macro

	GetMacroTypeDependencies() (deps []Query)
	GetMacroValueDependencies() (deps []Name)

	GetTypeDependencies() (deps []Query)
	GetValueDependencies() (deps []Name)
}

type Statement interface {
	StatementBase
	Lower() cpp.Statement
}

type VariableDeclaration struct {
	Span
	Name Name
	Type Type
	Body Block
}

// GetMacros implements Statement.
func (d *VariableDeclaration) GetMacros() []*Macro {
	return d.Body.GetMacros()
}

// TypeCheckBody implements Declaration.
func (d *VariableDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	panic("TypeCheckBody should not be called on VariableDeclaration (use InferType).")
}

// GetMacroTypeDependencies implements Statement.
func (d *VariableDeclaration) GetMacroTypeDependencies() (deps []Query) {
	return append(deps, d.Body.GetMacroTypeDependencies()...)
}

// GetTypeDependencies implements Statement.
func (d *VariableDeclaration) GetTypeDependencies() (deps []Query) {
	deps = append(deps, Query{
		Expression:   d.Type.Expression,
		Destination:  SetType{&d.Type.value},
		ExpectedType: TypeType{},
	})
	return append(deps, d.Body.GetTypeDependencies()...)
}

// GetMacroValueDependencies implements Statement.
func (d VariableDeclaration) GetMacroValueDependencies() []Name {
	return d.Body.GetMacroValueDependencies()
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
	if !typeEqual(declType, bodyType) {
		errors = append(errors, VariableTypeMismatch{
			Expected: declType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
		})
		return
	}
	return
}

// Lower implements Statement.
func (d VariableDeclaration) Lower() cpp.Statement {
	return cpp.VariableDeclaration{
		Name:  d.Name.String,
		Type:  d.Type.value.Lower(), // TODO: actually register the type too
		Value: d.Body.LowerVariableBody(),
	}
}

func (d VariableDeclaration) GetName() Name {
	return d.Name
}

func (d VariableDeclaration) GetType() TypeValue {
	return NilType{}
}

func (d VariableDeclaration) GetDeclaredType() TypeValue {
	return d.Type.Get()
}

type Assignment struct {
	Span
	Target Variable
	Op     AssignmentOp
	Body   Block
}

// GetMacros implements Statement.
func (a *Assignment) GetMacros() []*Macro {
	return a.Body.GetMacros()
}

// GetMacroTypeDependencies implements Statement.
func (a *Assignment) GetMacroTypeDependencies() []Query {
	return a.Body.GetMacroTypeDependencies()
}

// GetMacroValueDependencies implements Statement.
func (a *Assignment) GetMacroValueDependencies() []Name {
	return append(a.Target.GetMacroValueDependencies(), a.Body.GetMacroValueDependencies()...)
}

// GetTypeDependencies implements Statement.
func (a *Assignment) GetTypeDependencies() []Query {
	return a.Body.GetTypeDependencies()
}

// GetValueDependencies implements Statement.
func (a *Assignment) GetValueDependencies() []Name {
	return append(a.Target.GetValueDependencies(), a.Body.GetValueDependencies()...)
}

// InferType implements Statement.
func (a *Assignment) InferType(deps DeclarationTable) (errors Errors) {
	errors = append(a.Target.InferType(nil, deps), a.Body.InferType(deps.NewScope())...)
	if len(errors) > 0 {
		return
	}
	targetType := a.Target.GetType()
	bodyType := a.Body.GetType()
	if !typeEqual(targetType, bodyType) {
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

func (a Assignment) GetType() TypeValue {
	return NilType{}
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
	TypeValue
	Condition Expression
	Then      Block
	Else      Block
}

// GetMacros implements Statement.
func (b *BranchStatement) GetMacros() (macros []*Macro) {
	macros = append(b.Condition.GetMacros(), b.Then.GetMacros()...)
	macros = append(macros, b.Else.GetMacros()...)
	return
}

// GetType implements Statement.
func (b *BranchStatement) GetType() TypeValue {
	return b.TypeValue
}

// GetMacroTypeDependencies implements Statement.
func (b *BranchStatement) GetMacroTypeDependencies() (deps []Query) {
	return append(b.Then.GetMacroTypeDependencies(), b.Else.GetMacroTypeDependencies()...)
}

// GetMacroValueDependencies implements Statement.
func (b *BranchStatement) GetMacroValueDependencies() (deps []Name) {
	deps = b.Condition.GetMacroValueDependencies()
	deps = append(deps, b.Then.GetMacroValueDependencies()...)
	deps = append(deps, b.Else.GetMacroValueDependencies()...)
	return
}

// GetTypeDependencies implements Statement.
func (b *BranchStatement) GetTypeDependencies() (deps []Query) {
	return append(b.Then.GetTypeDependencies(), b.Else.GetTypeDependencies()...)
}

// GetValueDependencies implements Statement.
func (b *BranchStatement) GetValueDependencies() (deps []Name) {
	deps = b.Condition.GetValueDependencies()
	deps = append(deps, b.Then.GetValueDependencies()...)
	deps = append(deps, b.Else.GetValueDependencies()...)
	return
}

// InferType implements Statement.
func (b *BranchStatement) InferType(deps DeclarationTable) (errors Errors) {
	errors = b.Condition.InferType(BoolType{}, deps)
	errors = append(errors, b.Then.InferType(deps.NewScope())...)
	errors = append(errors, b.Else.InferType(deps.NewScope())...)
	if len(errors) > 0 {
		return
	}
	conditionType := b.Condition.GetType()
	thenType := b.Then.GetType()
	elseType := b.Else.GetType()

	if !typeEqual(conditionType, BoolType{}) {
		errors = append(errors, InvalidConditionType{
			Found: conditionType,
			At:    b.Condition.GetSpan(),
		})
	}
	if !typeEqual(thenType, elseType) {
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

func (b Block) GetType() TypeValue {
	return b.Statements[len(b.Statements)-1].GetType()
}

func (b *Block) GetMacroValueDependencies() (deps []Name) {
	locals := map[string]Declaration{}
	for _, stmt := range b.Statements {
		for _, dep := range stmt.GetMacroValueDependencies() {
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

func (b *Block) GetMacros() []*Macro {
	return util.FlatMap(b.Statements, Statement.GetMacros)
}

func (b *Block) GetMacroTypeDependencies() (deps []Query) {
	for _, stmt := range b.Statements {
		decl, ok := stmt.(Declaration)
		if ok {
			deps = append(deps, decl.GetMacroTypeDependencies()...)
		}
	}
	return
}

func (b *Block) GetTypeDependencies() (deps []Query) {
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
			if err := deps.Add(decl); err != nil {
				errors = append(errors, err)
			}
		}
	}
	return
}

func (b *Block) lowerStatements() []cpp.Statement {
	statements := util.Map(b.Statements, Statement.Lower)

	if !typeEqual(b.Statements[len(b.Statements)-1].GetType(), NilType{}) {
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
	return e.Expression.InferType(nil, deps)
}

// Lower implements Statement.
func (e *ExpressionStatement) Lower() cpp.Statement {
	return cpp.ExpressionStatement{Expression: e.Expression.Lower()}
}

var _ Statement = &VariableDeclaration{}
var _ Statement = &Assignment{}
var _ Statement = &BranchStatement{}
var _ Statement = &ExpressionStatement{}
var _ StatementBase = &Block{}
