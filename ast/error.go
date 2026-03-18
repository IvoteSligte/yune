package ast

import (
	"fmt"
	"slices"
	"strings"
)

func makeCodeError(errorText string, span Span, message string) string {
	sourceLine := slices.Collect(strings.Lines(span.Source))[span.Line]
	if sourceLine[len(sourceLine)-1] == '\n' {
		sourceLine = sourceLine[:len(sourceLine)-1]
	}
	leftPad := strings.Repeat(" ", len(fmt.Sprintf("%d", span.Line)))
	midPad := strings.Repeat(" ", span.Column)
	underline := strings.Repeat("~", max(span.Length, 1))
	return fmt.Sprintf(`
%s
'%s'
---> %s line %d column %d
%s |
%d |  %s
%s |  %s%s %s`,
		errorText,
		span.Content,
		span.File, span.Line, span.Column,
		leftPad,
		span.Line, sourceLine,
		leftPad, midPad, underline, message,
	)
}

type DuplicateDeclaration struct {
	First  Declaration
	Second Declaration
}

func (e DuplicateDeclaration) Error() string {
	text := fmt.Sprintf("'%s' previously defined at %s redefined at %s.", e.First.GetName().String, e.First.GetSpan(), e.Second.GetSpan())
	return makeCodeError(text, e.Second.GetSpan(), "redefined here")
}

type InvalidUnaryExpressionType struct {
	Op   UnaryOp
	Type TypeValue
	At   Span
}

func (e InvalidUnaryExpressionType) Error() string {
	text := fmt.Sprintf(
		"Unary operator %s cannot be applied to expression of type '%s'.",
		e.Op, e.Type,
	)
	return makeCodeError(text, e.At, "")
}

type InvalidBinaryExpressionTypes struct {
	Op    BinaryOp
	Left  TypeValue
	Right TypeValue
	At    Span
}

func (e InvalidBinaryExpressionTypes) Error() string {
	text := fmt.Sprintf(
		"Binary operator %s cannot be applied to expressions of types '%s' and '%s'.",
		e.Op, e.Left, e.Right,
	)
	return makeCodeError(text, e.At, "")
}

type UndefinedVariable Name

func (e UndefinedVariable) Error() string {
	text := fmt.Sprintf("Variable '%s' is not defined.", e.String)
	return makeCodeError(text, e.Span, "")
}

type NotAFunction struct {
	Found TypeValue
	At    Span
}

func (e NotAFunction) Error() string {
	text := fmt.Sprintf("Function call on non-function type '%s'.", e.Found)
	return makeCodeError(text, e.At, "")
}

type NotAStruct struct {
	Found TypeValue
	At    Span
}

func (e NotAStruct) Error() string {
	text := fmt.Sprintf("Tried to construct non-struct type '%s'.", e.Found)
	return makeCodeError(text, e.At, "")
}

type ArityMismatch struct {
	Expected int
	Found    int
	At       Span
}

func (e ArityMismatch) Error() string {
	text := fmt.Sprintf("Expected a tuple with arity %d, but found a tuple with arity %d.", e.Expected, e.Found)
	return makeCodeError(text, e.At, "")
}

type UnexpectedType struct {
	Expected TypeValue
	Found    TypeValue
	At       Span
}

func (e UnexpectedType) Error() string {
	var text string
	if e.Expected.Eq(&TypeType{}) {
		text = fmt.Sprintf("Non-type '%s' used as type.", e.Found)
	} else {
		text = fmt.Sprintf("Expected type '%s', but found type '%s'.", e.Expected, e.Found)
	}
	return makeCodeError(text, e.At, "")
}

type ExpectedTuple struct {
	Found TypeValue
	At    Span
}

func (e ExpectedTuple) Error() string {
	text := fmt.Sprintf("Expected tuple, but found type '%s'.", e.Found)
	return makeCodeError(text, e.At, "")
}

type AssignmentTypeMismatch struct {
	Expected TypeValue
	Found    TypeValue
	At       Span
}

func (e AssignmentTypeMismatch) Error() string {
	text := fmt.Sprintf("Expected variable type '%s' for assignment, but found type '%s'.", e.Expected, e.Found)
	return makeCodeError(text, e.At, "")
}

type ReturnTypeMismatch struct {
	Expected TypeValue
	Found    TypeValue
	At       Span
}

func (e ReturnTypeMismatch) Error() string {
	text := fmt.Sprintf("Expected return type '%s', but found type '%s'.", e.Expected, e.Found)
	return makeCodeError(text, e.At, "")
}

type VariableTypeMismatch struct {
	Expected TypeValue
	Found    TypeValue
	At       Span
}

func (e VariableTypeMismatch) Error() string {
	text := fmt.Sprintf("Expected declared variable type '%s', but found type '%s'.", e.Expected, e.Found)
	return makeCodeError(text, e.At, "")
}

type ConstantTypeMismatch struct {
	Expected TypeValue
	Found    TypeValue
	At       Span
}

func (e ConstantTypeMismatch) Error() string {
	text := fmt.Sprintf("Expected declared constant type '%s', but found type '%s'.", e.Expected, e.Found)
	return makeCodeError(text, e.At, "")
}

type ArgumentTypeMismatch struct {
	Expected TypeValue
	Found    TypeValue
	At       Span
}

func (e ArgumentTypeMismatch) Error() string {
	text := fmt.Sprintf("Expected argument type '%s', but found type '%s'.", e.Expected, e.Found)
	return makeCodeError(text, e.At, "")
}

type InvalidConditionType struct {
	Found TypeValue
	At    Span
}

func (e InvalidConditionType) Error() string {
	text := fmt.Sprintf("Expected type 'Bool' for condition, but found type '%s'.", e.Found)
	return makeCodeError(text, e.At, "")
}

type ImpossibleIsExpression struct {
	SuperType TypeValue
	SubType   TypeValue
	At        Span
}

func (e ImpossibleIsExpression) Error() string {
	text := fmt.Sprintf("Is-expression is always false because '%s' is not a sub-type of '%s'.", e.SubType, e.SuperType)
	return makeCodeError(text, e.At, "")
}

type InvalidMainSignature struct {
	Found TypeValue
	At    Span
}

func (e InvalidMainSignature) Error() string {
	text := fmt.Sprintf("The main function must have a type signature of '%s', found '%s'.", MainType, e.Found)
	return makeCodeError(text, e.At, "")
}

type CyclicTypeDependency struct {
	In Declaration
}

func (e CyclicTypeDependency) Error() string {
	text := fmt.Sprintf("Cyclic type dependency in declaration '%s'.", e.In.GetName())
	return makeCodeError(text, e.In.GetSpan(), "")
}

type CyclicConstantDependency struct {
	In Declaration
}

func (e CyclicConstantDependency) Error() string {
	text := fmt.Sprintf("Cyclic constant dependency in declaration '%s'.", e.In.GetName())
	return makeCodeError(text, e.In.GetSpan(), "")
}

type MacroRequestedUndefinedVariable struct {
	Macro Variable
	Name  string
}

func (e MacroRequestedUndefinedVariable) Error() string {
	// TODO: use span of e.Name instead of e.Macro
	return fmt.Sprintf(
		"Macro '%s' requested undefined variable '%s' at %s.",
		e.Macro.Name.String, e.Name, e.Macro.Name.Span,
	)
}

type MacroOutputError struct {
	Macro   Variable
	Message string
}

func (e MacroOutputError) Error() string {
	return fmt.Sprintf(
		"Macro '%s' at %s returned error: %s",
		e.Macro.Name.String, e.Macro.Name.Span, e.Message,
	)
}
