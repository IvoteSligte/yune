package ast

import (
	"fmt"
	"log"
	"yune/cpp"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

type Expression interface {
	Node
	Value
	GetMacros() []*Macro

	// Get*Dependencies, but retrieves the dependencies added by evaluated macros.
	GetMacroTypeDependencies() []Query
	GetMacroValueDependencies() []Name

	GetTypeDependencies() []Query // TODO: remove, as it is always empty (unless blocks in expressions are allowed!!!)
	GetValueDependencies() []Name

	// Infers type, with an optional `expected` type for backwards inference.
	InferType(expected TypeValue, deps DeclarationTable) (errors Errors) // TODO: check that types match `expected` types
	GetType() TypeValue
	Lower() cpp.Expression
}

// Tries to unmarshal an Expression, returning nil if the union key does not match an Expression.
func UnmarshalExpression(data *fj.Value) (expr Expression) {
	key, v := fjUnmarshalUnion(data)
	switch key {
	case "Integer":
		expr = fjUnmarshal(v, &Integer{})
	case "Float":
		expr = fjUnmarshal(v, &Float{})
	case "Bool":
		expr = fjUnmarshal(v, &Bool{})
	case "String":
		expr = fjUnmarshal(v, &String{})
	case "Variable":
		expr = fjUnmarshal(v, &Variable{})
	case "FunctionCall":
		expr = &FunctionCall{
			Span:     fjUnmarshal(v.Get("Span"), Span{}),
			Function: UnmarshalExpression(v.Get("function")),
			Argument: UnmarshalExpression(v.Get("argument")),
		}
	case "Tuple":
		expr = &Tuple{
			Span:     fjUnmarshal(v.Get("Span"), Span{}),
			Elements: util.Map(v.Get("elements").GetArray(), UnmarshalExpression),
		}
	case "Macro":
		panic("Macros are not supported for serialization right now.")
	case "UnaryExpression":
		expr = &UnaryExpression{
			Span:       fjUnmarshal(v.Get("Span"), Span{}),
			Op:         UnaryOp(v.GetStringBytes("op")),
			Expression: UnmarshalExpression(v.Get("expression")),
		}
	case "BinaryExpression":
		expr = &BinaryExpression{
			Span:  fjUnmarshal(v.Get("Span"), Span{}),
			Op:    BinaryOp(v.GetStringBytes("op")),
			Left:  UnmarshalExpression(v.Get("left")),
			Right: UnmarshalExpression(v.Get("right")),
		}
	default:
		// expr = nil
	}
	return
}

type DefaultExpression struct{}

// Deserialize implements Expression.
func (d DefaultExpression) Deserialize([]byte) error {
	panic("DefaultExpression.Deserialize must be overridden")
}

// value implements Expression.
func (d DefaultExpression) value() {}

// GetMacroTypeDependencies implements Expression.
func (d DefaultExpression) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetMacroValueDependencies implements Expression.
func (d DefaultExpression) GetMacroValueDependencies() []Name {
	return []Name{}
}

// GetMacros implements Expression.
func (d DefaultExpression) GetMacros() []*Macro {
	return []*Macro{}
}

// GetSpan implements Expression.
func (d DefaultExpression) GetSpan() Span {
	panic("DefaultExpression.GetSpan() should be overridden")
}

// GetType implements Expression.
func (d DefaultExpression) GetType() TypeValue {
	panic("DefaultExpression.GetType() should be overridden")
}

// GetTypeDependencies implements Expression.
func (d DefaultExpression) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements Expression.
func (d DefaultExpression) GetValueDependencies() []Name {
	return []Name{}
}

// InferType implements Expression.
func (d DefaultExpression) InferType(expected TypeValue, deps DeclarationTable) (errors []error) {
	return
}

// Lower implements Expression.
func (d DefaultExpression) Lower() cpp.Expression {
	panic("DefaultExpression.Lower() should be overridden")
}

var _ Expression = DefaultExpression{}

type Integer struct {
	DefaultExpression
	Span  Span  `json:"span"`
	Value int64 `json:"value"`
}

// GetSpan implements Expression.
func (i Integer) GetSpan() Span {
	return i.Span
}

// Lower implements Expression.
func (i Integer) Lower() cpp.Expression {
	return cpp.Integer(i.Value)
}

// GetType implements Expression.
func (i Integer) GetType() TypeValue {
	return IntType{}
}

type Float struct {
	DefaultExpression
	Span  Span
	Value float64
}

// GetSpan implements Expression.
func (f Float) GetSpan() Span {
	return f.Span
}

// Lower implements Expression.
func (f Float) Lower() cpp.Expression {
	return cpp.Float(f.Value)
}

// GetType implements Expression.
func (f Float) GetType() TypeValue {
	return FloatType{}
}

type Bool struct {
	DefaultExpression
	Span  Span
	Value bool
}

// GetSpan implements Expression.
func (b Bool) GetSpan() Span {
	return b.Span
}

// Lower implements Expression.
func (b Bool) Lower() cpp.Expression {
	return cpp.Bool(b.Value)
}

// GetType implements Expression.
func (b Bool) GetType() TypeValue {
	return BoolType{}
}

type String struct {
	DefaultExpression
	Span  Span
	Value string
}

// GetSpan implements Expression.
func (s String) GetSpan() Span {
	return s.Span
}

// Lower implements Expression.
func (s String) Lower() cpp.Expression {
	return cpp.String(s.Value)
}

// GetType implements Expression.
func (s String) GetType() TypeValue {
	return StringType{}
}

type Variable struct {
	DefaultExpression
	Type TypeValue
	Name Name
}

// GetSpan implements Expression.
func (v *Variable) GetSpan() Span {
	return v.Name.Span
}

// GetType implements Expression.
func (v *Variable) GetType() TypeValue {
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
func (v *Variable) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	decl, _ := deps.Get(v.Name.String)
	if decl.GetDeclaredType() == nil {
		log.Printf("WARN: Type{} queried at %s before being calculated on declaration '%s'.", v.Name.Span, v.Name.String)
	}
	v.Type = decl.GetDeclaredType()
	return
}

// Lower implements Expression.
func (v *Variable) Lower() cpp.Expression {
	return cpp.Variable(v.Name.String)
}

type FunctionCall struct {
	DefaultExpression
	Span     Span
	Type     TypeValue
	Function Expression
	Argument Expression
}

// GetSpan implements Expression.
func (f *FunctionCall) GetSpan() Span {
	return f.Span
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
func (f *FunctionCall) GetType() TypeValue {
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
func (f *FunctionCall) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	errors = f.Function.InferType(noType, deps)
	if len(errors) > 0 {
		return
	}
	maybeFunctionType := f.Function.GetType()
	functionType, isFunction := maybeFunctionType.(FnType)
	if !isFunction {
		errors = append(errors, NotAFunction{
			Found: maybeFunctionType,
			At:    f.Function.GetSpan(),
		})
		return
	}
	errors = f.Argument.InferType(functionType.Argument, deps)
	if len(errors) > 0 {
		return
	}
	// single-argument functions still expect a tuple type for comparison
	argumentType := f.Argument.GetType()
	// NOTE: should functions return () instead of Nil?
	if !argumentType.Eq(functionType.Argument) {
		errors = append(errors, ArgumentTypeMismatch{
			Expected: functionType.Argument,
			Found:    argumentType,
			At:       f.Argument.GetSpan(),
		})
		return
	}
	f.Type = functionType.Return
	return
}

// Lower implements Expression.
func (f *FunctionCall) Lower() cpp.Expression {
	argumentType := f.Argument.GetType()
	_, isTuple := argumentType.(TupleType)
	if isTuple {
		// functions called with the empty tuple are lowered to functions called with nothing
		if argumentType.Eq(TupleType{}) {
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
	DefaultExpression
	Span Span
	// Inferred type
	Type     TypeValue
	Elements []Expression
}

// GetSpan implements Expression.
func (t *Tuple) GetSpan() Span {
	return t.Span
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
func (t *Tuple) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	expectedTupleType, isTuple := expected.(TupleType)

	for i, elem := range t.Elements {
		var expectedElementType TypeValue
		if expected.Eq(TypeType{}) {
			expectedElementType = TypeType{}
		}
		if isTuple && len(expectedTupleType.Elements) == len(t.Elements) {
			expectedElementType = expectedTupleType.Elements[i]
		}
		errors = append(errors, elem.InferType(expectedElementType, deps)...)
	}
	if expected.Eq(TypeType{}) {
		t.Type = TypeType{}
	} else {
		t.Type = NewTupleType(util.Map(t.Elements, func(e Expression) TypeValue {
			return e.GetType()
		})...)
	}
	return
}

// Lower implements Expression.
func (t *Tuple) Lower() cpp.Expression {
	if t.Type.Eq(TypeType{}) {
		if len(t.Elements) == 0 {
			return cpp.Raw(`box(ty::TupleType{{}})`)
		}
		elements := util.JoinFunction(t.Elements, ", ", func(e Expression) string {
			return e.Lower().String()
		})
		return cpp.Raw(fmt.Sprintf(`box(ty::TupleType{{%s}})`, elements))
	} else {
		return cpp.FunctionCall{
			Function:  cpp.Variable("std::make_tuple"),
			Arguments: util.Map(t.Elements, Expression.Lower),
		}
	}
}

// GetType implements Expression.
func (t *Tuple) GetType() TypeValue {
	return t.Type
}

// TODO: type check Function
type Macro struct {
	DefaultExpression
	Span Span
	// Function that evaluates the macro.
	Function Variable
	Lines    []MacroLine
	// Result after evaluating the macro.
	Result Expression
}

// SetValue implements Destination.
func (m *Macro) SetValue(v Value) {
	// FIXME: in Yune: (String, Expression) -> after serialization: (String, String)
	//     because both are simply stored as Expression
	m.Result = v.(Expression)
}

// GetSpan implements Expression.
func (m *Macro) GetSpan() Span {
	return m.Span
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
	return m.Function.GetTypeDependencies()
}

// GetValueDependencies implements Expression.
func (m *Macro) GetValueDependencies() []Name {
	return m.Function.GetValueDependencies()
}

// InferType implements Expression.
func (m *Macro) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	return m.Result.InferType(expected, deps)
}

// Lower implements Expression.
func (m *Macro) Lower() cpp.Expression {
	return m.Result.Lower()
}

// GetType implements Expression.
func (m *Macro) GetType() TypeValue {
	return m.Result.GetType()
}

type MacroLine struct {
	Span
	Text string
}

type UnaryExpression struct {
	DefaultExpression
	Span       Span
	Type       TypeValue
	Op         UnaryOp
	Expression Expression
}

// GetSpan implements Expression.
func (u *UnaryExpression) GetSpan() Span {
	return u.Span
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
func (u *UnaryExpression) GetType() TypeValue {
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
func (u *UnaryExpression) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	errors = u.Expression.InferType(expected, deps)
	if len(errors) > 0 {
		return
	}
	expressionType := u.Expression.GetType()
	switch {
	case
		expressionType.Eq(IntType{}),
		expressionType.Eq(FloatType{}):
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
	DefaultExpression
	Span  Span
	Type  TypeValue
	Op    BinaryOp
	Left  Expression
	Right Expression
}

// GetSpan implements Expression.
func (b *BinaryExpression) GetSpan() Span {
	return b.Span
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
func (b *BinaryExpression) GetType() TypeValue {
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
func (b *BinaryExpression) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
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
		if !leftType.Eq(IntType{}) && !leftType.Eq(FloatType{}) {
			emitErr()
			return
		}
		b.Type = leftType
	case
		Equal,
		NotEqual:
		b.Type = BoolType{}
	case
		Or, And:
		if !leftType.Eq(BoolType{}) {
			emitErr()
			return
		}
		b.Type = BoolType{}
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

var _ Expression = &Integer{}
var _ Expression = &Float{}
var _ Expression = &Bool{}
var _ Expression = &String{}
var _ Expression = &Variable{}
var _ Expression = &FunctionCall{}
var _ Expression = &Tuple{}
var _ Expression = &Macro{}
var _ Expression = &UnaryExpression{}
var _ Expression = &BinaryExpression{}
