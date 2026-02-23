package ast

var BuiltinDeclarations = []BuiltinDeclaration{
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
type BuiltinDeclaration struct {
	Name string
	Type TypeValue
}

// Analyze implements TopLevelDeclaration.
func (b *BuiltinDeclaration) Analyze(anal Analyzer) {
	anal.Declare(b)
	anal.Define(b)
}

// GetDeclaredType implements TopLevelDeclaration.
func (b *BuiltinDeclaration) GetDeclaredType() TypeValue {
	return b.Type
}

// GetName implements TopLevelDeclaration.
func (b *BuiltinDeclaration) GetName() Name {
	return Name{String: b.Name}
}

// GetSpan implements TopLevelDeclaration.
func (b *BuiltinDeclaration) GetSpan() Span {
	return Span{}
}

// LowerDeclaration implements TopLevelDeclaration.
func (b *BuiltinDeclaration) LowerDeclaration() string { return "" }

// LowerDefinition implements TopLevelDeclaration.
func (b *BuiltinDeclaration) LowerDefinition() string { return "" }

var _ TopLevelDeclaration = (*BuiltinDeclaration)(nil)
