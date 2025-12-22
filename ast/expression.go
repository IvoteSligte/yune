package ast

import (
	"fmt"
	"log"
	"yune/cpp"
	"yune/util"
)

type Integer struct {
	Span
	Value int64
}

// GetGlobalDependencies implements Expression.
func (i Integer) GetGlobalDependencies() (deps []string) {
	return
}

// InferType implements Expression.
func (i Integer) InferType(deps DeclarationTable) (errors Errors) {
	return
}

// Lower implements Expression.
func (i Integer) Lower() cpp.Expression {
	return cpp.Integer(i.Value)
}

// GetType implements Expression.
func (i Integer) GetType() cpp.Type {
	return IntType
}

type Float struct {
	Span
	Value float64
}

// GetGlobalDependencies implements Expression.
func (f Float) GetGlobalDependencies() (deps []string) {
	return
}

// InferType implements Expression.
func (f Float) InferType(deps DeclarationTable) (errors Errors) {
	return
}

// Lower implements Expression.
func (f Float) Lower() cpp.Expression {
	return cpp.Float(f.Value)
}

// GetType implements Expression.
func (f Float) GetType() cpp.Type {
	return FloatType
}

type Variable struct {
	cpp.Type
	Name
}

// GetType implements Expression.
func (v *Variable) GetType() cpp.Type {
	return v.Type
}

// GetGlobalDependencies implements Expression.
func (v *Variable) GetGlobalDependencies() (deps []string) {
	deps = append(deps, v.GetName())
	return
}

// InferType implements Expression.
func (v *Variable) InferType(deps DeclarationTable) (errors Errors) {
	decl, ok := deps.Get(v.GetName())
	if !ok {
		errors = append(errors, UndefinedVariable(v.Name))
		return
	}
	if decl.GetType().IsUninit() {
		log.Printf("WARN: Type queried before being calculated on declaration '%s'.", v.GetName())
	}
	v.Type = decl.GetType()
	return
}

// Lower implements Expression.
func (v *Variable) Lower() cpp.Expression {
	return cpp.Variable(v.GetName())
}

type FunctionCall struct {
	Span
	cpp.Type
	Function Expression
	Argument Expression
}

// GetType implements Expression.
func (f *FunctionCall) GetType() cpp.Type {
	return f.Type
}

// GetGlobalDependencies implements Expression.
func (f *FunctionCall) GetGlobalDependencies() []string {
	return append(f.Function.GetGlobalDependencies(), f.Argument.GetGlobalDependencies()...)
}

// InferType implements Expression.
func (f *FunctionCall) InferType(deps DeclarationTable) (errors Errors) {
	errors = append(f.Function.InferType(deps), f.Argument.InferType(deps)...)
	if len(errors) > 0 {
		return
	}
	functionType := f.Function.GetType()
	argumentType := f.Argument.GetType()
	expectedType, isFunction := functionType.GetParameterType()
	if !isFunction {
		errors = append(errors, NotAFunction{
			Found: functionType,
			At:    f.Function.GetSpan(),
		})
		return
	}
	if !argumentType.Eq(expectedType) {
		errors = append(errors, ArgumentTypeMismatch{
			Expected: expectedType,
			Found:    argumentType,
			At:       f.Argument.GetSpan(),
		})
		return
	}
	f.Type, _ = functionType.GetReturnType()
	return
}

// Lower implements Expression.
func (f *FunctionCall) Lower() cpp.Expression {
	return cpp.FunctionCall{
		Function: f.Function.Lower(),
		Arguments: []cpp.Expression{
			f.Argument.Lower(),
		},
	}
}

type Tuple struct {
	Span
	Elements []Expression
}

// GetGlobalDependencies implements Expression.
func (t *Tuple) GetGlobalDependencies() (deps []string) {
	for i := range t.Elements {
		deps = append(deps, t.Elements[i].GetGlobalDependencies()...)
	}
	return
}

// InferType implements Expression.
func (t *Tuple) InferType(deps DeclarationTable) (errors Errors) {
	for _, elem := range t.Elements {
		elem.InferType(deps)
	}
	return
}

// Lower implements Expression.
func (t *Tuple) Lower() cpp.Expression {
	return cpp.FunctionCall{
		Function:  cpp.Variable("std::make_tuple"),
		Arguments: util.Map(t.Elements, Expression.Lower),
	}
}

// GetType implements Expression.
func (t *Tuple) GetType() cpp.Type {
	return cpp.Type{
		Name:     "Tuple",
		Generics: util.Map(t.Elements, Expression.GetType),
	}
}

type Macro struct {
	// TODO: indicate macro text with a special span or just keep it as macro lines
	Span
	Language Variable
	Text     string
	// Result after evaluating the macro.
	Result Expression
}

// GetGlobalDependencies implements Expression.
func (m *Macro) GetGlobalDependencies() []string {
	return m.Language.GetGlobalDependencies()
}

// InferType implements Expression.
func (m *Macro) InferType(deps DeclarationTable) (errors Errors) {
	// NOTE: this should already evaluate the macro
	panic("unimplemented")
}

// Lower implements Expression.
func (m *Macro) Lower() cpp.Expression {
	panic("unimplemented")
}

// GetType implements Expression.
func (m *Macro) GetType() cpp.Type {
	return m.Result.GetType()
}

type UnaryExpression struct {
	Span
	cpp.Type
	Op         UnaryOp
	Expression Expression
}

// GetType implements Expression.
func (u *UnaryExpression) GetType() cpp.Type {
	return u.Type
}

// GetGlobalDependencies implements Expression.
func (u *UnaryExpression) GetGlobalDependencies() []string {
	return u.Expression.GetGlobalDependencies()
}

// InferType implements Expression.
func (u *UnaryExpression) InferType(deps DeclarationTable) (errors Errors) {
	errors = u.Expression.InferType(deps)
	if len(errors) > 0 {
		return
	}
	expressionType := u.Expression.GetType()
	switch {
	case
		expressionType.Eq(IntType),
		expressionType.Eq(FloatType):
		break
	default:
		errors = append(errors, InvalidUnaryExpressionType{
			Op:   u.Op,
			Type: expressionType,
			At:   u.Span,
		})
		return
	}
	u.Type = expressionType
	return
}

// Lower implements Expression.
func (u *UnaryExpression) Lower() cpp.Expression {
	switch u.Op {
	case Negate:
		return cpp.UnaryExpression{
			Op:         "-",
			Expression: u.Expression.Lower(),
		}
	default:
		panic(fmt.Sprintf("unexpected ast.UnaryOp: %#v", u.Op))
	}
}

type UnaryOp string

const (
	Negate UnaryOp = "-"
)

type BinaryExpression struct {
	Span
	cpp.Type
	Op    BinaryOp
	Left  Expression
	Right Expression
}

// GetType implements Expression.
func (b *BinaryExpression) GetType() cpp.Type {
	return b.Type
}

// GetGlobalDependencies implements Expression.
func (b *BinaryExpression) GetGlobalDependencies() []string {
	return append(b.Left.GetGlobalDependencies(), b.Right.GetGlobalDependencies()...)
}

// InferType implements Expression.
func (b *BinaryExpression) InferType(deps DeclarationTable) (errors Errors) {
	errors = append(b.Left.InferType(deps), b.Right.InferType(deps)...)
	if len(errors) > 0 {
		return
	}
	leftType := b.Left.GetType()
	rightType := b.Right.GetType()
	if !leftType.Eq(rightType) {
		errors = append(errors, InvalidBinaryExpressionTypes{
			Op:    b.Op,
			Left:  leftType,
			Right: rightType,
			At:    b.Span,
		})
		return
	}
	switch {
	case leftType.Eq(IntType):
	case leftType.Eq(FloatType):
	default:
		errors = append(errors, InvalidBinaryExpressionTypes{
			Op:    b.Op,
			Left:  leftType,
			Right: rightType,
			At:    b.Span,
		})
		return
	}
	switch b.Op {
	case
		Add,
		Divide,
		Multiply,
		Subtract:
		b.Type = leftType
	case
		Equal,
		Greater,
		GreaterEqual,
		Less,
		LessEqual,
		NotEqual:
		b.Type = BoolType
	default:
		panic(fmt.Sprintf("unexpected ast.BinaryOp: %#v", b.Op))
	}
	return
}

// Lower implements Expression.
func (b *BinaryExpression) Lower() cpp.Expression {
	switch b.Op {
	case
		Add,
		Divide,
		Equal,
		Greater,
		GreaterEqual,
		Less,
		LessEqual,
		Multiply,
		NotEqual,
		Subtract:
		return cpp.BinaryExpression{
			Op:    cpp.BinaryOp(b.Op),
			Left:  b.Left.Lower(),
			Right: b.Right.Lower(),
		}
	default:
		panic(fmt.Sprintf("unexpected ast.BinaryOp: %#v", b.Op))
	}
}

type BinaryOp string

const (
	Add          BinaryOp = "+"
	Subtract     BinaryOp = "-"
	Multiply     BinaryOp = "*"
	Divide       BinaryOp = "/"
	Less         BinaryOp = "<"
	Greater      BinaryOp = ">"
	Equal        BinaryOp = "=="
	NotEqual     BinaryOp = "!="
	LessEqual    BinaryOp = "<="
	GreaterEqual BinaryOp = ">="
)

type Expression interface {
	Node
	GetGlobalDependencies() []string
	InferType(deps DeclarationTable) (errors Errors)
	GetType() cpp.Type
	Lower() cpp.Expression
}

type Variables = []*Variable

var _ Expression = Integer{}
var _ Expression = Float{}
var _ Expression = &Variable{}
var _ Expression = &FunctionCall{}
var _ Expression = &Tuple{}
var _ Expression = &Macro{}
var _ Expression = &UnaryExpression{}
var _ Expression = &BinaryExpression{}
