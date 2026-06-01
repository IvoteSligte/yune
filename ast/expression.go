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
	return fmt.Sprintf(`%q`, s.Value)
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
		anal.ReportError(ExpectedTuple{Found: t, At: at})
	}
	return
}

func checkTupleTypeArity(tupleType *TupleType, expected int, at Span, anal Analyzer) {
	if len(tupleType.Elements) != expected {
		anal.ReportError(ArityMismatch{
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
	if !functionIsVariable {
		return nil
	}
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
	// inject(<any type>): Expression
	case "inject":
		f.Argument.Analyze(nil, anal)
		return ExpressionType
	// len(Union[String, List(<any type>)]): Int
	case "len":
		argumentType := f.Argument.Analyze(nil, anal)
		_, isListType := argumentType.(*ListType)
		if !argumentType.Eq(&StringType{}) && !isListType {
			anal.ReportError(UnexpectedLenArgument{
				Found: argumentType,
				At:    f.Argument.GetSpan(),
			})
		}
		return &IntType{}
	// append(List(<element type>), <element type>): List(<element type>)
	case "append":
		argumentType := f.Argument.Analyze(nil, anal)
		tupleArgumentType := checkIsTuple(argumentType, f.Argument.GetSpan(), anal)
		checkTupleTypeArity(tupleArgumentType, 2, f.Argument.GetSpan(), anal)
		firstArgumentType := tupleArgumentType.Elements[0]
		firstArgumentListType, firstArgumentIsList := firstArgumentType.(*ListType)
		if !firstArgumentIsList {
			anal.ReportError(ExpectedList{
				Found: argumentType,
				At:    f.Argument.GetSpan(),
			})
		}
		listElementType := firstArgumentListType.Element
		secondArgumentType := tupleArgumentType.Elements[1]
		if !listElementType.Eq(secondArgumentType) {
			anal.ReportError(UnexpectedType{
				Expected: listElementType,
				Found:    secondArgumentType,
				At:       f.Argument.(*Tuple).Elements[1].GetSpan(),
			})
		}
		f.parameterIsTuple = true
		return firstArgumentListType
	// get(List(<any type>), Int): <same element type as argument>
	case "get":
		argumentType := f.Argument.Analyze(nil, anal)
		tupleArgumentType := checkIsTuple(argumentType, f.Argument.GetSpan(), anal)
		checkTupleTypeArity(tupleArgumentType, 2, f.Argument.GetSpan(), anal)
		firstArgumentType := tupleArgumentType.Elements[0]
		firstArgumentListType, firstArgumentIsList := firstArgumentType.(*ListType)
		if !firstArgumentIsList {
			anal.ReportError(ExpectedList{
				Found: firstArgumentType,
				At:    f.Argument.(*Tuple).Elements[0].GetSpan(),
			})
		}
		secondArgumentType := tupleArgumentType.Elements[1]
		if !secondArgumentType.Eq(&IntType{}) {
			anal.ReportError(UnexpectedType{
				Expected: &IntType{},
				Found:    secondArgumentType,
				At:       f.Argument.(*Tuple).Elements[1].GetSpan(),
			})
		}
		f.parameterIsTuple = true
		return firstArgumentListType.Element
	// set(List(<element type>), Int, <element type>): ()
	case "set":
		argumentType := f.Argument.Analyze(nil, anal)
		tupleArgumentType := checkIsTuple(argumentType, f.Argument.GetSpan(), anal)
		checkTupleTypeArity(tupleArgumentType, 3, f.Argument.GetSpan(), anal)
		firstArgumentType := tupleArgumentType.Elements[0]
		firstArgumentListType, firstArgumentIsList := firstArgumentType.(*ListType)
		if !firstArgumentIsList {
			anal.ReportError(ExpectedList{
				Found: firstArgumentType,
				At:    f.Argument.(*Tuple).Elements[0].GetSpan(),
			})
		}
		secondArgumentType := tupleArgumentType.Elements[1]
		if !secondArgumentType.Eq(&IntType{}) {
			anal.ReportError(UnexpectedType{
				Expected: &IntType{},
				Found:    secondArgumentType,
				At:       f.Argument.(*Tuple).Elements[1].GetSpan(),
			})
		}
		thirdArgumentType := tupleArgumentType.Elements[2]
		if !firstArgumentListType.Element.Eq(thirdArgumentType) {
			anal.ReportError(UnexpectedType{
				Expected: firstArgumentListType.Element,
				Found:    thirdArgumentType,
				At:       f.Argument.(*Tuple).Elements[2].GetSpan(),
			})
		}
		f.parameterIsTuple = true
		return &TupleType{}
	default:
		return nil
	}
}

// Analyze implements Expression.
func (f *FunctionCall) Analyze(expected TypeValue, anal Analyzer) (returnType TypeValue) {
	if returnType = f.AnalyzeBuiltins(anal); returnType != nil {
		return
	}
	maybeFunctionType := f.Function.Analyze(nil, anal)
	functionType, isFunction := maybeFunctionType.(*FnType)
	if !isFunction {
		anal.ReportError(NotAFunction{
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
		anal.ReportError(UnexpectedType{
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
			t.elementType = expectedListType.Element
			return expectedListType
		}
		t.elementType = &UnionType{}
		return &ListType{Element: &UnionType{}} // cannot determine element type
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
		anal.ReportError(ArityMismatch{
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
	for i, line := range m.Lines {
		if i > 0 {
			s += "    " // indentation
		}
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
		fmt.Printf("macro function span: %s\n", m.Function.GetSpan())
		anal.ReportError(UnexpectedType{
			Expected: MacroFunctionType,
			Found:    functionType,
			At:       m.Function.GetSpan(),
		})
	}
	v := anal.Evaluate(fmt.Sprintf(
		`(%s)(%q, getType_c)`,
		m.Function.Lower(anal.State), m.GetText(),
	), m)
	// v is Union[String, Expression]
	// First try to unmarshal a String.
	errorTupleElements, isErrorTuple := TryUnmarshalTuple(v)
	if isErrorTuple {
		at := UnmarshalSpan(errorTupleElements[0], m)
		message := UnmarshalNonEmptyString(errorTupleElements[1])
		anal.ReportError(MacroOutputError{
			Macro:   m,
			Message: message,
			At:      at,
		})
	}
	m.Result = UnmarshalExpression(v, m)
	anal.MacroStack = append(anal.MacroStack, m)
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
			anal.ReportError(InvalidUnaryExpressionType{
				Op:   u.Op,
				Type: expressionType,
				At:   u.Span,
			})
		}
	case ";":
		expressionType = u.Expression.Analyze(&BoolType{}, anal)
		if !expressionType.Eq(&BoolType{}) {
			anal.ReportError(InvalidUnaryExpressionType{
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
		anal.ReportError(InvalidBinaryExpressionTypes{
			Op:    b.Op,
			Left:  leftType,
			Right: rightType,
			At:    b.Span,
		})
	}
	emitErr := func() {
		anal.ReportError(InvalidBinaryExpressionTypes{
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
		anal.ReportError(NotAStruct{
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
		c.captures[capture.name] = capture.declaration.GetDeclaredType()
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

func (c *Closure) LowerComplex(state *State, captureValues map[string]string) cpp.Expression {
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
		if captureValues == nil {
			captureArguments += captureName // capture variable in environment
		} else {
			captureArguments += captureValues[captureName]
		}
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

// Lower implements Expression.
func (c *Closure) Lower(state *State) cpp.Expression {
	return c.LowerComplex(state, nil)
}

func (r *RawString) GetSpan() Span {
	return r.Span
}

func (r *RawString) String() string {
	return fmt.Sprintf("`%s`", r.string)
}

func (r *RawString) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	if expected == nil {
		anal.ReportError(CannotDetermineRawType{Span: r.Span})
	}
	return expected
}

func (r *RawString) GetFlags() Flags {
	return 0 // assuming no side-effects
}

func (r *RawString) Lower(state *State) cpp.Expression {
	return r.string
}

type ValueExpression struct {
	Span  Span
	value *fj.Value
}

// Analyze implements Expression.
func (v *ValueExpression) Analyze(expected TypeValue, anal Analyzer) TypeValue {
	_type := anal.State.getValueType(v.value)
	if _type == nil {
		panic(fmt.Sprintf("Failed to get type of ValueExpression. Value: %s. State: %#v", v.value, anal.State))
	}
	return _type
}

// GetFlags implements Expression.
func (v *ValueExpression) GetFlags() Flags {
	return 0
}

// GetSpan implements Expression.
func (v *ValueExpression) GetSpan() Span {
	return v.Span
}

// Lower implements Expression.
func (v *ValueExpression) Lower(state *State) string {
	return state.lowerExpressionValue(v.value)
}

// String implements Expression.
func (v *ValueExpression) String() string {
	panic("unimplemented")
}

// Tries to unmarshal an Expression, panicking if the union key does not match an Expression.
func UnmarshalExpression(data *fj.Value, in *Macro) (expr Expression) {
	object := data.GetObject()
	key, v := fjUnmarshalStruct(object)
	switch key {
	case "IntegerExpression":
		expr = &Integer{
			Span:  UnmarshalLocation(v, in),
			Value: UnmarshalItem(v, (*fj.Value).Int64, "value"),
		}
	case "FloatExpression":
		expr = &Float{
			Span:  UnmarshalLocation(v, in),
			Value: UnmarshalItem(v, (*fj.Value).Float64, "value"),
		}
	case "BoolExpression":
		expr = &Bool{
			Span:  UnmarshalLocation(v, in),
			Value: UnmarshalItem(v, (*fj.Value).Bool, "value"),
		}
	case "StringExpression":
		expr = &String{
			Span:  UnmarshalLocation(v, in),
			Value: UnmarshalString(v, "value"), // can be empty
		}
	case "VariableExpression":
		expr = &Variable{
			Name: Name{
				Span:   UnmarshalLocation(v, in),
				String: UnmarshalNonEmptyString(v, "name"),
			},
		}
	case "FunctionCallExpression":
		expr = &FunctionCall{
			Span:     UnmarshalLocation(v, in),
			Function: UnmarshalExpression(v.Get("function"), in),
			Argument: UnmarshalExpression(v.Get("argument"), in),
		}
	case "ListExpression":
		expr = &List{
			Span: UnmarshalLocation(v, in),
			Elements: util.Map(UnmarshalArray(v, "elements"), func(v *fj.Value) Expression {
				return UnmarshalExpression(v, in)
			}),
		}
	case "TupleExpression":
		expr = &Tuple{
			Span: UnmarshalLocation(v, in),
			Elements: util.Map(UnmarshalArray(v, "elements"), func(v *fj.Value) Expression {
				return UnmarshalExpression(v, in)
			}),
		}
	case "MacroExpression":
		expr = &Macro{
			Span:     UnmarshalLocation(v, in),
			Function: Variable{Name: Name{String: UnmarshalNonEmptyString(v, "macro"), Span: Span{}}},
			Lines: util.Map(strings.Split(UnmarshalString(v, "text"), "\n"), func(s string) MacroLine {
				return MacroLine{Span: Span{}, Text: s} // TODO: set span
			}),
		}
	case "UnaryExpression":
		expr = &UnaryExpression{
			Span:       UnmarshalLocation(v, in),
			Op:         UnaryOp(UnmarshalNonEmptyString(v, "op")),
			Expression: UnmarshalExpression(v.Get("expression"), in),
		}
	case "BinaryExpression":
		expr = &BinaryExpression{
			Span:  UnmarshalLocation(v, in),
			Op:    BinaryOp(UnmarshalNonEmptyString(v, "op")),
			Left:  UnmarshalExpression(v.Get("left"), in),
			Right: UnmarshalExpression(v.Get("right"), in),
		}
	case "StructExpression":
		panic("unimplemented")
	case "ClosureExpression":
		expr = &Closure{
			Span: UnmarshalLocation(v, in),
			Parameters: util.Map(UnmarshalArray(v, "parameters"), func(v *fj.Value) FunctionParameter {
				elements := UnmarshalTuple(v)
				return FunctionParameter{
					Name: Name{
						Span:   Span{},
						String: UnmarshalNonEmptyString(elements[0]),
					},
					Type: UnmarshalType(elements[1], in),
				}
			}),
			ReturnType: UnmarshalType(v.Get("returnType"), in),
			Body:       UnmarshalBlock(v.Get("body"), in),
		}
	case "ValueExpression":
		expr = &ValueExpression{Span: UnmarshalLocation(v, in), value: v.Get("value")}
	case "Box": // boxing is irrelevant when unmarshalling expressions
		expr = UnmarshalExpression(v, in)
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
var _ Expression = &RawString{}
var _ Expression = &ValueExpression{}
