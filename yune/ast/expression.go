package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"
)

type Integer struct {
	Span
	Value int64
}

// GetGlobalDependencies implements Expression.
func (i Integer) GetGlobalDependencies(locals DeclarationTable) (deps []string) {
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
func (i Integer) GetType() InferredType {
	return IntType
}

type Float struct {
	Span
	Value float64
}

// GetGlobalDependencies implements Expression.
func (f Float) GetGlobalDependencies(locals DeclarationTable) (deps []string) {
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
func (f Float) GetType() InferredType {
	return FloatType
}

type Variable struct {
	InferredType
	Name
}

// GetGlobalDependencies implements Expression.
func (v *Variable) GetGlobalDependencies(locals DeclarationTable) (deps []string) {
	_, ok := locals.Get(v.GetName())
	if !ok {
		deps = append(deps, v.GetName())
	}
	return
}

// InferType implements Expression.
func (v *Variable) InferType(deps DeclarationTable) (errors Errors) {
	decl, ok := deps.Get(v.GetName())
	if !ok {
		errors = append(errors, UndefinedVariable(v.Name))
		return
	}
	v.InferredType = decl.GetType()
	return
}

// Lower implements Expression.
func (v *Variable) Lower() cpp.Expression {
	panic("unimplemented")
}

type FunctionCall struct {
	Span
	InferredType
	Function Expression
	Argument Expression
}

// GetGlobalDependencies implements Expression.
func (f *FunctionCall) GetGlobalDependencies(locals DeclarationTable) []string {
	return append(f.Function.GetGlobalDependencies(locals), f.Argument.GetGlobalDependencies(locals)...)
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
		errors = append(errors, TypeMismatch{
			Expected: expectedType,
			Found:    argumentType,
			At:       f.Argument.GetSpan(),
		})
		return
	}
	f.InferredType, _ = functionType.GetReturnType()
	return
}

// Lower implements Expression.
func (f *FunctionCall) Lower() cpp.Expression {
	panic("unimplemented")
}

type Tuple struct {
	Span
	Elements []Expression
}

// GetGlobalDependencies implements Expression.
func (t *Tuple) GetGlobalDependencies(locals DeclarationTable) (deps []string) {
	for i := range t.Elements {
		deps = append(deps, t.Elements[i].GetGlobalDependencies(locals)...)
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
	panic("unimplemented")
}

// GetType implements Expression.
func (t *Tuple) GetType() InferredType {
	return InferredType{
		name:     "Tuple",
		generics: util.Map(t.Elements, Expression.GetType),
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
func (m *Macro) GetGlobalDependencies(locals DeclarationTable) []string {
	return m.Language.GetGlobalDependencies(locals)
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
func (m *Macro) GetType() InferredType {
	return m.Result.GetType()
}

type UnaryExpression struct {
	Span
	InferredType
	Op         UnaryOp
	Expression Expression
}

// GetGlobalDependencies implements Expression.
func (u *UnaryExpression) GetGlobalDependencies(locals DeclarationTable) []string {
	return u.Expression.GetGlobalDependencies(locals)
}

// InferType implements Expression.
func (u *UnaryExpression) InferType(deps DeclarationTable) (errors Errors) {
	errors = u.Expression.InferType(deps)
	if len(errors) > 0 {
		return
	}
	expressionType := u.Expression.GetType()
	switch {
	case expressionType.Eq(IntType):
	case expressionType.Eq(FloatType):
		return
	default:
		errors = append(errors, InvalidUnaryExpressionType{
			Op:   u.Op,
			Type: expressionType,
			At:   u.Span,
		})
		return
	}
	u.InferredType = expressionType
	return
}

// Lower implements Expression.
func (u *UnaryExpression) Lower() cpp.Expression {
	panic("unimplemented")
}

type UnaryOp string

const (
	Negate UnaryOp = "-"
)

type BinaryExpression struct {
	Span
	InferredType
	Op    BinaryOp
	Left  Expression
	Right Expression
}

// GetGlobalDependencies implements Expression.
func (b *BinaryExpression) GetGlobalDependencies(locals DeclarationTable) []string {
	return append(b.Left.GetGlobalDependencies(locals), b.Right.GetGlobalDependencies(locals)...)
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
	case Add:
	case Divide:
	case Equal:
	case Greater:
	case GreaterEqual:
	case Less:
	case LessEqual:
	case Multiply:
	case NotEqual:
	case Subtract:
		break
	default:
		panic(fmt.Sprintf("unexpected ast.BinaryOp: %#v", b.Op))
	}
	b.InferredType = leftType
	return
}

// Lower implements Expression.
func (b *BinaryExpression) Lower() cpp.Expression {
	panic("unimplemented")
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
	GetGlobalDependencies(locals DeclarationTable) []string
	GetType() InferredType
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
