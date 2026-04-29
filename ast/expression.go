package ast

import (
	"fmt"
	"strings"
	"yune/cpp"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

type Flags uint

const (
	IMPURE Flags = 1 << iota
	IMPURE_FUNCTION
)

type Expression interface {
	Node
	fmt.Stringer
	Analyze(expected TypeValue, anal Analyzer) TypeValue
	GetFlags() Flags
	Lower(state *State) cpp.Expression
}

type Integer struct {
	Span  Span
	Value int64
}

func (i *Integer) String() string {
	return fmt.Sprintf("%v", i.Value)
}

// GetSpan implements Expression.
func (i Integer) GetSpan() Span {
	return i.Span
}

// Lower implements Expression.
func (i Integer) Lower(state *State) cpp.Expression {
	return fmt.Sprintf("%v", i.Value)
}

// Analyze implements Expression.
func (i Integer) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return &IntType{}
}

func (Integer) GetFlags() Flags { return 0 }

type Float struct {
	Span  Span
	Value float64
}

func (f *Float) String() string {
	return fmt.Sprintf("%v", f.Value)
}

// GetSpan implements Expression.
func (f Float) GetSpan() Span {
	return f.Span
}

// Lower implements Expression.
func (f Float) Lower(state *State) cpp.Expression {
	return fmt.Sprintf("%v", f.Value)
}

// Analyze implements Expression.
func (f Float) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return &FloatType{}
}
func (Float) GetFlags() Flags { return 0 }

type Bool struct {
	Span  Span
	Value bool
}

func (b *Bool) String() string {
	return fmt.Sprintf("%v", b.Value)
}

// GetSpan implements Expression.
func (b Bool) GetSpan() Span {
	return b.Span
}

// Lower implements Expression.
func (b Bool) Lower(state *State) cpp.Expression {
	return fmt.Sprintf("%v", b.Value)
}

// Analyze implements Expression.
func (b Bool) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return &BoolType{}
}
func (Bool) GetFlags() Flags { return 0 }

type String struct {
	Span  Span
	Value string
}

func (s *String) String() string {
	return fmt.Sprintf("%q", s.Value)
}

// GetSpan implements Expression.
func (s String) GetSpan() Span {
	return s.Span
}

// Lower implements Expression.
func (s String) Lower(state *State) cpp.Expression {
	return fmt.Sprintf(`String_t(%q)`, s.Value)
}

// Analyze implements Expression.
func (s String) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	return &StringType{}
}

func (String) GetFlags() Flags { return 0 }

type Variable struct {
	Name  Name
	flags Flags
}

func (v Variable) String() string {
	if v.Name.String == "" {
		return "<empty variable name>"
	} else {
		return fmt.Sprintf("%s", v.Name.String)
	}
}

// GetSpan implements Expression.
func (v *Variable) GetSpan() Span {
	return v.Name.Span
}

// Analyze implements Expression.
func (v *Variable) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	_type, flags := anal.GetType(v.Name)
	v.flags = flags
	return _type
}

func (v *Variable) GetFlags() Flags {
	return v.flags
}

// Lower implements Expression.
func (v *Variable) Lower(state *State) cpp.Expression {
	return v.Name.Lower()
}

type FunctionCall struct {
	Span             Span
	Function         Expression
	Argument         Expression
	parameterIsTuple bool
	// A field for storing data that builtin functions need to transfer
	// between Analyze and Lower.
	builtinData any
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

func (f *FunctionCall) GetFlags() (flags Flags) {
	functionFlags := f.Function.GetFlags()
	argumentFlags := f.Argument.GetFlags()
	if functionFlags&IMPURE_FUNCTION != 0 {
		flags |= IMPURE
	}
	// If the argument is an IMPURE_FUNCTION it may or may not be evaluated.
	// To prevent complex control flow analysis, assume it is called.
	if argumentFlags&IMPURE_FUNCTION != 0 || argumentFlags&IMPURE != 0 {
		flags |= IMPURE
	}
	return
}

func checkIsTuple(t TypeValue, at Span, anal Analyzer) (tupleType *TupleType) {
	tupleType, argumentIsTuple := t.(*TupleType)
	if !argumentIsTuple {
		anal.PushError(ExpectedTuple{Found: t, At: at})
	}
	return
}

func checkTupleTypeArity(tupleType *TupleType, expected int, at Span, anal Analyzer) {
	if len(tupleType.Elements) != expected {
		anal.PushError(ArityMismatch{
			Expected: expected,
			Found:    len(tupleType.Elements),
			At:       at,
		})
	}
}

// Match builtin functions that need to be handled differently as their types
// cannot be expressed by Yune. Returns `nil` if it is not a special builtin.
func (f *FunctionCall) AnalyzeBuiltins(anal Analyzer) (returnType TypeValue) {
	name, functionIsVariable := f.getFunctionName()
	if functionIsVariable {
		switch name {
		// getTupleElement_(<any tuple type>, index Int): <element type at index>
		case "getTupleElement_":
			argumentType := f.Argument.Analyze(nil, anal)
			tupleArgumentType := checkIsTuple(argumentType, f.Argument.GetSpan(), anal)
			checkTupleTypeArity(tupleArgumentType, 2, f.Argument.GetSpan(), anal)
			firstElementTupleType := checkIsTuple(tupleArgumentType.Elements[0], f.Argument.GetSpan(), anal)
			// the second argument should always be an integer, since the compiler constructs this FunctionCall
			_ = tupleArgumentType.Elements[1].(*IntType)
			index := f.Argument.(*Tuple).Elements[1].(*Integer)
			return firstElementTupleType.Elements[index.Value]
		default:
		}
	}
	return nil
}

// Analyze implements Expression.
func (f *FunctionCall) Analyze(expected TypeValue, anal Analyzer) (returnType TypeValue) {
	if returnType = f.AnalyzeBuiltins(anal); returnType != nil {
		return
	}
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
	if !IsSubType(argumentType, expectedArgumentType) {
		anal.PushError(UnexpectedType{
			Expected: expectedArgumentType,
			Found:    argumentType,
			At:       f.Argument.GetSpan(),
		})
	}
	_, parameterIsTuple := functionType.Argument.(*TupleType)
	f.parameterIsTuple = parameterIsTuple
	return
}

func (f *FunctionCall) getFunctionName() (string, bool) {
	functionVariable, functionIsVariable := f.Function.(*Variable)
	if functionIsVariable {
		return functionVariable.Name.String, true
	}
	return "", false
}

// Lower implements Expression.
func (f *FunctionCall) Lower(state *State) cpp.Expression {
	// match builtin functions that need to be handled differently
	{
		name, functionIsVariable := f.getFunctionName()
		if functionIsVariable {
			switch name {
			case "getTupleElement_":
				tupleArgument, _ := f.Argument.(*Tuple)
				tuple := tupleArgument.Elements[0].Lower(state)
				index := tupleArgument.Elements[1].(*Integer).Value
				return fmt.Sprintf(`std::get<%d>(%s)`, index, tuple)
			}
		}
	}
	if f.parameterIsTuple {
		// calls the function with a tuple of arguments
		return fmt.Sprintf(`apply_(%s, %s)`, f.Function.Lower(state), f.Argument.Lower(state))
	}
	return fmt.Sprintf(`%s(%s)`, f.Function.Lower(state), f.Argument.Lower(state))
}

type List struct {
	Span        Span
	Elements    []Expression
	elementType TypeValue
}

func (t List) String() string {
	s := "["
	for i, element := range t.Elements {
		s += element.String()
		if i+1 < len(t.Elements) {
			s += ", "
		}
	}
	s += "]"
	return s
}

// GetSpan implements Expression.
func (t *List) GetSpan() Span {
	return t.Span
}

// Analyze implements Expression.
func (t *List) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	expectedListType, expectsList := expected.(*ListType)

	if len(t.Elements) == 0 {
		if expectsList {
			return expectedListType
		}
		return &ListType{Element: nil} // cannot determine element type
	}
	if expectsList {
		t.elementType = expectedListType.Element
	} else {
		t.elementType = &UnionType{}
	}
	for _, element := range t.Elements {
		elementType := element.Analyze(t.elementType, anal)
		t.elementType = NewUnionType(t.elementType, elementType)
	}
	return &ListType{Element: t.elementType}
}

func (t *List) GetFlags() (flags Flags) {
	for _, element := range t.Elements {
		flags |= element.GetFlags() & IMPURE
	}
	return
}

// Lower implements Expression.
func (t *List) Lower(state *State) cpp.Expression {
	return fmt.Sprintf(
		`List_t<%s>{%s}`,
		t.elementType.LowerType(), util.JoinFunc(t.Elements, ", ", func(e Expression) cpp.Expression {
			return e.Lower(state)
		}),
	)
}

type Tuple struct {
	Span     Span
	Elements []Expression
	isType   bool
}

func (t Tuple) String() string {
	s := "("
	for i, element := range t.Elements {
		s += element.String()
		if i+1 < len(t.Elements) {
			s += ", "
		}
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
		} else if expected != nil && !expected.Eq(&TypeType{}) {
			expectedElementType = &TypeType{}
		}
		elementType := t.Elements[i].Analyze(expectedElementType, anal)
		_type.Elements = append(_type.Elements, elementType)
	}
	t.isType = expected != nil && expected.Eq(&TypeType{})
	if t.isType {
		return &TypeType{}
	}
	return _type
}

func (t *Tuple) GetFlags() (flags Flags) {
	for _, element := range t.Elements {
		flags |= element.GetFlags() & IMPURE
	}
	return
}

// Lower implements Expression.
func (t *Tuple) Lower(state *State) cpp.Expression {
	if t.isType {
		if len(t.Elements) == 0 {
			return `box_f(TupleType_t { .elements = {} })`
		}
		elements := util.JoinFunc(t.Elements, ", ", func(e Expression) string {
			return e.Lower(state)
		})
		return fmt.Sprintf(`box_f(TupleType_t { .elements = {%s} })`, elements)
	} else {
		return fmt.Sprintf(`std::make_tuple(%s)`, util.JoinFunc(t.Elements, ", ", func(e Expression) cpp.Expression {
			return e.Lower(state)
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
	return util.JoinFunc(m.Lines, "\n", func(l MacroLine) string {
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
	v := anal.Evaluate(fmt.Sprintf(
		`(%s)(%q, get_type)`,
		m.Function.Lower(anal.State), m.GetText(),
	))
	// v is Union[String, Expression]
	// First try to unmarshal a String.
	errorBytes := v.GetStringBytes()
	if errorBytes != nil {
		anal.PushError(MacroOutputError{
			Macro:   m.Function,
			Message: string(errorBytes),
		})
	}
	m.Result = UnmarshalExpression(v)
	fmt.Printf("macro result: %s\n", m.Result)
	_type := m.Result.Analyze(expected, anal)
	return _type
}

func (m *Macro) GetFlags() (flags Flags) {
	return m.Result.GetFlags()
}

// Lower implements Expression.
func (m *Macro) Lower(state *State) cpp.Expression {
	return m.Result.Lower(state)
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
	var expressionType TypeValue
	switch u.Op {
	case "-":
		expressionType = u.Expression.Analyze(nil, anal)
		if !expressionType.Eq(&IntType{}) && expressionType.Eq(&FloatType{}) {
			anal.PushError(InvalidUnaryExpressionType{
				Op:   u.Op,
				Type: expressionType,
				At:   u.Span,
			})
		}
	case ";":
		expressionType = u.Expression.Analyze(&BoolType{}, anal)
		if !expressionType.Eq(&BoolType{}) {
			anal.PushError(InvalidUnaryExpressionType{
				Op:   u.Op,
				Type: expressionType,
				At:   u.Span,
			})
		}
	default:
		panic(fmt.Sprintf("unexpected ast.UnaryOp: %#v", u.Op))
	}
	return expressionType
}

func (u *UnaryExpression) GetFlags() (flags Flags) {
	return u.Expression.GetFlags()
}

// Lower implements Expression.
func (u *UnaryExpression) Lower(state *State) cpp.Expression {
	// TODO: parens as-needed
	switch u.Op {
	case "-":
		return "-" + u.Expression.Lower(state)
	case ";":
		return "!" + u.Expression.Lower(state)
	default:
		panic(fmt.Sprintf("unexpected ast.UnaryOp: %#v", u.Op))
	}
}

type UnaryOp string

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
	case Add:
		if !leftType.Eq(&IntType{}) && !leftType.Eq(&FloatType{}) && !leftType.Eq(&StringType{}) {
			emitErr()
		}
		return leftType
	case
		Divide,
		Multiply,
		Subtract:
		if !leftType.Eq(&IntType{}) && !leftType.Eq(&FloatType{}) {
			emitErr()
		}
		return leftType
	case
		Greater,
		GreaterEqual,
		Less,
		LessEqual:
		if !leftType.Eq(&IntType{}) && !leftType.Eq(&FloatType{}) {
			emitErr()
		}
		return &BoolType{}
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

func (b *BinaryExpression) GetFlags() (flags Flags) {
	return b.Left.GetFlags() | b.Right.GetFlags()
}

// Lower implements Expression.
func (b *BinaryExpression) Lower(state *State) cpp.Expression {
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
		Subtract,
		Or,
		And:
		op = string(b.Op)
	case NotEqual:
		op = "!="
	default:
		panic(fmt.Sprintf("unexpected ast.BinaryOp: %#v", b.Op))
	}
	return b.Left.Lower(state) + " " + op + " " + b.Right.Lower(state)
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
	NotEqual     BinaryOp = ";="
	LessEqual    BinaryOp = "<="
	GreaterEqual BinaryOp = ">="
	Or           BinaryOp = "or"
	And          BinaryOp = "and"
)

type StructExpression struct {
	Span   Span
	Name   Name
	Fields map[string]Expression
}

func (s StructExpression) String() string {
	return s.Name.String + "<fields>"
}

// GetSpan implements Expression.
func (s *StructExpression) GetSpan() Span {
	return s.Span
}

func (s StructExpression) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	maybeStructType, _ := anal.GetType(s.Name)
	structType, isStructType := maybeStructType.(*StructType)
	if !isStructType {
		anal.PushError(NotAStruct{
			Found: structType,
			At:    s.Name.Span,
		})
	}

	// compare and analyze fields and check for duplicate field names
	panic("unimplemented")
}

func (s *StructExpression) GetFlags() (flags Flags) {
	for _, expression := range s.Fields {
		flags |= expression.GetFlags()
	}
	return
}

func (s StructExpression) Lower(state *State) cpp.Expression {
	fields := ""
	for key, value := range s.Fields {
		fields += fmt.Sprintf(".%s = %s,\n", key, value.Lower(state))
	}
	return fmt.Sprintf("%s_t {%s\n}", s.Name.String, fields)
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
	for _, capture := range *anal.Table.localCaptures {
		c.captures[capture.name.String] = capture.declaration.GetDeclaredType()
	}
	return getFunctionType(c.Parameters, c.ReturnType)
}

func (c *Closure) GetFlags() (flags Flags) {
	bodyFlags := c.Body.GetFlags()
	if bodyFlags&IMPURE != 0 {
		flags |= IMPURE_FUNCTION
	}
	return
}

func (c Closure) LowerParameters() string {
	return util.JoinFunc(c.Parameters, ", ", FunctionParameter.Lower)
}

// Lower implements Expression.
func (c *Closure) Lower(state *State) cpp.Expression {
	id := state.registerClosure(c)
	fields := ""
	captureArguments := ""
	captures := ""
	if len(c.captures) == 0 {
		captures = `""`
	}
	for captureName, captureType := range c.captures {
		fields += captureType.LowerType() + " " + captureName + ";\n"
		if captureArguments != "" {
			captureArguments += ", "
			captures += ` + ", " + `
		}
		captureArguments += captureName
		// not using newlines because these are automatically escaped by the evaluator
		// which results in malformed JSON
		captures += fmt.Sprintf(
			`capture_toJson_(%q, %q, %s)`,
			captureName,
			state.registerTypeValue(captureType),
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
        std::string toJson_() const {
            return closure_toJson_(%s, %q);
        }
        %s
    } closure{%s};
    return closure;
}()`,
		lambdaSymbol,
		c.ReturnType.Lower(),
		c.LowerParameters(),
		cpp.Block(c.Body.Lower(state)),
		captures,
		id,
		fields,
		captureArguments)
	return lowered
}

// Tries to unmarshal an Expression, panicking if the union key does not match an Expression.
func UnmarshalExpression(data *fj.Value) (expr Expression) {
	object := data.GetObject()
	key, v := fjUnmarshalStruct(object)
	switch key {
	case "IntegerExpression":
		expr = &Integer{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Value: UnmarshalItem(v, (*fj.Value).Int64, "value"),
		}
	case "FloatExpression":
		expr = &Float{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Value: UnmarshalItem(v, (*fj.Value).Float64, "value"),
		}
	case "BoolExpression":
		expr = &Bool{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Value: UnmarshalItem(v, (*fj.Value).Bool, "value"),
		}
	case "StringExpression":
		expr = &String{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Value: UnmarshalString(v, "value"), // can be empty
		}
	case "VariableExpression":
		expr = &Variable{
			Name: Name{
				Span:   fjUnmarshal(v.Get("span"), Span{}),
				String: UnmarshalNonEmptyString(v, "name"),
			},
		}
	case "FunctionCallExpression":
		expr = &FunctionCall{
			Span:     fjUnmarshal(v.Get("span"), Span{}),
			Function: UnmarshalExpression(v.Get("function")),
			Argument: UnmarshalExpression(v.Get("argument")),
		}
	case "ListExpression":
		expr = &List{
			Span:     fjUnmarshal(v.Get("span"), Span{}),
			Elements: util.Map(v.Get("elements").GetArray(), UnmarshalExpression),
		}
	case "TupleExpression":
		expr = &Tuple{
			Span:     fjUnmarshal(v.Get("span"), Span{}),
			Elements: util.Map(v.Get("elements").GetArray(), UnmarshalExpression),
		}
	case "Macro":
		expr = &Macro{
			Span:     fjUnmarshal(v.Get("span"), Span{}),
			Function: Variable{Name: Name{String: UnmarshalNonEmptyString(v, "macro"), Span: Span{}}},
			Lines: util.Map(strings.Split(UnmarshalString(v, "text"), "\n"), func(s string) MacroLine {
				return MacroLine{Span: Span{}, Text: s}
			}),
		}
	case "UnaryExpression":
		expr = &UnaryExpression{
			Span:       fjUnmarshal(v.Get("span"), Span{}),
			Op:         UnaryOp(UnmarshalNonEmptyString(v, "op")),
			Expression: UnmarshalExpression(v.Get("expression")),
		}
	case "BinaryExpression":
		expr = &BinaryExpression{
			Span:  fjUnmarshal(v.Get("span"), Span{}),
			Op:    BinaryOp(UnmarshalNonEmptyString(v, "op")),
			Left:  UnmarshalExpression(v.Get("left")),
			Right: UnmarshalExpression(v.Get("right")),
		}
	case "StructExpression":
		panic("unimplemented")
	case "ClosureExpression":
		expr = &Closure{
			Span: fjUnmarshal(v.Get("span"), Span{}),
			Parameters: util.Map(v.GetArray("parameters"), func(v *fj.Value) FunctionParameter {
				return FunctionParameter{
					Span: fjUnmarshal(v.Get("span"), Span{}),
					Name: Name{
						Span:   Span{},
						String: UnmarshalNonEmptyString(v, "name"),
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
var _ Expression = &List{}
var _ Expression = &Tuple{}
var _ Expression = &Macro{}
var _ Expression = &UnaryExpression{}
var _ Expression = &BinaryExpression{}
var _ Expression = &StructExpression{}
var _ Expression = &Closure{}
