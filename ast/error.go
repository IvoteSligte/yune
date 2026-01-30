package ast

import (
	"fmt"
	"yune/value"
)

type DuplicateDeclaration struct {
	First  Declaration
	Second Declaration
}

func (e DuplicateDeclaration) Error() string {
	return fmt.Sprintf("'%s' previously defined at %s redefined at %s.", e.First.GetName(), e.First.GetSpan(), e.Second.GetSpan())
}

type InvalidUnaryExpressionType struct {
	Op   UnaryOp
	Type value.Type
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
	Left  value.Type
	Right value.Type
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
	return fmt.Sprintf("Variable '%s' used at %s is not defined.", e.String, e.Span)
}

type UndefinedType Name

func (e UndefinedType) Error() string {
	return fmt.Sprintf("Type '%s' used at %s is not defined.", e.String, e.Span)
}

type NotAFunction struct {
	Found value.Type
	At    Span
}

func (e NotAFunction) Error() string {
	return fmt.Sprintf("Function call on non-function type '%s' at %s.", e.Found, e.At)
}

type UnexpectedType struct {
	Expected value.Type
	Found    value.Type
	At       Span
}

func (e UnexpectedType) Error() string {
	if e.Expected.Eq(value.TypeType) {
		return fmt.Sprintf("Non-type '%s' used as type at %s.", e.Found, e.At)
	} else {
		return fmt.Sprintf("Expected type '%s', but found type '%s' at %s.", e.Expected, e.Found, e.At)
	}
}

type AssignmentTypeMismatch struct {
	Expected value.Type
	Found    value.Type
	At       Span
}

func (e AssignmentTypeMismatch) Error() string {
	return fmt.Sprintf("Expected variable type '%s' for assignment, but found type '%s' at %s.", e.Expected, e.Found, e.At)
}

type ReturnTypeMismatch struct {
	Expected value.Type
	Found    value.Type
	At       Span
}

func (e ReturnTypeMismatch) Error() string {
	return fmt.Sprintf("Expected return type '%s', but found type '%s' at %s.", e.Expected, e.Found, e.At)
}

type VariableTypeMismatch struct {
	Expected value.Type
	Found    value.Type
	At       Span
}

func (e VariableTypeMismatch) Error() string {
	return fmt.Sprintf("Expected declared variable type '%s', but found type '%s' at %s.", e.Expected, e.Found, e.At)
}

type ConstantTypeMismatch struct {
	Expected value.Type
	Found    value.Type
	At       Span
}

func (e ConstantTypeMismatch) Error() string {
	return fmt.Sprintf("Expected declared constant type '%s', but found type '%s' at %s.", e.Expected, e.Found, e.At)
}

type ArgumentTypeMismatch struct {
	Expected value.Type
	Found    value.Type
	At       Span
}

func (e ArgumentTypeMismatch) Error() string {
	return fmt.Sprintf("Expected argument type '%s', but found type '%s' at %s.", e.Expected, e.Found, e.At)
}

type InvalidConditionType struct {
	Found value.Type
	At    Span
}

func (e InvalidConditionType) Error() string {
	return fmt.Sprintf("Expected type 'Bool' for condition, but found type '%s' at %s.", e.Found, e.At)
}

type BranchTypeNotEqual struct {
	Then   value.Type
	ThenAt Span
	Else   value.Type
	ElseAt Span
}

func (e BranchTypeNotEqual) Error() string {
	return fmt.Sprintf(
		"Types of branches '%s' at %s and '%s' at %s are not equal.",
		e.Then,
		e.ThenAt,
		e.Else,
		e.ElseAt,
	)
}

type InvalidMainSignature struct {
	Found value.Type
	At    Span
}

func (e InvalidMainSignature) Error() string {
	return fmt.Sprintf("The main function at %s must have a type signature of '%s', found '%s'.", e.At, value.MainType, e.Found)
}

type CyclicTypeDependency struct {
	In Declaration
}

func (e CyclicTypeDependency) Error() string {
	return fmt.Sprintf("Cyclic type dependency in declaration '%s' at %s.", e.In.GetName(), e.In.GetSpan())
}

type CyclicConstantDependency struct {
	In Declaration
}

func (e CyclicConstantDependency) Error() string {
	return fmt.Sprintf("Cyclic constant dependency in declaration '%s' at %s.", e.In.GetName(), e.In.GetSpan())
}
