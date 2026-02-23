package ast

var BuiltinDeclarations = []Builtin{
	{"Type", &TypeType{}},
	{"Int", &TypeType{}},
	{"Float", &TypeType{}},
	{"Bool", &TypeType{}},
	{"String", &TypeType{}},
	{"List", &FnType{
		Argument: &TypeType{},
		Return:   &TypeType{},
	}},
	{"Fn", &FnType{
		Argument: &TupleType{Elements: []TypeValue{&TypeType{}, &TypeType{}}},
		Return:   &TypeType{},
	}},
	{"Expression", &TypeType{}},
	{"stringLiteral", &FnType{
		Argument: &StringType{},
		Return:   ExpressionType,
	}},
	{"printlnString", &FnType{
		Argument: &StringType{},
		Return:   &TupleType{},
	}},
}

// A declaration that is written in `pb.hpp`.
// This struct only makes the compiler aware of it.
type Builtin struct {
	Name string
	Type TypeValue
}

// Analyze implements TopLevelDeclaration.
func (b *Builtin) Analyze(anal Analyzer) {
	anal.Declare(b)
	anal.Define(b)
}

// GetDeclaredType implements TopLevelDeclaration.
func (b *Builtin) GetDeclaredType() TypeValue {
	return b.Type
}

// GetName implements TopLevelDeclaration.
func (b *Builtin) GetName() Name {
	return Name{String: b.Name}
}

// GetSpan implements TopLevelDeclaration.
func (b *Builtin) GetSpan() Span {
	return Span{}
}

// LowerDeclaration implements TopLevelDeclaration.
func (b *Builtin) LowerDeclaration() string { return "" }

// LowerDefinition implements TopLevelDeclaration.
func (b *Builtin) LowerDefinition() string { return "" }

var _ TopLevelDeclaration = (*Builtin)(nil)
