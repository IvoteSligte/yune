package ast

var BuiltinDeclarations = []BuiltinDeclaration{
	{"Type", &TypeType{}, 0},
	{"Int", &TypeType{}, 0},
	{"Float", &TypeType{}, 0},
	{"Bool", &TypeType{}, 0},
	{"String", &TypeType{}, 0},
	{"List", &FnType{
		Argument: &TypeType{},
		Return:   &TypeType{},
	}, 0},
	{"Fn", &FnType{
		Argument: &TupleType{Elements: []TypeValue{&TypeType{}, &TypeType{}}},
		Return:   &TypeType{},
	}, 0},
	{"toFloat", &FnType{
		Argument: &IntType{},
		Return:   &FloatType{},
	}, 0},
	{"Expression", &TypeType{}, 0},
	{"stringLiteral", &FnType{
		Argument: &StringType{},
		Return:   ExpressionType,
	}, 0},
	{"variable", &FnType{
		Argument: &StringType{},
		Return:   ExpressionType,
	}, 0},
	{"unaryExpression", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			&StringType{},
			ExpressionType,
		}},
		Return: ExpressionType,
	}, 0},
	{"binaryExpression", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			&StringType{},
			ExpressionType,
			ExpressionType,
		}},
		Return: ExpressionType,
	}, 0},
	{"functionCall", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			ExpressionType,
			ExpressionType,
		}},
		Return: ExpressionType,
	}, 0},
	// TODO: other Expression constructors
	{"panic", &FnType{
		Argument: &StringType{},
		// empty union type is non-constructable and Union(Union(), T, U) == Union(T, U)
		Return: &UnionType{},
	}, IMPURE_FUNCTION},
	{"printlnString", &FnType{
		Argument: &StringType{},
		Return:   &TupleType{},
	}, IMPURE_FUNCTION},
	{"len", &FnType{
		Argument: &StringType{},
		Return:   &IntType{},
	}, 0},
	{"subString", &FnType{
		Argument: &TupleType{Elements: []TypeValue{&StringType{}, &IntType{}, &IntType{}}},
		Return:   &StringType{},
	}, 0},
	{"Union", &FnType{
		Argument: &ListType{Element: &TypeType{}},
		Return:   &TypeType{},
	}, 0},
}

// A declaration that is written in `pb.hpp`.
// This struct only makes the compiler aware of it.
type BuiltinDeclaration struct {
	Name  string
	Type  TypeValue
	Flags Flags
}

// Analyze implements TopLevelDeclaration.
func (b *BuiltinDeclaration) Analyze(anal Analyzer) {
	anal.Declare(b)
	anal.Define(b)
}

func (b *BuiltinDeclaration) GetFlags() Flags {
	return b.Flags
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
func (b *BuiltinDeclaration) LowerDeclaration(state *State) string { return "" }

// LowerDefinition implements TopLevelDeclaration.
func (b *BuiltinDeclaration) LowerDefinition(state *State) string { return "" }

var _ TopLevelDeclaration = (*BuiltinDeclaration)(nil)
