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

type Statement interface {
	Node
	// Lower the statement, adding the "return" prefix if `isLast` is true.
	Lower(isLast bool) cpp.Statement
	Analyze(expected TypeValue, anal Analyzer) TypeValue
}

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

// InferType implements Statement.
func (d *VariableDeclaration) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	bodyType := d.Body.Analyze(d.Type.Get(), anal)
	declType := d.Type.Analyze(anal)
	if bodyType != nil && declType != nil && !declType.Eq(bodyType) {
		anal.PushError(VariableTypeMismatch{
			Expected: declType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
		})
	}
	return TupleType{}
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

// Analyze implements Statement.
func (a *Assignment) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	targetType := a.Target.Analyze(nil, anal)
	bodyType := a.Body.Analyze(targetType, anal.NewScope())
	if targetType != nil && bodyType != nil && !targetType.Eq(bodyType) {
		anal.PushError(AssignmentTypeMismatch{
			Expected: targetType,
			Found:    bodyType,
			At:       a.Body.GetSpan(),
		})
	}
	return TupleType{}
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

// Analyze implements Statement.
func (b *BranchStatement) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	conditionType := b.Condition.Analyze(BoolType{}, anal)
	thenType := b.Then.Analyze(expected, anal.NewScope())
	elseType := b.Else.Analyze(expected, anal.NewScope())

	if conditionType != nil && !conditionType.Eq(BoolType{}) {
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

func (b *Block) Analyze(expected TypeValue, anal Analyzer) (_type TypeValue) {
	scope := anal.GetScope()
	for i := range b.Statements {
		// Only the last statement has a known expected type, the rest should use the default.
		expected := expected
		if i+1 < len(b.Statements) {
			expected = nil
		}
		_type = b.Statements[i].Analyze(expected, anal)
		decl, ok := b.Statements[i].(Declaration)
		if ok {
			name := decl.GetName().String
			_, exists := scope[name]
			if exists {
				anal.PushError(DuplicateDeclaration{
					First:  scope[name],
					Second: decl,
				})
			}
			scope[name] = decl
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
