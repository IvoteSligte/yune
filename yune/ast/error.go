package ast

import "fmt"

type DuplicateDeclaration struct {
	First  Declaration
	Second Declaration
}

func (e DuplicateDeclaration) Error() string {
	return fmt.Sprintf("%s previously defined at %s redefined at %s.", e.First.GetName(), e.First.GetSpan(), e.Second.GetSpan())
}

type UndefinedVariable Variable

func (e UndefinedVariable) Error() string {
	return fmt.Sprintf("Variable %s used at %s is not defined.", e.Name, e.Span)
}

type InvalidUnaryExpressionType UnaryExpression

func (e InvalidUnaryExpressionType) Error() string {
	return fmt.Sprintf("Cannot apply unary operator %s to type %s at %s.", e.Op, e.Expression.GetExpressionType(), e.Span)
}

type InvalidBinaryExpressionTypes BinaryExpression

func (e InvalidBinaryExpressionTypes) Error() string {
	return fmt.Sprintf("Cannot apply binary operator %s to types %s and %s at %s.", e.Op, e.Left.GetExpressionType(), e.Right.GetExpressionType(), e.Span)
}
