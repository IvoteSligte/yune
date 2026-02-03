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
	NilDeclaration,
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

// GetMacros implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) GetMacros() []*Macro {
	return []*Macro{}
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

// GetMacroTypeDependencies implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetMacroValueDependencies implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) GetMacroValueDependencies() []Name {
	return []Name{}
}

// GetTypeDependencies implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) GetValueDependencies() []Name {
	return util.Map(b.Requires, func(s string) Name { return Name{String: s} })
}

// Lower implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.RawDeclaration{
		Header:         b.Header,
		Implementation: b.Implementation,
	}
}

// TypeCheckBody implements TopLevelDeclaration.
func (b BuiltinRawDeclaration) TypeCheckBody(deps DeclarationTable) (errors []error) {
	return
}

var _ TopLevelDeclaration = BuiltinRawDeclaration{}

var StringLiteralDeclaration = BuiltinRawDeclaration{
	Name:     "stringLiteral",
	Type:     FnType{Argument: StringType{}, Return: ExpressionType},
	Requires: []string{"Expression"},
	Implementation: `
ty::Expression stringLiteral(std::string str) {
    return ty::StringLiteral(str);
};`,
}

type BuiltinStructDeclaration struct {
	Name   string
	Fields []BuiltinFieldDeclaration
}

// GetMacros implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetMacros() []*Macro {
	return []*Macro{}
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

// GetMacroTypeDependencies implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetMacroValueDependencies implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetMacroValueDependencies() []Name {
	return []Name{}
}

// GetTypeDependencies implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetValueDependencies() []Name {
	return []Name{}
}

// Lower implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.StructDeclaration{
		Name: b.Name,
		Fields: util.Map(b.Fields, func(f BuiltinFieldDeclaration) cpp.Field {
			return cpp.Field{Name: f.Name, Type: cpp.Type(f.Type)}
		}),
	}
}

// TypeCheckBody implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) TypeCheckBody(deps DeclarationTable) (errors []error) {
	return
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

// GetMacros implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetMacros() []*Macro {
	return []*Macro{}
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

// GetMacroTypeDependencies implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetMacroValueDependencies implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetMacroValueDependencies() []Name {
	return []Name{}
}

// GetTypeDependencies implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetValueDependencies() []Name {
	return []Name{}
}

// Lower implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.ConstantDeclaration{
		Name:  b.Name,
		Type:  b.Type.Lower(),
		Value: cpp.Raw(b.Value),
	}
}

// TypeCheckBody implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) TypeCheckBody(deps DeclarationTable) (errors []error) {
	return
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
var NilDeclaration = BuiltinConstantDeclaration{
	Name:  "Nil",
	Type:  TypeType{},
	Value: `ty::NilType{}`,
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

// GetMacros implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetMacros() []*Macro {
	return []*Macro{}
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

// GetMacroTypeDependencies implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetMacroTypeDependencies() []Query {
	return []Query{}
}

// GetMacroValueDependencies implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetMacroValueDependencies() []Name {
	return []Name{}
}

// GetTypeDependencies implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetTypeDependencies() []Query {
	return []Query{}
}

// GetValueDependencies implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetValueDependencies() []Name {
	return []Name{}
}

// Lower implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.FunctionDeclaration{
		Name: b.Name,
		Parameters: util.Map(b.Parameters, func(p BuiltinFunctionParameter) cpp.FunctionParameter {
			return cpp.FunctionParameter{
				Name: p.Name,
				Type: p.Type.Lower(),
			}
		}),
		ReturnType: b.ReturnType.Lower(),
		Body: cpp.Block(util.Map(strings.Split(b.Body, "\n"), func(s string) cpp.Statement {
			return cpp.Statement(cpp.Raw(s))
		})),
	}
}

// TypeCheckBody implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) TypeCheckBody(deps DeclarationTable) (errors []error) {
	return
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
	Body:       `return box(ty::FnType(std::move(argumentType), std::move(returnType)));`,
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
	Body:       `return box(ty::ListType(std::move(elementType)));`,
}

var PrintStringDeclaration = BuiltinFunctionDeclaration{
	Name: "printlnString",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "string",
			Type: StringType{},
		},
	},
	ReturnType: NilType{},
	Body:       `std::cout << string << std::endl;`,
}

var _ TopLevelDeclaration = BuiltinFunctionDeclaration{}

type BuiltinFunctionParameter struct {
	Name string
	Type TypeValue
}
