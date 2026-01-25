package ast

import (
	"fmt"
	"log"
	"yune/cpp"
	"yune/util"
	"yune/value"
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
func (i Integer) GetType() value.Type {
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
func (f Float) GetType() value.Type {
	return FloatType
}

type Bool struct {
	Span
	Value bool
}

// GetGlobalDependencies implements Expression.
func (f Bool) GetGlobalDependencies() (deps []string) {
	return
}

// InferType implements Expression.
func (f Bool) InferType(deps DeclarationTable) (errors Errors) {
	return
}

// Lower implements Expression.
func (f Bool) Lower() cpp.Expression {
	return cpp.Bool(f.Value)
}

// GetType implements Expression.
func (f Bool) GetType() value.Type {
	return BoolType
}

type String struct {
	Span
	Value string
}

// GetGlobalDependencies implements Expression.
func (f String) GetGlobalDependencies() (deps []string) {
	return
}

// InferType implements Expression.
func (f String) InferType(deps DeclarationTable) (errors Errors) {
	return
}

// Lower implements Expression.
func (f String) Lower() cpp.Expression {
	return cpp.String(f.Value)
}

// GetType implements Expression.
func (f String) GetType() value.Type {
	return StringType
}

type Variable struct {
	value.Type
	Name
}

// GetType implements Expression.
func (v *Variable) GetType() value.Type {
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
	if decl.GetType() == value.Type("") {
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
	value.Type
	Function Expression
	Argument Expression
}

// GetType implements Expression.
func (f *FunctionCall) GetType() value.Type {
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
	maybeFunctionType := f.Function.GetType()
	argumentType := f.Argument.GetType()
	parameterType, returnType, isFunction := maybeFunctionType.ToFunction()
	if !isFunction {
		errors = append(errors, NotAFunction{
			Found: maybeFunctionType,
			At:    f.Function.GetSpan(),
		})
		return
	}
	if !argumentType.IsSubType(parameterType) {
		errors = append(errors, ArgumentTypeMismatch{
			Expected: parameterType,
			Found:    argumentType,
			At:       f.Argument.GetSpan(),
		})
		return
	}
	f.Type = returnType
	return
}

// Lower implements Expression.
func (f *FunctionCall) Lower() cpp.Expression {
	argumentType := f.Argument.GetType()
	if argumentType.IsTuple() {
		// functions called with the empty tuple are lowered to functions called with nothing
		if argumentType.IsEmptyTuple() {
			return cpp.FunctionCall{
				Function:  f.Function.Lower(),
				Arguments: []cpp.Expression{},
			}
		}
		// calls the function with a tuple of arguments
		return cpp.FunctionCall{
			Function: cpp.Variable("std::apply"),
			Arguments: []cpp.Expression{
				f.Function.Lower(),
				f.Argument.Lower(),
			},
		}
	}
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
	if t.GetType().IsTypeType() { // FIXME: Fn(Type, Type) now takes Fn(Type) because of this rule, is that correct?
		if len(t.Elements) == 0 {
			return cpp.RawCpp(`Type{"std::tuple<>"}`)
		}
		elements := util.JoinFunction(t.Elements, ` + ", " + `, func(e Expression) string {
			return fmt.Sprintf("(%s).id", e.Lower())
		})
		return cpp.RawCpp(fmt.Sprintf(`Type{"std::tuple<" + %s + ">"}`, elements))
	}
	return cpp.FunctionCall{
		Function:  cpp.Variable("std::make_tuple"),
		Arguments: util.Map(t.Elements, Expression.Lower),
	}
}

// GetType implements Expression.
func (t *Tuple) GetType() value.Type {
	_type := value.NewTupleType(util.Map(t.Elements, Expression.GetType))
	return _type
}

type Macro struct {
	Span
	Language Variable
	Lines    []MacroLine
	// Result after evaluating the macro.
	Result Expression
}

type MacroLine struct {
	Span
	Text string
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
func (m *Macro) GetType() value.Type {
	return m.Result.GetType()
}

type UnaryExpression struct {
	Span
	value.Type
	Op         UnaryOp
	Expression Expression
}

// GetType implements Expression.
func (u *UnaryExpression) GetType() value.Type {
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
	value.Type
	Op    BinaryOp
	Left  Expression
	Right Expression
}

// GetType implements Expression.
func (b *BinaryExpression) GetType() value.Type {
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
	emitErr := func() {
		errors = append(errors, InvalidBinaryExpressionTypes{
			Op:    b.Op,
			Left:  leftType,
			Right: rightType,
			At:    b.Span,
		})
	}
	switch b.Op {
	case
		Add,
		Divide,
		Multiply,
		Subtract,
		Greater,
		GreaterEqual,
		Less,
		LessEqual:
		if !leftType.Eq(IntType) && !leftType.Eq(FloatType) {
			emitErr()
			return
		}
		b.Type = leftType
	case
		Equal,
		NotEqual:
		b.Type = BoolType
	case
		Or, And:
		if !leftType.Eq(BoolType) {
			emitErr()
			return
		}
		b.Type = BoolType
	default:
		panic(fmt.Sprintf("unexpected ast.BinaryOp: %#v", b.Op))
	}
	return
}

// Lower implements Expression.
func (b *BinaryExpression) Lower() cpp.Expression {
	var op string
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
		op = string(b.Op)
	case Or:
		op = "||"
	case And:
		op = "&&"
	default:
		panic(fmt.Sprintf("unexpected ast.BinaryOp: %#v", b.Op))
	}
	return cpp.BinaryExpression{
		Op:    cpp.BinaryOp(op),
		Left:  b.Left.Lower(),
		Right: b.Right.Lower(),
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
	Or           BinaryOp = "or"
	And          BinaryOp = "and"
)

type Expression interface {
	Node
	GetGlobalDependencies() []string
	InferType(deps DeclarationTable) (errors Errors)
	GetType() value.Type
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
