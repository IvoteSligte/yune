package ast

import (
	"fmt"
)

type DuplicateDeclaration struct {
	First  Declaration
	Second Declaration
}

func (e DuplicateDeclaration) Error() string {
	return fmt.Sprintf("%s previously defined at %s redefined at %s.", e.First.GetName(), e.First.GetSpan(), e.Second.GetSpan())
}

type DuplicateParameter struct {
	First  FunctionParameter
	Second FunctionParameter
}

func (e DuplicateParameter) Error() string {
	return fmt.Sprintf("%s previously defined at %s redefined at %s.", e.First.Name, e.First.Span, e.Second.Span)
}

type InvalidUnaryExpressionType struct {
	Op   UnaryOp
	Type InferredType
	At   Span
}

func (e InvalidUnaryExpressionType) Error() string {
	return fmt.Sprintf(
		"Unary operator %s cannot be applied to expression of type %s at %s.",
		e.Op,
		e.Type,
		e.At,
	)
}

type InvalidBinaryExpressionTypes struct {
	Op    BinaryOp
	Left  InferredType
	Right InferredType
	At    Span
}

func (e InvalidBinaryExpressionTypes) Error() string {
	return fmt.Sprintf(
		"Binary operator %s cannot be applied to expressions of types %s and %s at %s.",
		e.Op,
		e.Left,
		e.Right,
		e.At,
	)
}

type UndefinedVariable Variable

func (e UndefinedVariable) Error() string {
	return fmt.Sprintf("Variable %s used at %s is not defined.", Variable(e).Name.String, e.Span)
}

type NotAFunction struct {
	Found InferredType
	At    Span
}

func (e NotAFunction) Error() string {
	return fmt.Sprintf("Function call on non-function type %s at %s.", e.Found, e.At)
}

type TypeMismatch struct {
	Expected InferredType
	Found    InferredType
	At       Span
}

func (e TypeMismatch) Error() string {
	return fmt.Sprintf("Expected type %s, found type %s at %s.", e.Expected, e.Found, e.At)
}

type BranchTypeNotEqual struct {
	Then   InferredType
	ThenAt Span
	Else   InferredType
	ElseAt Span
}

func (e BranchTypeNotEqual) Error() string {
	return fmt.Sprintf(
		"Types of branches %s at %s and %s at %s are not equal.",
		e.Then,
		e.ThenAt,
		e.Else,
		e.ElseAt,
	)
}
