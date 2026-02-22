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
	fmt.Stringer
	Analyze(expected TypeValue, anal Analyzer) TypeValue
	Lower() cpp.Expression
}

type Integer struct {
	Span  Span
	Value int64
}

func (i Integer) String() string {
	return fmt.Sprintf("%v", i.Value)
}

// GetSpan implements Expression.
func (i Integer) GetSpan() Span {
	return i.Span
}

// Lower implements Expression.
func (i Integer) Lower() cpp.Expression {
	return fmt.Sprintf("%v", i.Value)
}

// Analyze implements Expression.
func (i Integer) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return &IntType{}
}

type Float struct {
	Span  Span
	Value float64
}

func (f Float) String() string {
	return fmt.Sprintf("%v", f.Value)
}

// GetSpan implements Expression.
func (f Float) GetSpan() Span {
	return f.Span
}

// Lower implements Expression.
func (f Float) Lower() cpp.Expression {
	return fmt.Sprintf("%v", f.Value)
}

// Analyze implements Expression.
func (f Float) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return &FloatType{}
}

type Bool struct {
	Span  Span
	Value bool
}

func (b Bool) String() string {
	return fmt.Sprintf("%v", b.Value)
}

// GetSpan implements Expression.
func (b Bool) GetSpan() Span {
	return b.Span
}

// Lower implements Expression.
func (b Bool) Lower() cpp.Expression {
	return fmt.Sprintf("%v", b.Value)
}

// Analyze implements Expression.
func (b Bool) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return &BoolType{}
}

type String struct {
	Span  Span
	Value string
}

func (s String) String() string {
	return fmt.Sprintf("%v", s.Value)
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

// Analyze implements Expression.
func (s String) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return &StringType{}
}

type Variable struct {
	Name Name
}

func (v Variable) String() string {
	return fmt.Sprintf("%v", v.Name.String)
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
func (v *Variable) Lower() cpp.Expression {
	return v.Name.Lower()
}

type FunctionCall struct {
	Span            Span
	Function        Expression
	Argument        Expression
	ArgumentIsTuple bool
}

func (f FunctionCall) String() string {
	_, argumentIsTuple := f.Argument.(*Tuple)
	if argumentIsTuple {
		return fmt.Sprintf("%s%s", f.Function, f.Argument)
	} else {
		return fmt.Sprintf("%s %s", f.Function, f.Argument)
	}
}

// GetSpan implements Expression.
func (f *FunctionCall) GetSpan() Span {
	return f.Span
}

// Analyze implements Expression.
func (f *FunctionCall) Analyze(expected TypeValue, anal Analyzer) (returnType TypeValue) {
	maybeFunctionType := f.Function.Analyze(nil, anal)
	functionType, isFunction := maybeFunctionType.(*FnType)
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
	_, argumentIsTuple := argumentType.(*TupleType)
	f.ArgumentIsTuple = argumentIsTuple
	return
}

// Lower implements Expression.
func (f *FunctionCall) Lower() cpp.Expression {
	if f.ArgumentIsTuple {
		// calls the function with a tuple of arguments
		return fmt.Sprintf(`std::apply(%s, %s)`, f.Function.Lower(), f.Argument.Lower())
	}
	return fmt.Sprintf(`%s(%s)`, f.Function.Lower(), f.Argument.Lower())
}

type Tuple struct {
	Span     Span
	IsType   bool
	Elements []Expression
}

func (t Tuple) String() string {
	s := "("
	for _, element := range t.Elements {
		s += element.String() + ", "
	}
	s += ")"
	return s
}

// GetSpan implements Expression.
func (t *Tuple) GetSpan() Span {
	return t.Span
}

// Analyze implements Expression.
func (t *Tuple) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	expectedTupleType, isTuple := expected.(*TupleType)
	_type := &TupleType{}

	if isTuple && len(expectedTupleType.Elements) != len(t.Elements) {
		anal.PushError(ArityMismatch{
			Expected: len(expectedTupleType.Elements),
			Found:    len(t.Elements),
			At:       t.GetSpan(),
		})
	}
	for i := range t.Elements {
		expectedElementType := expected
		if isTuple && len(expectedTupleType.Elements) >= i {
			expectedElementType = expectedTupleType.Elements[i]
		} else if !expected.Eq(&TypeType{}) {
			expectedElementType = &TypeType{}
		}
		elementType := t.Elements[i].Analyze(expectedElementType, anal)
		_type.Elements = append(_type.Elements, elementType)
	}
	t.IsType = expected != nil && expected.Eq(&TypeType{})
	if t.IsType {
		return &TypeType{}
	}
	return _type
}

// Lower implements Expression.
func (t *Tuple) Lower() cpp.Expression {
	if t.IsType {
		if len(t.Elements) == 0 {
			return `box((ty::TupleType){})`
		}
		elements := util.JoinFunction(t.Elements, ", ", func(e Expression) string {
			return e.Lower()
		})
		return fmt.Sprintf(`box((ty::TupleType){ .elements = {%s} })`, elements)
	} else {
		return fmt.Sprintf(`std::make_tuple(%s)`, util.JoinFunction(t.Elements, ", ", func(e Expression) cpp.Expression {
			return e.Lower()
		}))
	}
}

type Macro struct {
	Span Span
	// Function that evaluates the macro.
	Function Variable
	Lines    []MacroLine
	// Result after evaluating the macro.
	Result Expression
}

func (m Macro) String() string {
	s := m.Function.String() + "#"
	// NOTE: missing indentation
	for _, line := range m.Lines {
		s += line.Text + "\n"
	}
	return s
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
	if !functionType.Eq(MacroFunctionType) {
		anal.PushError(UnexpectedType{
			Expected: MacroFunctionType,
			Found:    functionType,
			At:       m.Function.GetSpan(),
		})
	}
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
	if errorMessage != "" {
		panic("Macro returned error: " + errorMessage)
	}
	// unmarshalled closure should already be analyzed, since it is referenced
	// by the analyzed macro function
	closure := lowerClosureValue(elements[1].Get("Closure"))
	json = anal.EvaluateLowered(closure + "()")
	m.Result = UnmarshalExpression(fj.MustParse(json))
	return m.Result.Analyze(expected, anal)
}

// Lower implements Expression.
func (m *Macro) Lower() cpp.Expression {
	return m.Result.Lower()
}

type MacroLine struct {
	Span
	Text string
}

type UnaryExpression struct {
	Span       Span
	Op         UnaryOp
	Expression Expression
}

func (u UnaryExpression) String() string {
	return string(u.Op) + u.Expression.String()
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
		expressionType.Eq(&IntType{}),
		expressionType.Eq(&FloatType{}):
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
	Span  Span
	Op    BinaryOp
	Left  Expression
	Right Expression
}

func (b BinaryExpression) String() string {
	return b.Left.String() + " " + string(b.Op) + " " + b.Right.String()
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
		if !leftType.Eq(&IntType{}) && !leftType.Eq(&FloatType{}) {
			emitErr()
		}
		return leftType
	case
		Equal,
		NotEqual:
		return &BoolType{}
	case
		Or, And:
		if !leftType.Eq(&BoolType{}) {
			emitErr()
		}
		return &BoolType{}
	default:
		panic(fmt.Sprintf("unexpected ast.BinaryOp: %#v", b.Op))
	}
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
	Span   Span
	Name   string
	Fields map[string]Expression
}

func (s StructExpression) String() string {
	return s.Name + "<fields>"
}

// GetSpan implements Expression.
func (s *StructExpression) GetSpan() Span {
	return s.Span
}

func (s StructExpression) GetId() {
}

func (s StructExpression) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	panic("unimplemented")
}

func (s StructExpression) Lower() cpp.Expression {
	fields := ""
	for key, value := range s.Fields {
		fields += fmt.Sprintf(".%s = %s,\n", key, value.Lower())
	}
	return fmt.Sprintf(`(%s){\n%s}`, s.Name, fields)
}

type Closure struct {
	Span       Span
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
	captures   map[string]TypeValue
}

func (c Closure) String() string {
	return "<closure>"
}

// GetId implements Expression.
func (c *Closure) GetId() string {
	// NOTE: this requires unique Span for C++-generated Closure definitions
	return fmt.Sprintf("closure_%d_%d", c.Span.Line, c.Span.Column)
}

// GetSpan implements Expression.
func (c *Closure) GetSpan() Span {
	return c.Span
}

// Analyze implements Expression.
func (c *Closure) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	if c.captures != nil {
		panic("Closure analyzed multiple times. Expressions should only be analyzed once.")
	}
	c.captures = map[string]TypeValue{} // prevents nil dereference error when adding to map
	anal = anal.NewScope()
	analyzeFunctionHeader(anal, c.Parameters, &c.ReturnType)
	analyzeFunctionBody(anal, c.ReturnType.Get(), c.Body)
	// FIXME: this should not capture the types used in the closure's signature
	for _, capture := range *anal.Table.captures {
		c.captures[capture.name.String] = capture.declaration.GetDeclaredType()
	}
	return getFunctionType(c.Parameters, c.ReturnType)
}

func (c Closure) LowerParameters() string {
	return util.JoinFunction(c.Parameters, ", ", FunctionParameter.Lower)
}

// Lower implements Expression.
func (c *Closure) Lower() cpp.Expression {
	id := registerNode(c)
	fields := ""
	captureArguments := ""
	captures := ""
	for captureName, captureType := range c.captures {
		fields += captureType.Lower() + " " + captureName + ";\n"
		if captureArguments != "" {
			captureArguments += ", "
			captures += ` + ", " + `
		}
		captureArguments += captureName
		// not using newlines because these are automatically escaped by the evaluator
		// which results in malformed JSON
		captures += fmt.Sprintf(
			`ty::serialize_capture(%q, %q, %s)`,
			captureName,
			registerNode(TypeId{captureType}),
			captureName)
	}
	// C++ requires that closures that do not capture anything do not have a default capture symbol
	lambdaSymbol := "" // infer capture
	if len(c.captures) > 0 {
		lambdaSymbol = "=" // capture by value
	}
	// declares the struct and immediately captures the right variables from the environment
	lowered := fmt.Sprintf(`[%s](){
    struct {
        %s operator()(%s) const %s
        std::string serialize() const {
            return ty::serialize_closure(%s, %q);
        }
        %s
    } closure{%s};
    return closure;
}()`,
		lambdaSymbol,
		c.ReturnType.Lower(),
		c.LowerParameters(),
		cpp.Block(c.Body.Lower()),
		captures,
		id,
		fields,
		captureArguments)
	return lowered
}

// Tries to unmarshal an Expression, panicking if the union key does not match an Expression.
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
	case "ClosureExpression":
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
	case "Box": // boxing is irrelevant when unmarshalling
		expr = UnmarshalExpression(v)
	default:
		panic(fmt.Sprintf("unexpected expression key when unmarshalling: %s", key))
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
