package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"
)

var BuiltinDeclarations = []TopLevelDeclaration{
	&IntDeclaration,
	&FloatDeclaration,
	&BoolDeclaration,
	&StringDeclaration,
	&ListDeclaration,
	&FnDeclaration,
	&TypeDeclaration,
	&ExpressionDeclaration,
	&StringLiteralDeclaration,
	&PrintStringDeclaration,
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
	return &TypeType{}
}

// LowerDeclaration implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) LowerDeclaration() cpp.Declaration {
	fields := util.Map(b.Fields, func(f BuiltinFieldDeclaration) string {
		return fmt.Sprintf("%s %s;", f.Name, f.Type)
	})
	return fmt.Sprintf("struct %s %s;", b.Name, cpp.Block(fields))
}

// LowerDefinition implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) LowerDefinition() cpp.Definition {
	return ""
}

// Analyze implements TopLevelDeclaration.
func (b *BuiltinStructDeclaration) Analyze(anal Analyzer) {
	// NOTE: probably needs to Analyze parameters/returnType for ordering purposes
	anal.Declare(b)
	anal.Define(b)
}

var _ TopLevelDeclaration = (*BuiltinStructDeclaration)(nil)

type BuiltinFieldDeclaration struct {
	Name string
	Type string
}

func builtinTypeDeclaration(name string, value Expression) ConstantDeclaration {
	return ConstantDeclaration{
		Name: Name{String: name},
		Type: Type{Expression: &Variable{Name: Name{String: "Type"}}},
		Body: Block{Statements: []Statement{&ExpressionStatement{value}}},
	}
}

type BuiltinConstantDeclaration struct {
	Name  string
	Type  TypeValue
	Value string
}

// Analyze implements TopLevelDeclaration.
func (b *BuiltinConstantDeclaration) Analyze(anal Analyzer) {
	// NOTE: probably needs to Analyze parameters/returnType for ordering purposes
	anal.Declare(b)
	anal.Define(b)
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

// LowerDeclaration implements TopLevelDeclaration.
func (d BuiltinConstantDeclaration) LowerDeclaration() cpp.Declaration {
	return fmt.Sprintf("extern %s %s;", d.Type.Lower(), d.Name)
}

// LowerDefinition implements TopLevelDeclaration.
func (d BuiltinConstantDeclaration) LowerDefinition() cpp.Definition {
	return fmt.Sprintf("%s %s = %s;", d.Type.Lower(), d.Name, d.Value)
}

var _ TopLevelDeclaration = (*BuiltinConstantDeclaration)(nil)

func unitStruct(name string) *StructExpression {
	return &StructExpression{Name: Name{String: name}}
}

var TypeDeclaration = ConstantDeclaration{
	Name: Name{String: "Type"},
	Type: Type{value: &TypeType{}},
	Body: Block{Statements: []Statement{&ExpressionStatement{unitStruct("TypeType")}}},
}
var IntDeclaration = builtinTypeDeclaration("Int", unitStruct("IntType"))
var FloatDeclaration = BuiltinConstantDeclaration{
	Name:  "Float",
	Type:  &TypeType{},
	Value: `ty::FloatType{}`,
}
var BoolDeclaration = BuiltinConstantDeclaration{
	Name:  "Bool",
	Type:  &TypeType{},
	Value: `ty::BoolType{}`,
}
var StringDeclaration = BuiltinConstantDeclaration{
	Name:  "String",
	Type:  &TypeType{},
	Value: `ty::StringType{}`,
}
var ExpressionDeclaration = BuiltinConstantDeclaration{
	Name:  "Expression",
	Type:  &TypeType{},
	Value: `box(ty::StructType{"Expression"})`,
}

type BuiltinFunctionDeclaration struct {
	Name       string
	Parameters []BuiltinFunctionParameter
	ReturnType TypeValue
	Body       string
}

// Analyze implements TopLevelDeclaration.
func (b *BuiltinFunctionDeclaration) Analyze(anal Analyzer) {
	// NOTE: probably needs to Analyze parameters/returnType for ordering purposes
	anal.Declare(b)
	anal.Define(b)
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
		argument = &TupleType{Elements: params}
	}
	return &FnType{Argument: argument, Return: b.ReturnType}
}

// LowerDeclaration implements TopLevelDeclaration.
func (d *BuiltinFunctionDeclaration) LowerDeclaration() cpp.Declaration {
	params := util.JoinFunction(d.Parameters, ", ", BuiltinFunctionParameter.Lower)
	return fmt.Sprintf(`struct %s_ {
    %s operator()(%s) const;
    std::string serialize() const;
} %s;`, d.Name, d.ReturnType.Lower(), params, d.Name)
}

// LowerDefinition implements TopLevelDeclaration.
func (d *BuiltinFunctionDeclaration) LowerDefinition() cpp.Definition {
	params := util.JoinFunction(d.Parameters, ", ", BuiltinFunctionParameter.Lower)
	return fmt.Sprintf(`%s %s_::operator()(%s) const %s
std::string %s_::serialize() const {
    return R"({ "Function": "%s" })";
}`, d.ReturnType.Lower(), d.Name, params, "{"+d.Body+"\n}", d.Name, d.Name)
}

var FnDeclaration = BuiltinFunctionDeclaration{
	Name: "Fn",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "argumentType",
			Type: &TypeType{},
		},
		{
			Name: "returnType",
			Type: &TypeType{},
		},
	},
	ReturnType: &TypeType{},
	Body: `return box((ty::FnType){
    .argument = std::move(argumentType),
    .returnType = std::move(returnType),
});`,
}

var ListDeclaration = BuiltinFunctionDeclaration{
	Name: "List",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "elementType",
			Type: &TypeType{},
		},
	},
	ReturnType: &TypeType{},
	Body:       `return box((ty::ListType){ .element = std::move(elementType) });`,
}

var PrintStringDeclaration = BuiltinFunctionDeclaration{
	Name: "printlnString",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "string",
			Type: &StringType{},
		},
	},
	ReturnType: &TupleType{},
	Body: `std::cout << string << std::endl;
return std::make_tuple();`,
}

var StringLiteralDeclaration = BuiltinFunctionDeclaration{
	Name: "stringLiteral",
	Parameters: []BuiltinFunctionParameter{
		{Name: "str", Type: &StringType{}},
	},
	ReturnType: ExpressionType,
	Body:       `return ty::StringLiteral { .value = str };`,
}

var _ TopLevelDeclaration = (*BuiltinFunctionDeclaration)(nil)

type BuiltinFunctionParameter struct {
	Name string
	Type TypeValue
}

func (b BuiltinFunctionParameter) Lower() cpp.Declaration {
	return b.Type.Lower() + " " + b.Name
}
