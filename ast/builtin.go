package ast

import (
	"strings"
	"yune/cpp"
	"yune/util"
)

var BuiltinDeclarations = []TopLevelDeclaration{
	IntDeclaration,
	FloatDeclaration,
	BoolDeclaration,
	StringDeclaration,
	ListDeclaration,
	FnDeclaration,
	TypeDeclaration,
	ExpressionDeclaration,
	StringLiteralDeclaration,
	PrintStringDeclaration,
}

// Declares a type that will exist in the C++ code, but not in the Yune code.
type BuiltinRawDeclaration struct {
	Name           string
	Type           TypeValue
	Requires       []string
	Header         string
	Implementation string
}

// GetName implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) GetName() Name {
	return Name{String: b.Name}
}

// GetSpan implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) GetSpan() Span {
	return Span{}
}

// GetDeclaredType implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) GetDeclaredType() TypeValue {
	return b.Type
}

// Lower implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) Lower() cpp.Declaration {
	return cpp.Declaration{
		Header:         b.Header,
		Implementation: b.Implementation,
	}
}

// Analyze implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) Analyze(anal Analyzer) {
	// NOTE: probably needs to Analyze parameters/returnType for ordering purposes
}

var _ TopLevelDeclaration = BuiltinRawDeclaration{}

var StringLiteralDeclaration = BuiltinRawDeclaration{
	Name:     "stringLiteral",
	Type:     FnType{Argument: StringType{}, Return: ExpressionType},
	Requires: []string{"Expression"},
	Implementation: `
ty::Expression stringLiteral(std::string str) {
    return ty::StringLiteral { .value = str };
};`,
}

type BuiltinStructDeclaration struct {
	Name   string
	Fields []BuiltinFieldDeclaration
}

// GetName implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetName() Name {
	return Name{String: b.Name}
}

// GetSpan implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetSpan() Span {
	return Span{}
}

// GetDeclaredType implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetDeclaredType() TypeValue {
	return TypeType{}
}

// Lower implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) Lower() cpp.Declaration {
	return cpp.StructDeclaration(
		b.Name,
		util.Map(b.Fields, func(f BuiltinFieldDeclaration) string {
			return cpp.NewField(f.Name, f.Type)
		}),
	)
}

// Analyze implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) Analyze(anal Analyzer) {
	// NOTE: probably needs to Analyze parameters/returnType for ordering purposes
}

var _ TopLevelDeclaration = BuiltinStructDeclaration{}

type BuiltinFieldDeclaration struct {
	Name string
	Type string
}

type BuiltinConstantDeclaration struct {
	Name  string
	Type  TypeValue
	Value string
}

// Analyze implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) Analyze(anal Analyzer) {
	// NOTE: probably needs to Analyze parameters/returnType for ordering purposes
}

// GetName implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetName() Name {
	return Name{String: b.Name}
}

// GetSpan implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetSpan() Span {
	return Span{}
}

// GetDeclaredType implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetDeclaredType() TypeValue {
	return b.Type
}

// Lower implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) Lower() cpp.Declaration {
	return cpp.ConstantDeclaration(b.Name, b.Type.Lower(), b.Value)
}

var _ TopLevelDeclaration = BuiltinConstantDeclaration{}

var TypeDeclaration = BuiltinConstantDeclaration{
	Name:  "Type",
	Type:  TypeType{},
	Value: `ty::TypeType{}`,
}
var IntDeclaration = BuiltinConstantDeclaration{
	Name:  "Int",
	Type:  TypeType{},
	Value: `ty::IntType{}`,
}
var FloatDeclaration = BuiltinConstantDeclaration{
	Name:  "Float",
	Type:  TypeType{},
	Value: `ty::FloatType{}`,
}
var BoolDeclaration = BuiltinConstantDeclaration{
	Name:  "Bool",
	Type:  TypeType{},
	Value: `ty::BoolType{}`,
}
var StringDeclaration = BuiltinConstantDeclaration{
	Name:  "String",
	Type:  TypeType{},
	Value: `ty::StringType{}`,
}
var ExpressionDeclaration = BuiltinConstantDeclaration{
	Name:  "Expression",
	Type:  TypeType{},
	Value: `box(ty::StructType{"Expression"})`,
}

type BuiltinFunctionDeclaration struct {
	Name       string
	Parameters []BuiltinFunctionParameter
	ReturnType TypeValue
	Body       string
}

// Analyze implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) Analyze(anal Analyzer) {
	// NOTE: probably needs to Analyze parameters/returnType for ordering purposes
}

// GetName implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetName() Name {
	return Name{String: b.Name}
}

// GetSpan implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetSpan() Span {
	return Span{}
}

// GetDeclaredType implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetDeclaredType() TypeValue {
	params := util.Map(b.Parameters, func(p BuiltinFunctionParameter) TypeValue { return p.Type })
	var argument TypeValue
	if len(b.Parameters) == 1 {
		argument = params[0]
	} else {
		argument = NewTupleType(params...)
	}
	return FnType{Argument: argument, Return: b.ReturnType}
}

// Lower implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) Lower() cpp.Declaration {
	return cpp.FunctionDeclaration(
		registerNode(b),
		b.Name,
		util.Map(b.Parameters, func(p BuiltinFunctionParameter) string {
			return p.Type.Lower() + " " + p.Name
		}),
		b.ReturnType.Lower(),
		cpp.Block(strings.Split(b.Body, "\n")),
	)
}

var FnDeclaration = BuiltinFunctionDeclaration{
	Name: "Fn",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "argumentType",
			Type: TypeType{},
		},
		{
			Name: "returnType",
			Type: TypeType{},
		},
	},
	ReturnType: TypeType{},
	Body: `
return box((ty::FnType){
    .argument = std::move(argumentType),
    .returnType = std::move(returnType),
});`,
}

var ListDeclaration = BuiltinFunctionDeclaration{
	Name: "List",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "elementType",
			Type: TypeType{},
		},
	},
	ReturnType: TypeType{},
	Body:       `return box((ty::ListType){ .element = std::move(elementType) });`,
}

var PrintStringDeclaration = BuiltinFunctionDeclaration{
	Name: "printlnString",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "string",
			Type: StringType{},
		},
	},
	ReturnType: TupleType{},
	Body: `
std::cout << string << std::endl;
return std::make_tuple();`,
}

var _ TopLevelDeclaration = BuiltinFunctionDeclaration{}

type BuiltinFunctionParameter struct {
	Name string
	Type TypeValue
}
