package ast

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"yune/cpp"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

type Expression interface {
	Node
	SetId()
	Analyze(expected TypeValue, anal Analyzer) TypeValue
	Lower(defs *[]cpp.Definition) cpp.Expression
}

type DefaultExpression struct{}

type Integer struct {
	DefaultExpression
	Span  Span
	Value int64
}

// SetId implements Expression.
func (i *Integer) SetId() {
}

// GetSpan implements Expression.
func (i Integer) GetSpan() Span {
	return i.Span
}

// Lower implements Expression.
func (i Integer) Lower(defs *[]cpp.Definition) cpp.Expression {
	return fmt.Sprintf("%v", i.Value)
}

// Analyze implements Expression.
func (i Integer) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return IntType{}
}

type Float struct {
	DefaultExpression
	Span  Span
	Value float64
}

// SetId implements Expression.
func (f *Float) SetId() {
}

// GetSpan implements Expression.
func (f Float) GetSpan() Span {
	return f.Span
}

// Lower implements Expression.
func (f Float) Lower(defs *[]cpp.Definition) cpp.Expression {
	return fmt.Sprintf("%v", f.Value)
}

// Analyze implements Expression.
func (f Float) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return FloatType{}
}

type Bool struct {
	DefaultExpression
	Span  Span
	Value bool
}

// SetId implements Expression.
func (b *Bool) SetId() {
}

// GetSpan implements Expression.
func (b Bool) GetSpan() Span {
	return b.Span
}

// Lower implements Expression.
func (b Bool) Lower(defs *[]cpp.Definition) cpp.Expression {
	return fmt.Sprintf("%v", b.Value)
}

// Analyze implements Expression.
func (b Bool) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return BoolType{}
}

type String struct {
	DefaultExpression
	Span  Span
	Value string
}

// SetId implements Expression.
func (s *String) SetId() {
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

// Analyze implements Expression.
func (s String) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return StringType{}
}

type Variable struct {
	DefaultExpression
	Name Name
}

// SetId implements Expression.
func (v *Variable) SetId() {
}

// GetSpan implements Expression.
func (v *Variable) GetSpan() Span {
	return v.Name.Span
}

// Analyze implements Expression.
func (v *Variable) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return anal.GetType(v.Name)
}

// Lower implements Expression.
func (v *Variable) Lower(defs *[]cpp.Definition) cpp.Expression {
	return v.Name.Lower()
}

type FunctionCall struct {
	DefaultExpression
	Span            Span
	Function        Expression
	Argument        Expression
	ArgumentIsTuple bool
}

// SetId implements Expression.
func (f *FunctionCall) SetId() {
}

// GetSpan implements Expression.
func (f *FunctionCall) GetSpan() Span {
	return f.Span
}

// Analyze implements Expression.
func (f *FunctionCall) Analyze(expected TypeValue, anal Analyzer) (returnType TypeValue) {
	maybeFunctionType := f.Function.Analyze(nil, anal)
	functionType, isFunction := maybeFunctionType.(FnType)
	if !isFunction {
		anal.PushError(NotAFunction{
			Found: maybeFunctionType,
			At:    f.Function.GetSpan(),
		})
	}
	var expectedArgumentType TypeValue
	if isFunction {
		expectedArgumentType = functionType.Argument
		returnType = functionType.Return
	}
	argumentType := f.Argument.Analyze(expectedArgumentType, anal)
	if !argumentType.Eq(expectedArgumentType) {
		anal.PushError(UnexpectedType{
			Expected: expectedArgumentType,
			Found:    argumentType,
			At:       f.Argument.GetSpan(),
		})
	}
	_, argumentIsTuple := argumentType.(TupleType)
	f.ArgumentIsTuple = argumentIsTuple
	return
}

// Lower implements Expression.
func (f *FunctionCall) Lower(defs *[]cpp.Definition) cpp.Expression {
	if f.ArgumentIsTuple {
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
func (t *Tuple) SetId() {
}

// GetSpan implements Expression.
func (t *Tuple) GetSpan() Span {
	return t.Span
}

// Analyze implements Expression.
func (t *Tuple) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	expectedTupleType, isTuple := expected.(TupleType)
	_type := TupleType{}

	if isTuple && len(expectedTupleType.Elements) != len(t.Elements) {
		anal.PushError(ArityMismatch{
			Expected: len(expectedTupleType.Elements),
			Found:    len(t.Elements),
			At:       t.GetSpan(),
		})
	}
	for i := range t.Elements {
		var expected TypeValue
		if isTuple && len(expectedTupleType.Elements) >= i {
			expected = expectedTupleType.Elements[i]
		} else if expected != nil && !expected.Eq(TypeType{}) {
			expected = TypeType{}
		}
		elementType := t.Elements[i].Analyze(expected, anal)
		_type.Elements = append(_type.Elements, elementType)
	}
	t.IsType = expected != nil && expected.Eq(TypeType{})
	if t.IsType {
		return TypeType{}
	}
	return _type
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
func (m *Macro) SetId() {
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

// Analyze implements Expression.
func (m *Macro) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	functionType := m.Function.Analyze(MacroFunctionType, anal)
	anal.PushError(UnexpectedType{
		Expected: MacroFunctionType,
		Found:    functionType,
		At:       m.Function.GetSpan(),
	})
	macroFunctionCall := FunctionCall{
		Span:     m.Span,
		Function: &m.Function,
		Argument: &String{
			Span:  m.Lines[0].Span,
			Value: m.GetText(),
		},
	}
	json := anal.Evaluate(&macroFunctionCall)
	v := fj.MustParse(json)
	elements := v.GetArray("Tuple", "elements")
	if elements == nil {
		log.Fatalf("Failed to parse macro output as Tuple. Output: %s", json)
	}
	errorMessage := string(elements[0].GetStringBytes())
	expression := UnmarshalExpression(elements[1])
	m.Result = expression
	if errorMessage != "" {
		panic("Macro returned error: " + errorMessage)
	}
	return m.Result.Analyze(expected, anal)
}

// Lower implements Expression.
func (m *Macro) Lower(defs *[]cpp.Definition) cpp.Expression {
	return m.Result.Lower(defs)
}

type MacroLine struct {
	Span
	Text string
}

type UnaryExpression struct {
	DefaultExpression
	Span       Span
	Op         UnaryOp
	Expression Expression
}

// SetId implements Expression.
func (u *UnaryExpression) SetId() {
}

// GetSpan implements Expression.
func (u *UnaryExpression) GetSpan() Span {
	return u.Span
}

// Analyze implements Expression.
func (u *UnaryExpression) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	expressionType := u.Expression.Analyze(nil, anal)
	switch {
	case
		expressionType.Eq(IntType{}),
		expressionType.Eq(FloatType{}):
		break
	default:
		anal.PushError(InvalidUnaryExpressionType{
			Op:   u.Op,
			Type: expressionType,
			At:   u.Span,
		})
	}
	return expressionType
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
	Op    BinaryOp
	Left  Expression
	Right Expression
}

// SetId implements Expression.
func (b *BinaryExpression) SetId() {
}

// GetSpan implements Expression.
func (b *BinaryExpression) GetSpan() Span {
	return b.Span
}

// Analyze implements Expression.
func (b *BinaryExpression) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	// TODO: the expected type (used by Analyze) for Left and Right differs depending on the operator
	leftType := b.Left.Analyze(nil, anal)
	rightType := b.Right.Analyze(nil, anal)
	if !leftType.Eq(rightType) {
		anal.PushError(InvalidBinaryExpressionTypes{
			Op:    b.Op,
			Left:  leftType,
			Right: rightType,
			At:    b.Span,
		})
	}
	emitErr := func() {
		anal.PushError(InvalidBinaryExpressionTypes{
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
		}
		return leftType
	case
		Equal,
		NotEqual:
		return BoolType{}
	case
		Or, And:
		if !leftType.Eq(BoolType{}) {
			emitErr()
		}
		return BoolType{}
	default:
		panic(fmt.Sprintf("unexpected ast.BinaryOp: %#v", b.Op))
	}
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

func (s StructExpression) SetId() {
}

func (s StructExpression) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	panic("unimplemented")
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
func (c *Closure) SetId() {
}

// GetSpan implements Expression.
func (c *Closure) GetSpan() Span {
	return c.Span
}

// Analyze implements Expression.
func (c *Closure) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	_type, captures := analyzeFunction(anal, nil, c.Parameters, &c.ReturnType, c.Body)
	for _, name := range captures {
		c.captures[name.String] = anal.GetType(name)
	}
	return _type
}

// Lower implements Expression.
func (c *Closure) Lower(defs *[]cpp.Definition) cpp.Expression {
	// TODO: fully prevent naming conflicts instead of using rand
	id := registerNode(c)
	name := fmt.Sprintf("closure_%x_", id)
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
    %s operator()(%s) const %s
    std::string serialize() const {
        return R"({ "ClosureId": "%d" })";
    }
    %s
} %s{%s};`,
		c.ReturnType.Lower(),
		util.JoinFunction(c.Parameters, ", ", FunctionParameter.Lower),
		cpp.Block(c.Body.Lower()),
		id,
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
	case "ClosureId":
		base := 10
		id, err := strconv.ParseUint(string(v.GetStringBytes()), base, 64)
		if err != nil {
			log.Fatalf("Failed to parse closure id. JSON: %s. Error: %s", v, err)
		}
		expr = registeredNodes[id].(Expression)
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
