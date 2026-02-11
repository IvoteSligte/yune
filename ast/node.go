package ast

import (
	"fmt"
	"yune/cpp"
)

type Span struct {
	Line   int
	Column int
}

func (s Span) GetSpan() Span {
	return s
}

func (s Span) String() string {
	return fmt.Sprintf("%d:%d", s.Line, s.Column)
}

func (s Span) Lower() cpp.Expression {
	return fmt.Sprintf(`Span(%d, %d)`, s.Line, s.Column)
}

type Name struct {
	Span
	String string
}

// Lowers a name, renaming in case of naming conflicts with reserved identifiers.
func (n Name) Lower() string {
	switch n.String {
	// main function cannot be a C++ struct with operator(), which Yune generates by default
	// so it is renamed and a wrapper is generated
	case "main":
		return "main_"
	// C++ keywords
	case "alignas",
		"alignof",
		"and",
		"and_eq",
		"asm",
		"atomic_cancel",
		"atomic_commit",
		"atomic_noexcept",
		"auto",
		"bitand",
		"bitor",
		"bool",
		"break",
		"case",
		"catch",
		"char",
		"char8_t",
		"char16_t",
		"char32_t",
		"class",
		"compl",
		"concept",
		"const",
		"consteval",
		"constexpr",
		"constinit",
		"const_cast",
		"continue",
		"contract_assert",
		"co_await",
		"co_return",
		"co_yield",
		"decltype",
		"default",
		"delete",
		"do",
		"double",
		"dynamic_cast",
		"else",
		"enum",
		"explicit",
		"export",
		"extern",
		"false",
		"float",
		"for",
		"friend",
		"goto",
		"if",
		"inline",
		"int",
		"long",
		"mutable",
		"namespace",
		"new",
		"noexcept",
		"not",
		"not_eq",
		"nullptr",
		"operator",
		"or",
		"or_eq",
		"private",
		"protected",
		"public",
		"reflexpr",
		"register",
		"reinterpret_cast",
		"requires",
		"return",
		"short",
		"signed",
		"sizeof",
		"static",
		"static_assert",
		"static_cast",
		"struct",
		"switch",
		"synchronized",
		"template",
		"this",
		"thread_local",
		"throw",
		"true",
		"try",
		"typedef",
		"typeid",
		"typename",
		"union",
		"unsigned",
		"using",
		"virtual",
		"void",
		"volatile",
		"wchar_t",
		"while",
		"xor",
		"xor_eq":
		return n.String + "_"
	default:
		return n.String
	}
}

type Node interface {
	GetSpan() Span
}

type IName interface {
	GetName() string
	GetSpan() Span
}

type Errors = []error

type Query struct {
	Expression
	Setter func(json string)
}

func (q Query) SetValue(json string) {
	panic("unimplemented") // NOTE: also need to call Analyze again on wrapper around destination (macro/type)
}

type Analyzer struct {
	Errors      *Errors
	NeedsTypeOf *[]Name
	Queries     *[]Query
	Table       DeclarationTable
}

func (a Analyzer) PushError(err error) {
	*a.Errors = append(*a.Errors, err)
}

func (a Analyzer) AppendErrors(errors ...error) {
	*a.Errors = append(*a.Errors, errors...)
}

func (a Analyzer) HasErrors() bool {
	return len(*a.Errors) > 0
}

func SubAnalyze[T any](a Analyzer, f func(Analyzer) T) T {
	sub := Analyzer{
		Errors:      a.Errors,
		Queries:     a.Queries,
		NeedsTypeOf: &[]Name{},
		// Evaluation cannot depend on variable declarations in the outer scope
		// as the values of these are not known.
		Table: a.Table.TopLevel(),
	}
	return f(sub)
}

func (a Analyzer) Evaluate(expr Expression, setter func(json string)) {
	*a.Queries = append(*a.Queries, Query{
		Expression: expr,
		Setter:     setter,
	})
}

func (a Analyzer) IsDone() bool {
	return len(*a.Queries) == 0 && len(*a.Errors) == 0
}

func (a Analyzer) NewScope() Analyzer {
	a.Table = a.Table.NewScope()
	return a
}

func (a Analyzer) GetType(name Name) TypeValue {
	decl, ok := a.Table.Get(name.String)
	if !ok {
		panic("Unknown declaration: " + name.String)
	}
	_type := decl.GetDeclaredType()
	// if _type == nil {
	// 	*a.NeedsTypeOf = append(*a.NeedsTypeOf, name)
	// }
	return _type
}

func (a Analyzer) GetScope() map[string]Declaration {
	return a.Table.declarations
}
