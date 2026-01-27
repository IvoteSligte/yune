package ast

import (
	"fmt"
	"log"
	"slices"
	"yune/cpp"
	"yune/util"
	"yune/value"
)

type Integer struct {
	Span
	Value int64
}

// GetMacros implements Expression.
func (i Integer) GetMacros() []*Macro {
	return []*Macro{}
}

// GetMacroTypeDependencies implements Expression.
func (i Integer) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetMacroValueDependencies implements Expression.
func (i Integer) GetMacroValueDependencies() (deps []Name) {
	return
}

// GetTypeDependencies implements Expression.
func (i Integer) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements Expression.
func (i Integer) GetValueDependencies() (deps []Name) {
	return
}

// InferType implements Expression.
func (i Integer) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
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

// GetMacros implements Expression.
func (f Float) GetMacros() []*Macro {
	return []*Macro{}
}

// GetMacroTypeDependencies implements Expression.
func (f Float) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetMacroValueDependencies implements Expression.
func (f Float) GetMacroValueDependencies() (deps []Name) {
	return
}

// GetTypeDependencies implements Expression.
func (f Float) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements Expression.
func (f Float) GetValueDependencies() (deps []Name) {
	return
}

// InferType implements Expression.
func (f Float) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
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

// GetMacros implements Expression.
func (f Bool) GetMacros() []*Macro {
	return []*Macro{}
}

// GetMacroTypeDependencies implements Expression.
func (f Bool) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetMacroValueDependencies implements Expression.
func (f Bool) GetMacroValueDependencies() (deps []Name) {
	return
}

// GetTypeDependencies implements Expression.
func (f Bool) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements Expression.
func (f Bool) GetValueDependencies() (deps []Name) {
	return
}

// InferType implements Expression.
func (f Bool) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
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

// GetMacros implements Expression.
func (f String) GetMacros() []*Macro {
	return []*Macro{}
}

// GetMacroTypeDependencies implements Expression.
func (f String) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetMacroValueDependencies implements Expression.
func (f String) GetMacroValueDependencies() (deps []Name) {
	return
}

// GetTypeDependencies implements Expression.
func (f String) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements Expression.
func (f String) GetValueDependencies() (deps []Name) {
	return
}

// InferType implements Expression.
func (f String) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
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

// GetMacros implements Expression.
func (v *Variable) GetMacros() []*Macro {
	return []*Macro{}
}

// GetMacroTypeDependencies implements Expression.
func (v *Variable) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetTypeDependencies implements Expression.
func (v *Variable) GetTypeDependencies() []Query {
	return []Query{}
}

// GetType implements Expression.
func (v *Variable) GetType() value.Type {
	return v.Type
}

// GetMacroValueDependencies implements Expression.
func (v *Variable) GetMacroValueDependencies() (deps []Name) {
	return []Name{v.Name}
}

// GetValueDependencies implements Expression.
func (v *Variable) GetValueDependencies() (deps []Name) {
	return []Name{v.Name}
}

// InferType implements Expression.
func (v *Variable) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
	decl, ok := deps.Get(v.Name.String)
	if !ok {
		errors = append(errors, UndefinedVariable(v.Name))
		return
	}
	if decl.GetType() == value.Type("") {
		log.Printf("WARN: Type queried at %s before being calculated on declaration '%s'.", v.Span, v.Name.String)
	}
	v.Type = decl.GetType()
	return
}

// Lower implements Expression.
func (v *Variable) Lower() cpp.Expression {
	return cpp.Variable(v.Name.String)
}

type FunctionCall struct {
	Span
	value.Type
	Function Expression
	Argument Expression
}

// GetMacros implements Expression.
func (f *FunctionCall) GetMacros() []*Macro {
	return append(f.Function.GetMacros(), f.Argument.GetMacros()...)
}

// GetMacroTypeDependencies implements Expression.
func (f *FunctionCall) GetMacroTypeDependencies() []Query {
	return append(f.Function.GetMacroTypeDependencies(), f.Argument.GetMacroTypeDependencies()...)
}

// GetTypeDependencies implements Expression.
func (f *FunctionCall) GetTypeDependencies() []Query {
	return append(f.Function.GetTypeDependencies(), f.Argument.GetTypeDependencies()...)
}

// GetType implements Expression.
func (f *FunctionCall) GetType() value.Type {
	return f.Type
}

// GetMacroValueDependencies implements Expression.
func (f *FunctionCall) GetMacroValueDependencies() []Name {
	return append(f.Function.GetMacroValueDependencies(), f.Argument.GetMacroValueDependencies()...)
}

// GetValueDependencies implements Expression.
func (f *FunctionCall) GetValueDependencies() []Name {
	return append(f.Function.GetValueDependencies(), f.Argument.GetValueDependencies()...)
}

// InferType implements Expression.
func (f *FunctionCall) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
	errors = f.Function.InferType(unknownType, deps)
	if len(errors) > 0 {
		return
	}
	maybeFunctionType := f.Function.GetType()
	parameterType, returnType, isFunction := maybeFunctionType.ToFunction()
	if !isFunction {
		errors = append(errors, NotAFunction{
			Found: maybeFunctionType,
			At:    f.Function.GetSpan(),
		})
		return
	}
	errors = f.Argument.InferType(parameterType, deps)
	if len(errors) > 0 {
		return
	}
	// single-argument functions still expect a std::tuple type for comparison
	argumentType := f.Argument.GetType()
	if argumentType.Eq(NilType) {
		// NOTE: should functions return () instead of Nil?
		argumentType = value.Type("std::tuple<>")
	} else if !argumentType.IsTuple() {
		argumentType = value.NewTupleType([]value.Type{argumentType})
	}
	if !argumentType.Eq(parameterType) {
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
				Arguments: []cpp.Expression{}, // FIXME: currently does not execute argument
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
	// Inferred type
	Type     value.Type
	Elements []Expression
}

// GetMacros implements Expression.
func (t *Tuple) GetMacros() []*Macro {
	return util.FlatMap(t.Elements, Expression.GetMacros)
}

// GetMacroTypeDependencies implements Expression.
func (t *Tuple) GetMacroTypeDependencies() []Query {
	return util.FlatMap(t.Elements, Expression.GetMacroTypeDependencies)
}

// GetMacroValueDependencies implements Expression.
func (t *Tuple) GetMacroValueDependencies() (deps []Name) {
	for i := range t.Elements {
		deps = append(deps, t.Elements[i].GetMacroValueDependencies()...)
	}
	return
}

// GetTypeDependencies implements Expression.
func (t *Tuple) GetTypeDependencies() []Query {
	return util.FlatMap(t.Elements, Expression.GetTypeDependencies)
}

// GetValueDependencies implements Expression.
func (t *Tuple) GetValueDependencies() (deps []Name) {
	for i := range t.Elements {
		deps = append(deps, t.Elements[i].GetValueDependencies()...)
	}
	return
}

// InferType implements Expression.
func (t *Tuple) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
	expectedElementTypes := slices.Repeat([]value.Type{unknownType}, len(t.Elements))
	if expected.Eq(TypeType) {
		expectedElementTypes = slices.Repeat([]value.Type{TypeType}, len(t.Elements))
	} else {
		elementTypes, isTuple := expected.ToTuple()
		// if false, type checking is done by the caller
		if isTuple && len(elementTypes) == len(t.Elements) {
			expectedElementTypes = elementTypes
		}
	}
	for i, elem := range t.Elements {
		errors = append(errors, elem.InferType(expectedElementTypes[i], deps)...)
	}
	if expected.Eq(TypeType) {
		t.Type = TypeType
	} else {
		t.Type = value.NewTupleType(util.Map(t.Elements, func(e Expression) value.Type {
			return e.GetType()
		}))
	}
	return
}

// Lower implements Expression.
func (t *Tuple) Lower() cpp.Expression {
	if t.Type.Eq(TypeType) {
		if len(t.Elements) == 0 {
			return cpp.Raw(`Type{"std::tuple<>"}`)
		}
		elements := util.JoinFunction(t.Elements, ` + ", " + `, func(e Expression) string {
			return fmt.Sprintf("(%s).id", e.Lower())
		})
		return cpp.Raw(fmt.Sprintf(`Type{"std::tuple<" + %s + ">"}`, elements))
	} else {
		return cpp.FunctionCall{
			Function:  cpp.Variable("std::make_tuple"),
			Arguments: util.Map(t.Elements, Expression.Lower),
		}
	}
}

// GetType implements Expression.
func (t *Tuple) GetType() value.Type {
	return t.Type
}

type Macro struct {
	Span
	// Function that evaluates the macro.
	Function Variable
	Lines    []MacroLine
	// Result after evaluating the macro.
	Result Expression
}

func (m *Macro) GetText() string {
	return util.JoinFunction(m.Lines, "\n", func(l MacroLine) string {
		return l.Text
	})
}

func (m *Macro) AsFunctionCall() FunctionCall {
	return FunctionCall{
		Span:     m.Span,
		Function: &m.Function,
		Argument: String{
			Span:  m.Lines[0].Span,
			Value: m.GetText(),
		},
	}
}

// GetMacros implements Expression.
func (m *Macro) GetMacros() []*Macro {
	return []*Macro{m}
}

// GetMacroTypeDependencies implements Expression.
func (m *Macro) GetMacroTypeDependencies() []Query {
	return m.Result.GetTypeDependencies()
}

// GetMacroValueDependencies implements Expression.
func (m *Macro) GetMacroValueDependencies() []Name {
	return m.Result.GetValueDependencies()
}

// GetTypeDependencies implements Expression.
func (m *Macro) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements Expression.
func (m *Macro) GetValueDependencies() []Name {
	return []Name{}
}

// InferType implements Expression.
func (m *Macro) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
	return m.Result.InferType(expected, deps)
}

// Lower implements Expression.
func (m *Macro) Lower() cpp.Expression {
	return m.Result.Lower()
}

// GetType implements Expression.
func (m *Macro) GetType() value.Type {
	return m.Result.GetType()
}

type MacroLine struct {
	Span
	Text string
}

// SetValue implements value.Destination.
func (m *Macro) SetValue(s string) {
	panic("unimplemented")
}

type UnaryExpression struct {
	Span
	value.Type
	Op         UnaryOp
	Expression Expression
}

// GetMacros implements Expression.
func (u *UnaryExpression) GetMacros() []*Macro {
	return u.Expression.GetMacros()
}

// GetMacroTypeDependencies implements Expression.
func (u *UnaryExpression) GetMacroTypeDependencies() []Query {
	return u.Expression.GetMacroTypeDependencies()
}

// GetTypeDependencies implements Expression.
func (u *UnaryExpression) GetTypeDependencies() []Query {
	return u.Expression.GetTypeDependencies()
}

// GetType implements Expression.
func (u *UnaryExpression) GetType() value.Type {
	return u.Type
}

// GetMacroValueDependencies implements Expression.
func (u *UnaryExpression) GetMacroValueDependencies() []Name {
	return u.Expression.GetMacroValueDependencies()
}

// GetValueDependencies implements Expression.
func (u *UnaryExpression) GetValueDependencies() []Name {
	return u.Expression.GetValueDependencies()
}

// InferType implements Expression.
func (u *UnaryExpression) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
	errors = u.Expression.InferType(expected, deps)
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

// GetMacros implements Expression.
func (b *BinaryExpression) GetMacros() []*Macro {
	return append(b.Left.GetMacros(), b.Right.GetMacros()...)
}

// GetMacroTypeDependencies implements Expression.
func (b *BinaryExpression) GetMacroTypeDependencies() []Query {
	return append(b.Left.GetMacroTypeDependencies(), b.Right.GetMacroTypeDependencies()...)
}

// GetTypeDependencies implements Expression.
func (b *BinaryExpression) GetTypeDependencies() []Query {
	return append(b.Left.GetTypeDependencies(), b.Right.GetTypeDependencies()...)
}

// GetType implements Expression.
func (b *BinaryExpression) GetType() value.Type {
	return b.Type
}

// GetMacroValueDependencies implements Expression.
func (b *BinaryExpression) GetMacroValueDependencies() []Name {
	return append(b.Left.GetMacroValueDependencies(), b.Right.GetMacroValueDependencies()...)
}

// GetValueDependencies implements Expression.
func (b *BinaryExpression) GetValueDependencies() []Name {
	return append(b.Left.GetValueDependencies(), b.Right.GetValueDependencies()...)
}

// InferType implements Expression.
func (b *BinaryExpression) InferType(expected value.Type, deps DeclarationTable) (errors Errors) {
	errors = append(b.Left.InferType(expected, deps), b.Right.InferType(expected, deps)...)
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
	GetMacros() []*Macro

	// Get*Dependencies, but retrieves the dependencies added by evaluated macros.
	GetMacroTypeDependencies() []Query
	GetMacroValueDependencies() []Name

	GetTypeDependencies() []Query // TODO: remove, as it is always empty (unless blocks in expressions are allowed!!!)
	GetValueDependencies() []Name

	// Infers type, with an optional `expected` type for backwards inference.
	InferType(expected value.Type, deps DeclarationTable) (errors Errors) // TODO: check that types match `expected` types
	GetType() value.Type
	Lower() cpp.Expression
}

var _ Expression = Integer{}
var _ Expression = Float{}
var _ Expression = Bool{}
var _ Expression = String{}
var _ Expression = &Variable{}
var _ Expression = &FunctionCall{}
var _ Expression = &Tuple{}
var _ Expression = &Macro{}
var _ Expression = &UnaryExpression{}
var _ Expression = &BinaryExpression{}
