package ast

import (
	"fmt"
	"yune/cpp"
)

type DuplicateDeclaration struct {
	First  Declaration
	Second Declaration
}

func (e DuplicateDeclaration) Error() string {
	return fmt.Sprintf("'%s' previously defined at '%s' redefined at %s.", e.First.GetName(), e.First.GetSpan(), e.Second.GetSpan())
}

type InvalidUnaryExpressionType struct {
	Op   UnaryOp
	Type cpp.Type
	At   Span
}

func (e InvalidUnaryExpressionType) Error() string {
	return fmt.Sprintf(
		"Unary operator %s cannot be applied to expression of type '%s' at %s.",
		e.Op,
		e.Type,
		e.At,
	)
}

type InvalidBinaryExpressionTypes struct {
	Op    BinaryOp
	Left  cpp.Type
	Right cpp.Type
	At    Span
}

func (e InvalidBinaryExpressionTypes) Error() string {
	return fmt.Sprintf(
		"Binary operator %s cannot be applied to expressions of types '%s' and '%s' at %s.",
		e.Op,
		e.Left,
		e.Right,
		e.At,
	)
}

type UndefinedVariable Name

func (e UndefinedVariable) Error() string {
	return fmt.Sprintf("Variable %s used at %s is not defined.", e.String, e.Span)
}

type UndefinedType Name

func (e UndefinedType) Error() string {
	return fmt.Sprintf("Type %s used at %s is not defined.", e.String, e.Span)
}

type NotAFunction struct {
	Found cpp.Type
	At    Span
}

func (e NotAFunction) Error() string {
	return fmt.Sprintf("Function call on non-function type %s at %s.", e.Found, e.At)
}

type NotAType struct {
	Found cpp.Type
	At    Span
}

func (e NotAType) Error() string {
	return fmt.Sprintf("Non-type %s used as type at %s.", e.Found, e.At)
}

type TypeMismatch struct {
	Expected cpp.Type
	Found    cpp.Type
	At       Span
}

func (e TypeMismatch) Error() string {
	return fmt.Sprintf("Expected type %s, found type %s at %s.", e.Expected, e.Found, e.At)
}

type BranchTypeNotEqual struct {
	Then   cpp.Type
	ThenAt Span
	Else   cpp.Type
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

type CyclicDependency struct {
	First  string
	Second string
}

func (e CyclicDependency) Error() string {
	return fmt.Sprintf("Cyclic dependency between declarations %s and %s.", e.First, e.Second)
}
