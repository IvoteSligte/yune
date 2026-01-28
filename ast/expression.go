package ast

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"slices"
	"strings"
	"yune/cpp"
	"yune/util"
	"yune/value"
)

func TaggedMarshalJSON(tag string, value any) ([]byte, error) {
	type Alias any
	return json.Marshal(&struct {
		Tag string `json:"$tag"`
		Alias
	}{
		Tag:   tag,
		Alias: Alias(value),
	})
}

var typeToTag = map[reflect.Type]string{
	// Expressions
	reflect.TypeFor[Integer]():          "Integer",
	reflect.TypeFor[Float]():            "Float",
	reflect.TypeFor[Bool]():             "Bool",
	reflect.TypeFor[String]():           "String",
	reflect.TypeFor[Variable]():         "Variable",
	reflect.TypeFor[FunctionCall]():     "FunctionCall",
	reflect.TypeFor[Tuple]():            "Tuple",
	reflect.TypeFor[UnaryExpression]():  "UnaryExpression",
	reflect.TypeFor[BinaryExpression](): "BinaryExpression",
}

var tagToType = util.MapMap(typeToTag, func(_type reflect.Type, tag string) (string, reflect.Type) {
	return tag, _type
})

func Marshal(v any) (_ []byte, err error) {
	r := reflect.ValueOf(v)
	t := r.Type()
	switch t.Kind() {
	case reflect.Interface, reflect.Pointer:
		return Marshal(r.Elem())
	case reflect.Struct:
		tag, ok := typeToTag[t]
		if !ok {
			panic("Tried to serialize struct type " + t.String() + " that does not have known tag.")
		}
		m := map[string][]byte{}
		m["$tag"] = []byte(tag)
		for i := range t.NumField() {
			field := t.Field(i)
			m[field.Name], err = Marshal(r.Field(i))
			if err != nil {
				return
			}
		}
		return json.Marshal(m)
	case reflect.Int, reflect.Float64, reflect.Bool, reflect.String:
		return json.Marshal(v)
	default:
		panic("Tried to serialize unserializable type " + t.String())
	}
}

func Unmarshal(data []byte) (v any, err error) {
	var tagged struct {
		Tag string `json:"$tag"`
	}
	err = json.Unmarshal(data, &tagged)
	if err != nil {
		return
	}
	if tagged.Tag == "" {
		// assume unambiguous
		err = json.Unmarshal(data, &v)
		return
	}
	t, ok := tagToType[tagged.Tag]
	if !ok {
		panic("Tried to deserialize JSON with unknown tag " + tagged.Tag)
	}
	m := map[string]json.RawMessage{}
	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	var r = reflect.Zero(t)
	for i := range t.NumField() {
		field := r.Field(i)
		v, err = Unmarshal(m[t.Field(i).Name])
		if err != nil {
			return
		}
		field.Set(reflect.ValueOf(v))
	}
	v = r.Interface()
	return
}

func GetJSONTag(data []byte) (tag string, err error) {
	var tagged struct {
		Tag string `json:"$tag"`
	}
	err = json.Unmarshal(data, &tagged)
	tag = tagged.Tag
	return
}

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

func ExpressionUnmarshalJSON(data []byte, e *Expression) error {
	tag, err := GetJSONTag(data)
	if err != nil {
		return err
	}
	switch tag {
	case "Integer":
		*e = Integer{}
	case "Float":
		*e = Float{}
	case "Bool":
		*e = Bool{}
	case "String":
		*e = String{}
	case "Variable":
		*e = &Variable{}
	case "FunctionCall":
		*e = &FunctionCall{}
	case "Tuple":
		*e = &Tuple{}
	case "UnaryExpression":
		*e = &UnaryExpression{}
	case "BinaryExpression":
		*e = &BinaryExpression{}
	default:
		panic("Invalid JSON Expression tag: " + tag)
	}
	return json.Unmarshal(data, e)
}

type DefaultExpression struct{}

var _ Expression = DefaultExpression{}

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
func (d DefaultExpression) GetType() value.Type {
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
func (d DefaultExpression) InferType(expected value.Type, deps DeclarationTable) (errors []error) {
	return
}

// Lower implements Expression.
func (d DefaultExpression) Lower() cpp.Expression {
	panic("DefaultExpression.Lower() should be overridden")
}

type Integer struct {
	DefaultExpression
	Span  Span
	Value int64
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
func (i Integer) GetType() value.Type {
	return IntType
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
func (f Float) GetType() value.Type {
	return FloatType
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
func (b Bool) GetType() value.Type {
	return BoolType
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
func (s String) GetType() value.Type {
	return StringType
}

type Variable struct {
	DefaultExpression
	Type value.Type
	Name Name
}

// GetSpan implements Expression.
func (v *Variable) GetSpan() Span {
	return v.Name.Span
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
	decl, _ := deps.Get(v.Name.String)
	if decl.GetDeclaredType() == value.Type("") {
		log.Printf("WARN: Type queried at %s before being calculated on declaration '%s'.", v.Name.Span, v.Name.String)
	}
	v.Type = decl.GetDeclaredType()
	return
}

// Lower implements Expression.
func (v *Variable) Lower() cpp.Expression {
	return cpp.Variable(v.Name.String)
}

type FunctionCall struct {
	Span     Span
	Type     value.Type
	Function Expression
	Argument Expression
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *FunctionCall) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	return util.FirstError(
		json.Unmarshal(data, &m),
		json.Unmarshal(m["Span"], &f.Span),
		json.Unmarshal(m["Type"], &f.Type),
		ExpressionUnmarshalJSON(m["Function"], &f.Function),
		ExpressionUnmarshalJSON(m["Argument"], &f.Argument),
	)
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
	// NOTE: should functions return () instead of Nil?
	if !argumentType.IsTuple() {
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
	Span Span
	// Inferred type
	Type     value.Type
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

// TODO: type check Function
type Macro struct {
	Span Span
	// Function that evaluates the macro.
	Function Variable
	Lines    []MacroLine
	// Result after evaluating the macro.
	Result Expression
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
	splits := strings.Split(s, "; ")
	errorMessage := splits[0]
	if errorMessage != "" {
		panic("Macro error message: " + errorMessage)
	}
	expressionString := splits[1]
	stringLiteral := expressionString[1 : len(expressionString)-1]
	stringLiteral = strings.ReplaceAll(stringLiteral, `\"`, `"`)
	// stringLiteral = strings.ReplaceAll(stringLiteral, `\\`, `\`)
	m.Result = String{
		Span:  Span{}, // TODO: span
		Value: stringLiteral,
	}
}

type UnaryExpression struct {
	Span       Span
	Type       value.Type
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
	Span  Span
	Type  value.Type
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
