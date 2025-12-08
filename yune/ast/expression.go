package ast

import (
	"yune/util"
)

type Variable struct {
	Span
	ExpressionType
	Name string
}

// InferExpressionType implements Expression.
func (e *Variable) InferExpressionType(env Environment) (_type Type, err error) {
	declaration, ok := env.Get(e.Name)
	if !ok {
		err = UndefinedVariable(*e)
		return
	}
	_type = declaration.GetDeclarationType()
	e.ExpressionType = ExpressionType(_type)
	return
}

type FunctionCall struct {
	Span
	ExpressionType
	Function Expression
	Argument Expression
}

// InferExpressionType implements Expression.
func (f *FunctionCall) InferExpressionType(env Environment) (_type Type, err error) {
	_type, err = f.Function.InferExpressionType(env)
	f.ExpressionType = ExpressionType(_type)
	return
}

type Tuple struct {
	Span
	ExpressionType
	Elements []Expression
}

// InferExpressionType implements Expression.
func (t *Tuple) InferExpressionType(env Environment) (_type Type, err error) {
	for _, element := range t.Elements {
		_, err = element.InferExpressionType(env)
		if err != nil {
			return
		}
	}
	_type = Type{
		Name:     "Tuple",
		Generics: util.Map(t.Elements, Expression.GetExpressionType),
	}
	t.ExpressionType = ExpressionType(_type)
	return
}

type Macro struct {
	// TODO: indicate macro text with a special span
	Span
	ExpressionType
	Language string
	Text     string
}

// InferExpressionType implements Expression.
func (m Macro) InferExpressionType(env Environment) (_type Type, err error) {
	panic("Macros are not supported yet.")
}

type UnaryExpression struct {
	Span
	ExpressionType
	Op         UnaryOp
	Expression Expression
}

// InferExpressionType implements Expression.
func (u UnaryExpression) InferExpressionType(env Environment) (_type Type, err error) {
	_type, err = u.Expression.InferExpressionType(env)
	if err != nil {
		return
	}
	return
}

type UnaryOp string

const (
	Negate UnaryOp = "-"
)

type BinaryExpression struct {
	Span
	ExpressionType
	Op    BinaryOp
	Left  Expression
	Right Expression
}

// InferExpressionType implements Expression.
func (b BinaryExpression) InferExpressionType(env Environment) (_type Type, err error) {
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

type Integer struct {
	Span
	Value int64
}

// InferExpressionType implements Expression.
func (i Integer) InferExpressionType(env Environment) (_type Type, err error) {
	return i.GetExpressionType(), nil
}

// GetExpressionType implements Expression.
func (Integer) GetExpressionType() Type {
	return Type{Name: "Int"}
}

type Float struct {
	Span
	Value float64
}

// InferExpressionType implements Expression.
func (f Float) InferExpressionType(env Environment) (_type Type, err error) {
	return f.GetExpressionType(), nil
}

// GetExpressionType implements Expression.
func (Float) GetExpressionType() Type {
	return Type{Name: "F64"}
}

type Expression interface {
	InferExpressionType(env Environment) (_type Type, err error)
	GetExpressionType() Type
}

type ExpressionType Type

func (t ExpressionType) GetExpressionType() Type {
	return Type(t)
}

var _ Expression = Integer{}
var _ Expression = Float{}
var _ Expression = &Variable{}
var _ Expression = &FunctionCall{}
var _ Expression = &Tuple{}
var _ Expression = &Macro{}
var _ Expression = &UnaryExpression{}
var _ Expression = &BinaryExpression{}
