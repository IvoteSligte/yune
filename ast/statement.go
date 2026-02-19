package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

type Statement interface {
	Node
	// Lower the statement, adding the "return" prefix if `isLast` is true.
	Lower(isLast bool) cpp.Statement
	Analyze(expected TypeValue, anal Analyzer) TypeValue
}

type VariableDeclaration struct {
	Span
	Name        Name
	Type        Type
	Body        Block
	HasCaptures bool
}

// TypeCheckBody implements Declaration.
func (d *VariableDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	panic("TypeCheckBody should not be called on VariableDeclaration (use InferType).")
}

// InferType implements Statement.
func (d *VariableDeclaration) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	declType := d.Type.Analyze(anal)
	scope := anal.NewScope()
	bodyType := d.Body.Analyze(d.Type.Get(), scope)
	if !declType.Eq(bodyType) {
		anal.PushError(VariableTypeMismatch{
			Expected: declType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
		})
	}
	d.HasCaptures = len(*scope.Table.captures) > 0
	return &TupleType{}
}

// Lower implements Statement.
func (d VariableDeclaration) Lower(isLast bool) cpp.Statement {
	lowered := fmt.Sprintf(`%s %s = %s;`,
		d.Type.Lower(), // TODO: actually register the type too (if a StructType)
		d.Name.Lower(),
		cpp.LambdaBlock(d.Body.Lower(), d.HasCaptures),
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
	return &TupleType{}
}

func (d VariableDeclaration) GetDeclaredType() TypeValue {
	return d.Type.Get()
}

type Assignment struct {
	Span
	Target      Variable
	Op          AssignmentOp
	Body        Block
	HasCaptures bool
}

// Analyze implements Statement.
func (a *Assignment) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	targetType := a.Target.Analyze(nil, anal)
	scope := anal.NewScope()
	bodyType := a.Body.Analyze(targetType, scope)
	if !targetType.Eq(bodyType) {
		anal.PushError(AssignmentTypeMismatch{
			Expected: targetType,
			Found:    bodyType,
			At:       a.Body.GetSpan(),
		})
	}
	a.HasCaptures = len(*scope.Table.captures) > 0
	return &TupleType{}
}

// Lower implements Statement.
func (a *Assignment) Lower(isLast bool) cpp.Statement {
	lowered := fmt.Sprintf(`%s %s %s;`,
		a.Target.Name.String,
		a.Op,
		cpp.LambdaBlock(a.Body.Lower(), a.HasCaptures),
	)
	if isLast {
		lowered += "\nreturn std::make_tuple();"
	}
	return lowered
}

func (a Assignment) GetType() TypeValue {
	return &TupleType{}
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

// Analyze implements Statement.
func (b *BranchStatement) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	conditionType := b.Condition.Analyze(&BoolType{}, anal)
	thenType := b.Then.Analyze(expected, anal.NewScope())
	elseType := b.Else.Analyze(expected, anal.NewScope())

	if conditionType != nil && !conditionType.Eq(&BoolType{}) {
		anal.PushError(InvalidConditionType{
			Found: conditionType,
			At:    b.Condition.GetSpan(),
		})
	}
	if thenType != nil && elseType != nil && !thenType.Eq(elseType) {
		anal.PushError(BranchTypeNotEqual{
			Then:   thenType,
			ThenAt: b.Then.GetSpan(),
			Else:   elseType,
			ElseAt: b.Else.GetSpan(),
		})
	}
	return thenType // TODO: union with elseType
}

// Lower implements Statement.
func (b *BranchStatement) Lower(isLast bool) cpp.Statement {
	if !isLast {
		panic("Branch statement should always be the last statement in a block.")
	}
	lowered := fmt.Sprintf(`if (%s) %s else %s`,
		b.Condition.Lower(),
		cpp.Block(b.Then.Lower()),
		cpp.Block(b.Else.Lower()),
	)
	return lowered
}

type Block struct {
	Span       Span
	Statements []Statement
}

func (b Block) GetSpan() Span {
	return b.Span
}

func (b *Block) Analyze(expected TypeValue, anal Analyzer) (_type TypeValue) {
	for i := range b.Statements {
		// Only the last statement has a known expected type, the rest should use the default.
		expected := expected
		if i+1 < len(b.Statements) {
			expected = nil
		}
		_type = b.Statements[i].Analyze(expected, anal)
		decl, isDeclaration := b.Statements[i].(Declaration)
		if isDeclaration {
			err := anal.Table.Add(decl)
			if err != nil {
				anal.PushError(err)
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
	lowered := e.Expression.Lower()
	if isLast {
		return "return " + lowered + ";"
	} else {
		return lowered + ";"
	}
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
