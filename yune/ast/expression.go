package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"
)

type Variable struct {
	InferredType
	Name
}

// Lower implements Expression.
func (v *Variable) Lower() cpp.Expression {
	panic("unimplemented")
}

// Analyze implements Expression.
func (v *Variable) Analyze() (queries Queries, finalizer Finalizer) {
	queries = []Query{
		v.Name,
	}
	finalizer = func(env Env) (errors Errors) {
		v.InferredType = env.Get(v.Name.String).GetType()
		return
	}
	return
}

type FunctionCall struct {
	Span
	InferredType
	Function Expression
	Argument Expression
}

// GetSpan implements Express
// Lower implements Expression.
func (f *FunctionCall) Lower() cpp.Expression {
	panic("unimplemented")
}

// FIXME: since type queries and constant queries are combined, it is not possible to do recursion
// 	considering that that forms a cycle

// Analyze implements Expression.
func (f *FunctionCall) Analyze() (queries Queries, finalizer Finalizer) {
	queries, functionFinalizer := f.Function.Analyze()
	_queries, argumentFinalizer := f.Argument.Analyze()
	queries = append(queries, _queries...)
	finalizer = func(env Env) (errors Errors) {
		errors = append(functionFinalizer(env), argumentFinalizer(env)...)
		if len(errors) > 0 {
			return
		}
		if !f.Function.GetType().IsFunction() {
			errors = append(errors, NotAFunction{Found: f.Function.GetType()})
			return
		}
		expected := f.Function.GetType().GetGeneric(0)
		found := f.Argument.GetType()
		if !expected.Eq(found) {
			errors = append(errors, TypeMismatch{
				Expected: expected,
				Found:    found,
				At:       f.Argument.GetSpan(),
			})
		}
		f.InferredType = f.Function.GetType().GetGeneric(1)
		return
	}
	return
}

type Tuple struct {
	Span
	Elements []Expression
}

// Lower implements Expression.
func (t *Tuple) Lower() cpp.Expression {
	panic("unimplemented")
}

// Analyze implements Expression.
func (t *Tuple) Analyze() (queries Queries, _finalizer Finalizer) {
	finalizers := make([]Finalizer, len(t.Elements))
	for i := range t.Elements {
		_queries, _finalizer := t.Elements[i].Analyze()
		queries = append(queries, _queries...)
		finalizers = append(finalizers, _finalizer)
	}
	_finalizer = func(env Env) (errors Errors) {
		for _, fin := range finalizers {
			errors = append(errors, fin(env)...)
		}
		return
	}
	return
}

// GetType implements Expression.
func (t *Tuple) GetType() InferredType {
	return InferredType{
		name:     "Tuple",
		generics: util.Map(t.Elements, Expression.GetType),
	}
}

type Macro struct {
	// TODO: indicate macro text with a special span
	Span
	Language Name
	Text     string
	// Result after evaluating the macro.
	Result Expression
}

// Lower implements Expression.
func (m *Macro) Lower() cpp.Expression {
	panic("unimplemented")
}

// Analyze implements Expression.
func (m *Macro) Analyze() (queries Queries, finalizer Finalizer) {
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

// Lower implements Expression.
func (u *UnaryExpression) Lower() cpp.Expression {
	panic("unimplemented")
}

// Analyze implements Expression.
func (u *UnaryExpression) Analyze() (queries Queries, finalizer Finalizer) {
	queries, fin := u.Expression.Analyze()
	finalizer = func(env Env) (errors Errors) {
		errors = fin(env)
		if len(errors) > 0 {
			return
		}
		expressionType := u.Expression.GetType()
		switch u.Op {
		case Negate:
			switch {
			case !expressionType.Eq(IntType):
				u.InferredType = IntType
			case !expressionType.Eq(FloatType):
				u.InferredType = FloatType
			default:
				errors = append(errors, InvalidUnaryExpressionType{
					Op:   u.Op,
					Type: expressionType,
					At:   u.Span,
				})
			}
		default:
			panic(fmt.Sprintf("unexpected ast.UnaryOp: %#v", u.Op))
		}
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
	InferredType
	Op    BinaryOp
	Left  Expression
	Right Expression
}

// Lower implements Expression.
func (b *BinaryExpression) Lower() cpp.Expression {
	panic("unimplemented")
}

// Analyze implements Expression.
func (b *BinaryExpression) Analyze() (queries Queries, finalizer Finalizer) {
	queries, finLeft := b.Left.Analyze()
	_queries, finRight := b.Right.Analyze()
	queries = append(queries, _queries...)

	finalizer = func(env Env) (errors Errors) {
		errors = append(finLeft(env), finRight(env)...)
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

		switch b.Op {
		case Add:
		case Subtract:
		case Multiply:
		case Divide:
		case Greater:
		case GreaterEqual:
		case Less:
		case LessEqual:
		case Equal:
		case NotEqual:
			switch {
			case !leftType.Eq(IntType):
				b.InferredType = IntType
			case !leftType.Eq(FloatType):
				b.InferredType = FloatType
			default:
				errors = append(errors, InvalidBinaryExpressionTypes{
					Op:    b.Op,
					Left:  leftType,
					Right: rightType,
					At:    b.Span,
				})
			}
		default:
			panic(fmt.Sprintf("unexpected ast.BinaryOp: %#v", b.Op))
		}
		return
	}
	return
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

// Lower implements Expression.
func (i Integer) Lower() cpp.Expression {
	panic("unimplemented")
}

// GetType implements Expression.
func (i Integer) GetType() InferredType {
	return IntType
}

// Analyze implements Expression.
func (i Integer) Analyze() (queries Queries, finalizer Finalizer) {
	finalizer = func(Env) (errors Errors) { return }
	return
}

type Float struct {
	Span
	Value float64
}

// Lower implements Expression.
func (i Float) Lower() cpp.Expression {
	return cpp.Float(i.Value)
}

// Analyze implements Expression.
func (i Float) Analyze() (queries Queries, finalizer Finalizer) {
	finalizer = func(Env) (errors Errors) { return }
	return
}

// GetType implements Expression.
func (f Float) GetType() InferredType {
	return FloatType
}

type Finalizer = func(env Env) (errors Errors)

type Expression interface {
	INode
	Analyze() (queries Queries, finalizer Finalizer)
	GetType() InferredType
	Lower() cpp.Expression
}

var _ Expression = Integer{}
var _ Expression = Float{}
var _ Expression = &Variable{}
var _ Expression = &FunctionCall{}
var _ Expression = &Tuple{}
var _ Expression = &Macro{}
var _ Expression = &UnaryExpression{}
var _ Expression = &BinaryExpression{}
