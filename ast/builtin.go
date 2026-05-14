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
	{"Statement", &TypeType{}, 0},
	{"integerExpression", &FnType{Argument: &TupleType{Elements: []TypeValue{&IntType{}, &IntType{}}}, Return: ExpressionType}, 0},
	{"floatExpression", &FnType{Argument: &TupleType{Elements: []TypeValue{&IntType{}, &FloatType{}}}, Return: ExpressionType}, 0},
	{"boolExpression", &FnType{Argument: &TupleType{Elements: []TypeValue{&IntType{}, &BoolType{}}}, Return: ExpressionType}, 0},
	{"stringExpression", &FnType{Argument: &TupleType{Elements: []TypeValue{&IntType{}, &StringType{}}}, Return: ExpressionType}, 0},
	{"variableExpression", &FnType{Argument: &TupleType{Elements: []TypeValue{&IntType{}, &StringType{}}}, Return: ExpressionType}, 0},
	{"unaryExpression", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			&IntType{},
			&StringType{},
			ExpressionType,
		}},
		Return: ExpressionType,
	}, 0},
	{"binaryExpression", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			&IntType{},
			&StringType{},
			ExpressionType,
			ExpressionType,
		}},
		Return: ExpressionType,
	}, 0},
	{"functionCallExpression", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			&IntType{},
			ExpressionType,
			ExpressionType,
		}},
		Return: ExpressionType,
	}, 0},
	{"closureExpression", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			&IntType{},
			// parameters: List((String, Expression))
			&ListType{Element: &TupleType{Elements: []TypeValue{&StringType{}, ExpressionType}}},
			// returnType: Expression
			ExpressionType,
			// statements: List(Statement)
			&ListType{Element: StatementType},
		}},
		Return: ExpressionType}, 0},
	{"macroExpression", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			&IntType{},
			// name: String
			&StringType{},
			// text: String
			&StringType{},
		}},
		Return: ExpressionType,
	}, 0},
	{"listExpression", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			&IntType{},
			&ListType{Element: ExpressionType},
		}},
		Return: ExpressionType,
	}, 0},
	{"tupleExpression", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			&IntType{},
			&ListType{Element: ExpressionType}}},
		Return: ExpressionType,
	}, 0},
	// TODO: tuple-pattern-matching variableDeclaration
	{"variableDeclaration", &FnType{
		Argument: &TupleType{Elements: []TypeValue{&StringType{}, ExpressionType, BlockType}},
		Return:   StatementType,
	}, 0},
	{"assignStatement", &FnType{
		Argument: &TupleType{Elements: []TypeValue{&StringType{}, BlockType}},
		Return:   StatementType,
	}, 0},
	{"branchStatement", &FnType{
		Argument: &TupleType{Elements: []TypeValue{ExpressionType, BlockType, BlockType}},
		Return:   StatementType,
	}, 0},
	{"isBranchStatement", &FnType{
		Argument: &TupleType{Elements: []TypeValue{
			ExpressionType, &StringType{}, ExpressionType, BlockType, BlockType,
		}},
		Return: StatementType,
	}, 0},
	{"expressionStatement", &FnType{Argument: ExpressionType, Return: StatementType}, 0},
	{"panic", &FnType{
		Argument: &StringType{},
		// empty union type is non-constructable and Union(Union(), T, U) == Union(T, U)
		Return: &UnionType{},
	}, 0},
	{"printlnString", &FnType{
		Argument: &StringType{},
		Return:   &TupleType{},
	}, IMPURE_FUNCTION},
	{"len", &FnType{
		// also works for lists, but this cannot be expressed in Yune
		Argument: &StringType{},
		Return:   &IntType{},
	}, 0},
	// appends an element to a list of any type
	{"append", &FnType{Argument: &UnionType{}, Return: &UnionType{}}, 0},
	{"subString", &FnType{
		Argument: &TupleType{Elements: []TypeValue{&StringType{}, &IntType{}, &IntType{}}},
		Return:   &StringType{},
	}, 0},
	// extracts an element from a list
	{"get", &FnType{
		// (list: List(<any type>), index: Int)
		Argument: &TupleType{Elements: []TypeValue{&UnionType{}, &IntType{}}},
		// <element type>
		Return: &UnionType{},
	}, 0},
	// overwrites an element in a list
	{"set", &FnType{
		// (list: List(<element type>), index: Int, element: <element type>)
		Argument: &TupleType{Elements: []TypeValue{&UnionType{}, &IntType{}, &UnionType{}}},
		// ()
		Return: &TupleType{},
	}, 0},
	{"Union", &FnType{
		Argument: &ListType{Element: &TypeType{}},
		Return:   &TypeType{},
	}, 0},
	// actual signature is `inject(Any): Expression`, but this cannot be expressed in Yune
	{"inject", &FnType{Argument: &TupleType{}, Return: ExpressionType}, 0},
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
