package ast

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"yune/cpp"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

type Expression interface {
	Node
	StatementBase
	GetMacros() []*Macro

	SetId(t TypeValue)
	// `expected` indicates the expected type, if any, in order to resolve ambiguities
	// when the expression needs to be inferred differently from the default.
	// This is the case when an Expression is used in a Type, or represents syntax
	// the user has generated.
	InferType(expected TypeValue, deps DeclarationTable) (errors Errors)
	GetType() TypeValue
	Lower(defs *[]cpp.Definition) cpp.Expression
}

type DefaultExpression struct{}

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

// SetId implements Expression.
func (i *Integer) SetId(t TypeValue) {
	i.Type = t
}

// GetSpan implements Expression.
func (i Integer) GetSpan() Span {
	return i.Span
}

// Lower implements Expression.
func (i Integer) Lower(defs *[]cpp.Definition) cpp.Expression {
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
func (i Integer) InferType(expected TypeValue, deps DeclarationTable) (errors []error) {
	if i.Type != nil && !i.Type.Eq(IntType{}) {
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

// SetId implements Expression.
func (f *Float) SetId(t TypeValue) {
	f.Type = t
}

// GetSpan implements Expression.
func (f Float) GetSpan() Span {
	return f.Span
}

// Lower implements Expression.
func (f Float) Lower(defs *[]cpp.Definition) cpp.Expression {
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
func (f Float) InferType(expected TypeValue, deps DeclarationTable) (errors []error) {
	if f.Type != nil && !f.Type.Eq(FloatType{}) {
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

// SetId implements Expression.
func (b *Bool) SetId(t TypeValue) {
	b.Type = t
}

// GetSpan implements Expression.
func (b Bool) GetSpan() Span {
	return b.Span
}

// Lower implements Expression.
func (b Bool) Lower(defs *[]cpp.Definition) cpp.Expression {
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
func (b Bool) InferType(expected TypeValue, deps DeclarationTable) (errors []error) {
	if b.Type != nil && !b.Type.Eq(BoolType{}) {
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

// SetId implements Expression.
func (s *String) SetId(t TypeValue) {
	s.Type = t
}

// GetSpan implements Expression.
func (s String) GetSpan() Span {
	return s.Span
}

// Lower implements Expression.
func (s String) Lower(defs *[]cpp.Definition) cpp.Expression {
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
func (s String) InferType(expected TypeValue, deps DeclarationTable) (errors []error) {
	if s.Type != nil && !s.Type.Eq(StringType{}) {
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

// SetId implements Expression.
func (v *Variable) SetId(t TypeValue) {
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
func (v *Variable) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
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
func (v *Variable) Lower(defs *[]cpp.Definition) cpp.Expression {
	return v.Name.Lower()
}

type FunctionCall struct {
	DefaultExpression
	Span     Span
	Function Expression
	Argument Expression
}

// SetId implements Expression.
func (f *FunctionCall) SetId(t TypeValue) {
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
	return f.Function.GetType().(FnType).Return
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
	errors = f.Function.InferType(nil, deps)
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
	return
}

// Lower implements Expression.
func (f *FunctionCall) Lower(defs *[]cpp.Definition) cpp.Expression {
	argumentType := f.Argument.GetType()
	_, isTuple := argumentType.(TupleType)
	if isTuple {
		// calls the function with a tuple of arguments
		return fmt.Sprintf(`std::apply(%s, %s)`, f.Function.Lower(defs), f.Argument.Lower(defs))
	}
	return fmt.Sprintf(`%s(%s)`, f.Function.Lower(defs), f.Argument.Lower(defs))
}

type Tuple struct {
	DefaultExpression
	Span     Span
	IsType   bool
	Elements []Expression
}

// SetId implements Expression.
func (t *Tuple) SetId(tv TypeValue) {
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

	if isTuple && len(expectedTupleType.Elements) != len(t.Elements) {
		errors = append(errors, ArityMismatch{
			Expected: len(expectedTupleType.Elements),
			Found:    len(t.Elements),
			At:       t.GetSpan(),
		})
		return
	}
	for i := range t.Elements {
		var expected TypeValue
		if isTuple {
			expected = expectedTupleType.Elements[i]
		} else if typeEqual(expected, TypeType{}) {
			expected = TypeType{}
		}
		errors = append(errors, t.Elements[i].InferType(expected, deps)...)
	}
	if len(errors) > 0 {
		return
	}
	t.IsType = typeEqual(expected, TypeType{})
	return
}

// Lower implements Expression.
func (t *Tuple) Lower(defs *[]cpp.Definition) cpp.Expression {
	if t.IsType {
		if len(t.Elements) == 0 {
			return `box((ty::TupleType){})`
		}
		elements := util.JoinFunction(t.Elements, ", ", func(e Expression) string {
			return e.Lower(defs)
		})
		return fmt.Sprintf(`box((ty::TupleType){ .elements = {%s} })`, elements)
	} else {
		return fmt.Sprintf(`std::make_tuple(%s)`, util.JoinFunction(t.Elements, ", ", func(e Expression) cpp.Expression {
			return e.Lower(defs)
		}))
	}
}

// GetType implements Expression.
func (t *Tuple) GetType() TypeValue {
	if t.IsType {
		return TypeType{}
	}
	return NewTupleType(util.Map(t.Elements, func(e Expression) TypeValue {
		return e.GetType()
	})...)
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

// SetId implements Expression.
func (m *Macro) SetId(t TypeValue) {
	m.Result.SetId(t)
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
func (m *Macro) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	return m.Result.InferType(expected, deps)
}

// Lower implements Expression.
func (m *Macro) Lower(defs *[]cpp.Definition) cpp.Expression {
	return m.Result.Lower(defs)
}

// GetType implements Expression.
func (m *Macro) GetType() TypeValue {
	return m.Result.GetType()
}

type MacroLine struct {
	Span
	Text string
}

type MacroQuery struct {
	FunctionCall
	Macro *Macro
}

var _ Query = (*MacroQuery)(nil)

func NewMacroQuery(m *Macro) *MacroQuery {
	return &MacroQuery{
		FunctionCall: m.AsFunctionCall(),
		Macro:        m,
	}
}

// CheckType implements Query
func (m MacroQuery) CheckType(deps DeclarationTable) (errors Errors) {
	errors = m.Macro.Function.InferType(MacroFunctionType, deps)
	if len(errors) > 0 {
		return
	}
	functionType := m.Macro.Function.GetType()
	if !functionType.Eq(MacroFunctionType) {
		// TODO: custom "invalid macro function type" error
		errors = append(errors, UnexpectedType{
			Expected: MacroFunctionType,
			Found:    functionType,
			At:       m.Macro.Function.GetSpan(),
		})
	}
	return
}

// SetValue implements Query
func (m MacroQuery) SetValue(json string) {
	v := fj.MustParse(json)
	elements := v.GetArray("Tuple", "elements")
	if elements == nil {
		log.Fatalf("Failed to parse macro output as Tuple. Output: %s", json)
	}
	errorMessage := string(elements[0].GetStringBytes())
	expression := UnmarshalExpression(elements[1])
	m.Macro.Result = expression
	if errorMessage != "" {
		panic("Macro returned error: " + errorMessage)
	}
}

type UnaryExpression struct {
	DefaultExpression
	Span       Span
	Type       TypeValue
	Op         UnaryOp
	Expression Expression
}

// SetId implements Expression.
func (u *UnaryExpression) SetId(t TypeValue) {
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
func (u *UnaryExpression) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	errors = u.Expression.InferType(nil, deps)
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
func (u *UnaryExpression) Lower(defs *[]cpp.Definition) cpp.Expression {
	switch u.Op {
	case Negate:
		return "-" + u.Expression.Lower(defs)
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

// SetId implements Expression.
func (b *BinaryExpression) SetId(t TypeValue) {
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
func (b *BinaryExpression) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	// TODO: the expected type (used by InferType) for Left and Right differs depending on the operator
	errors = append(b.Left.InferType(nil, deps), b.Right.InferType(nil, deps)...)
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
func (b *BinaryExpression) Lower(defs *[]cpp.Definition) cpp.Expression {
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
	return b.Left.Lower(defs) + " " + op + " " + b.Right.Lower(defs)
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
	Span   Span
	Name   string
	Fields map[string]Expression
}

// GetSpan implements Expression.
func (s *StructExpression) GetSpan() Span {
	return s.Span
}

func (s StructExpression) SetId(t TypeValue) {
	// TODO: check union conversibility
	if !(StructType{Name: s.Name}.Eq(t)) {
		panic("StructValue type does not match type provided to SetId")
	}
}

func (s StructExpression) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	return
}

func (s StructExpression) GetType() TypeValue {
	return StructType{Name: s.Name}
}

func (s StructExpression) Lower(defs *[]cpp.Definition) cpp.Expression {
	fields := ""
	for key, value := range s.Fields {
		fields += fmt.Sprintf(".%s = %s,\n", key, value.Lower(defs))
	}
	return fmt.Sprintf(`(%s){\n%s}`, s.Name, fields)
}

type Closure struct {
	DefaultExpression
	Span       Span
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
	captures   map[string]TypeValue
}

// SetId implements Expression.
func (c *Closure) SetId(t TypeValue) {
	println("TODO: Closure.SetId")
}

// GetSpan implements Expression.
func (c *Closure) GetSpan() Span {
	return c.Span
}

// GetMacros implements Expression.
func (c *Closure) GetMacros() (macros []*Macro) {
	macros = util.FlatMap(c.Parameters, FunctionParameter.GetMacros)
	macros = append(macros, c.Body.GetMacros()...)
	return
}

// GetMacroTypeDependencies implements Expression.
func (c *Closure) GetMacroTypeDependencies() (deps []Query) {
	deps = util.FlatMapPtr(c.Parameters, (*FunctionParameter).GetMacroTypeDependencies)
	deps = append(deps, c.Body.GetMacroTypeDependencies()...)
	return
}

// GetTypeDependencies implements Expression.
func (c *Closure) GetTypeDependencies() (deps []Query) {
	return getFunctionTypeDependencies(c.Parameters, &c.ReturnType, c.Body)
}

// GetType implements Expression.
func (c *Closure) GetType() TypeValue {
	return getFunctionType(c.Parameters, c.ReturnType)
}

// GetMacroValueDependencies implements Expression.
func (c *Closure) GetMacroValueDependencies() (deps []Name) {
	for _, depName := range c.Body.GetMacroValueDependencies() {
		equals := func(param FunctionParameter) bool {
			return depName.String == param.GetName().String
		}
		if !util.Any(equals, c.Parameters...) {
			deps = append(deps, depName)
		}
	}
	return
}

// GetValueDependencies implements Expression.
func (c *Closure) GetValueDependencies() (deps []Name) {
	// make non-nil to prevent nil-dereference error when adding elements
	c.captures = map[string]TypeValue{}
	for _, depName := range c.Body.GetValueDependencies() {
		equals := func(param FunctionParameter) bool {
			return depName.String == param.GetName().String
		}
		if !util.Any(equals, c.Parameters...) {
			deps = append(deps, depName)
			c.captures[depName.String] = nil // type is not known yet
		}
	}
	return
}

// InferType implements Expression.
func (c *Closure) InferType(expected TypeValue, deps DeclarationTable) (errors Errors) {
	for name := range c.captures {
		declaration, ok := deps.Get(name)
		if !ok {
			log.Fatalf("Declaration table does not contain closure capture '%s'", name)
		}
		c.captures[name] = declaration.GetDeclaredType()
	}
	return typeCheckFunction(nil, c.Parameters, c.ReturnType, c.Body, deps)
}

// Lower implements Expression.
func (c *Closure) Lower(defs *[]cpp.Definition) cpp.Expression {
	// TODO: fully prevent naming conflicts instead of using rand
	name := fmt.Sprintf("closure_%x_", rand.Uint64())
	if c.captures == nil {
		panic("Closure.Lower called without callng GetValueDependencies first.")
	}
	fields := ""
	captures := ""
	for captureName, captureType := range c.captures {
		fields += captureType.Lower() + " " + captureName + ";\n"
		if captures != "" {
			captures += ", "
		}
		captures += captureName
	}
	// declares the struct and immediately captures the right variables from the environment
	definition := fmt.Sprintf(`struct {
    %s operator()(%s) const {
        %s
    }
    %s
} %s{%s};`,
		c.ReturnType.Lower(),
		util.JoinFunction(c.Parameters, ", ", FunctionParameter.Lower),
		strings.Join(c.Body.Lower(), "\n"),
		fields,
		name, captures)
	*defs = append(*defs, definition)
	return name
}

// Tries to unmarshal an Expression, returning nil if the union key does not match an Expression.
func UnmarshalExpression(data *fj.Value) (expr Expression) {
	object := data.GetObject()
	key, v := fjUnmarshalUnion(object)
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
		}
	case "FunctionCall":
		expr = &FunctionCall{
			Span:     fjUnmarshal(v.Get("span"), Span{}),
			Function: UnmarshalExpression(v.Get("function")),
			Argument: UnmarshalExpression(v.Get("argument")),
		}
	case "TupleExpression":
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
	case "StructExpression":
		panic("TODO: unmarshal StructExpression")
	case "Closure":
		expr = &Closure{
			Span: fjUnmarshal(v.Get("span"), Span{}),
			Parameters: util.Map(v.GetArray("parameters"), func(v *fj.Value) FunctionParameter {
				return FunctionParameter{
					Span: fjUnmarshal(v.Get("span"), Span{}),
					Name: Name{
						Span:   Span{},
						String: string(v.GetStringBytes("name")),
					},
					Type: UnmarshalType(v.Get("type")),
				}
			}),
			ReturnType: UnmarshalType(v.Get("returnType")),
			Body:       UnmarshalBlock(v.Get("body")),
		}
	default:
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
var _ Expression = &Closure{}
