package ast

import (
	"encoding/json"
	"fmt"
	"log"
	"yune/cpp"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

type Expression interface {
	Node
	Value
	StatementBase
	GetMacros() []*Macro

	// Sets the type, in order to resolve ambiguities when the expression needs
	// to be inferred differently from the default. This is the case when an
	// Expression is used in a Type, or represents syntax the user has generated.
	SetType(t TypeValue)
	InferType(deps DeclarationTable) (errors Errors)
	GetType() TypeValue
	Lower() cpp.Expression
}

type DefaultExpression struct{}

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

// GetTypeDependencies implements Expression.
func (d DefaultExpression) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements Expression.
func (d DefaultExpression) GetValueDependencies() []Name {
	return []Name{}
}


type Integer struct {
	DefaultExpression
	Span  Span
	Type  TypeValue
	Value int64
}

// SetType implements Expression.
func (i *Integer) SetType(t TypeValue) {
	i.Type = t
}

// GetSpan implements Expression.
func (i Integer) GetSpan() Span {
	return i.Span
}

// Lower implements Expression.
func (i Integer) Lower() cpp.Expression {
	return fmt.Sprintf("%v", i.Value)
}

// GetType implements Expression.
func (i Integer) GetType() TypeValue {
	if i.Type == nil {
		return IntType{}
	} else {
		return i.Type
	}
}

// InferType implements Expression.
func (i Integer) InferType(deps DeclarationTable) (errors []error) {
	if i.Type != nil && !i.Type.Eq(IntType{}) && !i.Type.Eq(IntegerLiteralType) && !i.Type.Eq(ExpressionType) {
		errors = append(errors, UnexpectedType{
			Expected: i.Type,
			Found:    IntType{},
			At:       i.Span,
		})
	}
	return
}

type Float struct {
	DefaultExpression
	Span  Span
	Type  TypeValue
	Value float64
}

// SetType implements Expression.
func (f *Float) SetType(t TypeValue) {
	f.Type = t
}

// GetSpan implements Expression.
func (f Float) GetSpan() Span {
	return f.Span
}

// Lower implements Expression.
func (f Float) Lower() cpp.Expression {
	return fmt.Sprintf("%v", f.Value)
}

// GetType implements Expression.
func (f Float) GetType() TypeValue {
	if f.Type == nil {
		return FloatType{}
	} else {
		return f.Type
	}
}

// InferType implements Expression.
func (f Float) InferType(deps DeclarationTable) (errors []error) {
	if f.Type != nil && !f.Type.Eq(FloatType{}) && !f.Type.Eq(FloatLiteralType) && !f.Type.Eq(ExpressionType) {
		errors = append(errors, UnexpectedType{
			Expected: f.Type,
			Found:    FloatType{},
			At:       f.Span,
		})
	}
	return
}

type Bool struct {
	DefaultExpression
	Span  Span
	Type  TypeValue
	Value bool
}

// SetType implements Expression.
func (b *Bool) SetType(t TypeValue) {
	b.Type = t
}

// GetSpan implements Expression.
func (b Bool) GetSpan() Span {
	return b.Span
}

// Lower implements Expression.
func (b Bool) Lower() cpp.Expression {
	return fmt.Sprintf("%v", b.Value)
}

// GetType implements Expression.
func (b Bool) GetType() TypeValue {
	if b.Type == nil {
		return BoolType{}
	} else {
		return b.Type
	}
}

// InferType implements Expression.
func (b Bool) InferType(deps DeclarationTable) (errors []error) {
	if b.Type != nil && !b.Type.Eq(BoolType{}) && !b.Type.Eq(BoolLiteralType) && !b.Type.Eq(ExpressionType) {
		errors = append(errors, UnexpectedType{
			Expected: b.Type,
			Found:    BoolType{},
			At:       b.Span,
		})
	}
	return
}

type String struct {
	DefaultExpression
	Span  Span
	Type  TypeValue
	Value string
}

// SetType implements Expression.
func (s *String) SetType(t TypeValue) {
	s.Type = t
}

// GetSpan implements Expression.
func (s String) GetSpan() Span {
	return s.Span
}

// Lower implements Expression.
func (s String) Lower() cpp.Expression {
	bytes, _ := json.Marshal(s.Value)
	return string(bytes)
}

// GetType implements Expression.
func (s String) GetType() TypeValue {
	if s.Type == nil {
		return StringType{}
	} else {
		return s.Type
	}
}

// InferType implements Expression.
func (s String) InferType(deps DeclarationTable) (errors []error) {
	if s.Type != nil && !s.Type.Eq(StringType{}) && !s.Type.Eq(StringLiteralType) && !s.Type.Eq(ExpressionType) {
		errors = append(errors, UnexpectedType{
			Expected: s.Type,
			Found:    StringType{},
			At:       s.Span,
		})
	}
	return
}

type Variable struct {
	DefaultExpression
	Type TypeValue
	Name Name
}

// SetType implements Expression.
func (v *Variable) SetType(t TypeValue) {
	v.Type = t
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
func (v *Variable) InferType(deps DeclarationTable) (errors Errors) {
	decl, _ := deps.Get(v.Name.String)
	if decl.GetDeclaredType() == nil {
		log.Fatalf("Type queried at %s before being calculated on declaration '%s'.", v.Name.Span, v.Name.String)
	}
	if v.Type != nil && !v.Type.Eq(decl.GetDeclaredType()) {
		errors = append(errors, UnexpectedType{
			Expected: v.Type,
			Found:    decl.GetDeclaredType(),
			At:       v.GetSpan(),
		})
	}
	v.Type = decl.GetDeclaredType()
	return
}

// Lower implements Expression.
func (v *Variable) Lower() cpp.Expression {
	return v.Name.String
}

type FunctionCall struct {
	DefaultExpression
	Span     Span
	Type     TypeValue
	Function Expression
	Argument Expression
}

// SetType implements Expression.
func (f *FunctionCall) SetType(t TypeValue) {
	f.Type = t
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
func (f *FunctionCall) InferType(deps DeclarationTable) (errors Errors) {
	errors = f.Function.InferType(deps)
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
	f.Argument.SetType(functionType.Argument)
	if errors = f.Argument.InferType(deps); len(errors) > 0 {
		return
	}
	if f.Type != nil && !f.Type.Eq(functionType.Return) {
		errors = append(errors, UnexpectedType{
			Expected: f.Type,
			Found:    functionType.Return,
			At:       f.GetSpan(),
		})
	}
	f.Type = functionType.Return
	return
}

// Lower implements Expression.
func (f *FunctionCall) Lower() cpp.Expression {
	argumentType := f.Argument.GetType()
	_, isTuple := argumentType.(TupleType)
	if isTuple {
		// calls the function with a tuple of arguments
		return fmt.Sprintf(`std::apply(%s, %s)`, f.Function.Lower(), f.Argument.Lower())
	}
	return fmt.Sprintf(`%s(%s)`, f.Function.Lower(), f.Argument.Lower())
}

type Tuple struct {
	DefaultExpression
	Span Span
	// Inferred type
	Type     TypeValue
	Elements []Expression
}

// SetType implements Expression.
func (t *Tuple) SetType(tv TypeValue) {
	t.Type = tv
	if tv == nil || tv.Eq(TypeType{}) {
		for i := range t.Elements {
			t.Elements[i].SetType(tv)
		}
	}
	tupleType, ok := tv.(TupleType)
	if ok && len(tupleType.Elements) == len(t.Elements) {
		for i := range t.Elements {
			t.Elements[i].SetType(tupleType.Elements[i])
		}
	}
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
func (t *Tuple) InferType(deps DeclarationTable) (errors Errors) {
	expectedTupleType, isTuple := t.Type.(TupleType)

	if isTuple && len(expectedTupleType.Elements) != len(t.Elements) {
		errors = append(errors, ArityMismatch{
			Expected: len(expectedTupleType.Elements),
			Found:    len(t.Elements),
			At:       t.GetSpan(),
		})
		return
	}
	for i := range t.Elements {
		errors = append(errors, t.Elements[i].InferType(deps)...)
	}
	if len(errors) > 0 {
		return
	}
	if t.Type == nil || !t.Type.Eq(TypeType{}) {
		t.Type = NewTupleType(util.Map(t.Elements, func(e Expression) TypeValue {
			return e.GetType()
		})...)
	}
	return
}

// Lower implements Expression.
func (t *Tuple) Lower() cpp.Expression {
	if typeEqual(t.Type, TypeType{}) {
		if len(t.Elements) == 0 {
			return `box(ty::TupleType({}))`
		}
		elements := util.JoinFunction(t.Elements, ", ", func(e Expression) string {
			return e.Lower()
		})
		return fmt.Sprintf(`box(ty::TupleType({%s}))`, elements)
	} else {
		return fmt.Sprintf(`std::make_tuple(%s)`, util.JoinFunction(t.Elements, ", ", Expression.Lower))
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

// SetType implements Expression.
func (m *Macro) SetType(t TypeValue) {
	m.Result.SetType(t)
}

// SetValue implements Destination.
func (m *Macro) SetValue(json string) {
	util.PrettyPrint(json)
	v := fj.MustParse(json)
	elements, err := v.Get("Tuple", "elements").Array()
	if err != nil {
		log.Fatalf("Failed to parse macro output as Tuple. Output: %s", json)
	}
	errorMessage := string(elements[0].GetStringBytes())
	expression := UnmarshalExpression(elements[1])
	m.Result = expression
	if errorMessage != "" {
		panic("Macro returned error: " + errorMessage)
	}
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
		Argument: &String{
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
func (m *Macro) InferType(deps DeclarationTable) (errors Errors) {
	return m.Result.InferType(deps)
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

// TODO: allow "lowering" UnaryExpression, BinaryExpression and Tuple to a cpp ty::Expression
//     just like the literals

type UnaryExpression struct {
	DefaultExpression
	Span       Span
	Type       TypeValue
	Op         UnaryOp
	Expression Expression
}

// SetType implements Expression.
func (u *UnaryExpression) SetType(t TypeValue) {
	u.Type = t
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
func (u *UnaryExpression) InferType(deps DeclarationTable) (errors Errors) {
	errors = u.Expression.InferType(deps)
	if len(errors) > 0 {
		return
	}
	expressionType := u.Expression.GetType()
	switch {
	case
		typeEqual(expressionType, IntType{}),
		typeEqual(expressionType, FloatType{}):
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
		return "-" + u.Expression.Lower()
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

// SetType implements Expression.
func (b *BinaryExpression) SetType(t TypeValue) {
	b.Type = t
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
func (b *BinaryExpression) InferType(deps DeclarationTable) (errors Errors) {
	errors = append(b.Left.InferType(deps), b.Right.InferType(deps)...)
	if len(errors) > 0 {
		return
	}
	leftType := b.Left.GetType()
	rightType := b.Right.GetType()
	if !typeEqual(leftType, rightType) {
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
		if !typeEqual(leftType, IntType{}) && !typeEqual(leftType, FloatType{}) {
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
		if !typeEqual(leftType, BoolType{}) {
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
	return b.Left.Lower() + " " + op + " " + b.Right.Lower()
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

type StructExpression struct {
	DefaultExpression
	Span Span
	Name   string
	Fields map[string]Expression
}

// GetSpan implements Expression.
func (s *StructExpression) GetSpan() Span {
	return s.Span
}

func (s StructExpression) SetType(t TypeValue) {
	// TODO: check union conversibility
	if !(StructType{Name: s.Name}.Eq(t)) {
		panic("StructValue type does not match type provided to SetType")
	}
}

func (s StructExpression) InferType(deps DeclarationTable) (errors Errors) {
	return
}

func (s StructExpression) GetType() TypeValue {
	return StructType{Name: s.Name}
}

func (s StructExpression) Lower() cpp.Expression {
	fields := ""
	for key, value := range s.Fields {
		fields += fmt.Sprintf(".%s = %s,\n", key, value.Lower())
	}
	return fmt.Sprintf(`(%s){\n%s}`, s.Name, fields)
}

// Tries to unmarshal an Expression, returning nil if the union key does not match an Expression.
func UnmarshalExpression(data *fj.Value) (expr Expression) {
	object := data.GetObject()
	if object == nil { // primitive or array
		integer, err := data.Int64()
		if err == nil {
			expr = &Integer{Value: integer}
			return
		}
		float, err := data.Float64()
		if err == nil {
			expr = &Float{Value: float}
			return
		}
		boolean, err := data.Bool()
		if err == nil {
			expr = &Bool{Value: boolean}
			return
		}
		stringBytes, err := data.StringBytes()
		if err == nil {
			expr = &String{Value: string(stringBytes)}
			return
		}
		return nil // not a primitive
	}
	key, v := fjUnmarshalUnion(object)
	expr = &StructExpression {
		Span:              fjUnmarshal(v.Get("span"), Span{}),
		Name:              key,
			Fields:            panic("TODO map[string]Expression"),
	}
	
	switch key {
	case "IntegerLiteral":
		expr = &Integer{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Value: v.GetInt64("value"),
		}
	case "FloatLiteral":
		expr = &Float{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Value: v.GetFloat64("value"),
		}
	case "BoolLiteral":
		expr = &Bool{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Value: v.GetBool("value"),
		}
	case "StringLiteral":
		expr = &String{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Value: string(v.GetStringBytes("value")),
		}
	case "Variable":
		expr = &Variable{
			Name: Name{
				Span:   fjUnmarshal(v.Get("span"), Span{}),
				String: string(v.GetStringBytes("value")),
			},
			Type: UnmarshalType(v.Get("type")),
		}
	case "FunctionCall":
		expr = &FunctionCall{
			Span:     fjUnmarshal(v.Get("span"), Span{}),
			Type:     UnmarshalType(v.Get("type")),
			Function: UnmarshalExpression(v.Get("function")),
			Argument: UnmarshalExpression(v.Get("argument")),
		}
	case "Tuple":
		expr = &Tuple{
			Span:     fjUnmarshal(v.Get("span"), Span{}),
			Elements: util.Map(v.Get("elements").GetArray(), UnmarshalExpression),
		}
	case "Macro":
		panic("Macros are not supported for serialization right now.")
	case "UnaryExpression":
		expr = &UnaryExpression{
			Span:       fjUnmarshal(v.Get("span"), Span{}),
			Op:         UnaryOp(v.GetStringBytes("op")),
			Expression: UnmarshalExpression(v.Get("expression")),
		}
	case "BinaryExpression":
		expr = &BinaryExpression{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Op:    BinaryOp(v.GetStringBytes("op")),
			Left:  UnmarshalExpression(v.Get("left")),
			Right: UnmarshalExpression(v.Get("right")),
		}
	default: // unknown key
		// expr = nil
	}
	return
}

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
var _ Expression = &StructExpression{}
