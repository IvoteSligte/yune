package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

func defString(defs []cpp.Definition) string {
	s := ""
	for _, def := range defs {
		s += def + "\n"
	}
	return s
}

type StatementBase interface {
	Node

	GetType() TypeValue
	GetMacros() []*Macro

	// Get*Dependencies, but retrieves the dependencies added by evaluated macros.
	GetMacroTypeDependencies() (deps []Query)
	GetMacroValueDependencies() (deps []Name)

	GetTypeDependencies() (deps []Query)
	GetValueDependencies() (deps []Name)

	// Infers the type, returning errors in case of mismatches.
	// GetType() should return a non-nil result if this returns no errors.
	InferType(deps DeclarationTable) Errors
}

type Statement interface {
	StatementBase
	// Lower the statement, adding the "return" prefix if `isLast` is true.
	Lower(isLast bool) cpp.Statement
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
	deps = append(deps, NewTypeQuery(&d.Type))
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
func (d VariableDeclaration) Lower(isLast bool) cpp.Statement {
	lowered := fmt.Sprintf(`%s %s = %s;`,
		d.Type.value.Lower(), // TODO: actually register the type too (if a StructType)
		d.Name.Lower(),
		cpp.LambdaBlock(d.Body.Lower()),
	)
	if isLast {
		lowered += "\nreturn std::make_tuple();"
	}
	return lowered
}

func (d VariableDeclaration) GetName() Name {
	return d.Name
}

func (d VariableDeclaration) GetType() TypeValue {
	return TupleType{}
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
	errors = append(a.Target.InferType(deps), a.Body.InferType(deps.NewScope())...)
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
func (a *Assignment) Lower(isLast bool) cpp.Statement {
	lowered := fmt.Sprintf(`%s %s %s;`,
		a.Target.Name.String,
		a.Op,
		cpp.LambdaBlock(a.Body.Lower()),
	)
	if isLast {
		lowered += "\nreturn std::make_tuple();"
	}
	return lowered
}

func (a Assignment) GetType() TypeValue {
	return TupleType{}
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
	Type      TypeValue
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
	return b.Type
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
	b.Condition.SetId(BoolType{})
	errors = b.Condition.InferType(deps)
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
	b.Type = thenType
	return
}

// Lower implements Statement.
func (b *BranchStatement) Lower(isLast bool) cpp.Statement {
	if !isLast {
		panic("Branch statement should always be the last statement in a block.")
	}
	var defs []cpp.Definition
	lowered := fmt.Sprintf(`if (%s) %s else %s`,
		b.Condition.Lower(&defs),
		cpp.Block(b.Then.Lower()),
		cpp.Block(b.Else.Lower()),
	)
	return defString(defs) + lowered
}

type Block struct {
	Span       Span
	Statements []Statement
}

func (b Block) GetSpan() Span {
	return b.Span
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
		deps = append(deps, stmt.GetTypeDependencies()...)
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

func (b *Block) Lower() (statements []cpp.Statement) {
	for i, stmt := range b.Statements {
		isLast := i+1 == len(b.Statements)
		statements = append(statements, stmt.Lower(isLast))
	}
	return
}

var _ Node = &Block{}

type ExpressionStatement struct {
	Expression
}

// Lower implements Statement.
func (e *ExpressionStatement) Lower(isLast bool) cpp.Statement {
	var defs []cpp.Definition
	lowered := e.Expression.Lower(&defs)
	if isLast {
		lowered = "return " + lowered
	}
	return defString(defs) + lowered + ";"
}

func UnmarshalBlock(data *fj.Value) (block Block) {
	return Block{
		Span:       Span{},
		Statements: util.Map(data.GetArray(), UnmarshalStatement),
	}
}

func UnmarshalStatement(data *fj.Value) (stmt Statement) {
	object := data.GetObject()
	key, v := fjUnmarshalUnion(object)
	switch key {
	case "VariableDeclaration":
		stmt = &VariableDeclaration{
			Span: fjUnmarshal(v.Get("span"), Span{}),
			Name: Name{
				Span:   Span{},
				String: string(v.GetStringBytes("name")),
			},
			Body: UnmarshalBlock(v),
		}
	case "Assignment":
		stmt = &Assignment{
			Span:   fjUnmarshal(v.Get("span"), Span{}),
			Target: *UnmarshalExpression(v.Get("target")).(*Variable),
			Op:     AssignmentOp(v.GetStringBytes("op")),
			Body:   UnmarshalBlock(v),
		}
	case "BranchStatement":
		stmt = &BranchStatement{
			Span:      fjUnmarshal(v.Get("span"), Span{}),
			Condition: UnmarshalExpression(v.Get("condition")),
			Then:      UnmarshalBlock(v.Get("then")),
			Else:      UnmarshalBlock(v.Get("else")),
		}
	default:
		stmt = &ExpressionStatement{Expression: UnmarshalExpression(data)}
	}
	return
}

var _ Statement = &VariableDeclaration{}
var _ Statement = &Assignment{}
var _ Statement = &BranchStatement{}
var _ Statement = &ExpressionStatement{}
var _ StatementBase = &Block{}
