package ast

import (
	"strings"
	"yune/cpp"
	"yune/util"
	"yune/value"
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
	Type           value.Type
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
func (b BuiltinRawDeclaration) GetDeclaredType() value.Type {
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

var TypeDeclaration = BuiltinRawDeclaration{
	Name: "Type",
	Type: value.TypeType,
	Header: `
struct Type {
    std::string id;
};`,
	Implementation: `
std::ostream& operator<<(std::ostream& out, const Type& t) {
    return out << t.id;
}`,
}

var ExpressionDeclaration = BuiltinRawDeclaration{
	Name:     "Expression",
	Type:     value.TypeType,
	Requires: []string{"Type"},
	Header: `
extern Type Expression;

struct Expression_type_ {
    std::string expr;
};`,
	Implementation: `
Type Expression = Type{"Expression_type_"};

// TODO: Lisp/JSON style serialization or something
std::ostream& operator<<(std::ostream& out, const Expression_type_& e) {
    return out << e.expr;
}

// TEMP (macro return type)
std::ostream& operator<<(std::ostream& out, const std::tuple<std::string, Expression_type_>& t) {
    return out << std::get<0>(t) << "; " << std::get<1>(t);
}`,
}

var StringLiteralDeclaration = BuiltinRawDeclaration{
	Name:     "stringLiteral",
	Type:     value.NewFnType(value.StringType, value.ExpressionType),
	Requires: []string{"Expression"},
	Implementation: `
Expression_type_ stringLiteral(std::string str) {
    return Expression_type_{str};
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
func (b BuiltinStructDeclaration) GetDeclaredType() value.Type {
	return value.TypeType
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

// var TypeDeclaration = BuiltinStructDeclaration{
// 	Name: "Type",
// 	Fields: []BuiltinFieldDeclaration{
// 		{Name: "id", Type: "std::string"},
// 	},
// }

type BuiltinConstantDeclaration struct {
	Name  string
	Type  value.Type
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
func (b BuiltinConstantDeclaration) GetDeclaredType() value.Type {
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

var IntDeclaration = BuiltinConstantDeclaration{
	Name:  "Int",
	Type:  value.TypeType,
	Value: `Type{"int"}`,
}
var FloatDeclaration = BuiltinConstantDeclaration{
	Name:  "Float",
	Type:  value.TypeType,
	Value: `Type{"float"}`,
}
var BoolDeclaration = BuiltinConstantDeclaration{
	Name:  "Bool",
	Type:  value.TypeType,
	Value: `Type{"bool"}`,
}
var StringDeclaration = BuiltinConstantDeclaration{
	Name:  "String",
	Type:  value.TypeType,
	Value: `Type{"std::string"}`,
}
var NilDeclaration = BuiltinConstantDeclaration{
	Name:  "Nil",
	Type:  value.TypeType,
	Value: `Type{"void"}`,
}

type BuiltinFunctionDeclaration struct {
	Name       string
	Parameters []BuiltinFunctionParameter
	ReturnType value.Type
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
func (b BuiltinFunctionDeclaration) GetDeclaredType() value.Type {
	// NOTE: does this work for single parameters? it's the same in FunctionDeclaration.GetDeclaredType
	params := util.Map(b.Parameters, func(p BuiltinFunctionParameter) value.Type { return p.Type })
	return value.NewFnType(value.NewTupleType(params...), b.ReturnType)
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
			Type: value.TypeType,
		},
		{
			Name: "returnType",
			Type: value.TypeType,
		},
	},
	ReturnType: value.TypeType,
	// FIXME: this does not map Fn((A, B), C) -> std::function<C(A, B)> but to std::function<C(std::tuple<A, B>)>
	Body: `
std::string tuplePrefix("std::tuple<");
if (argumentType.id.substr(0, tuplePrefix.size()) == tuplePrefix) {
    argumentType = Type{argumentType.id.substr(
        tuplePrefix.size(),
        argumentType.id.size() - 1 - tuplePrefix.size()
    )};
}
return Type{"std::function<" + returnType.id + "(" + argumentType.id + ")>"};
`,
}

var ListDeclaration = BuiltinFunctionDeclaration{
	Name: "List",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "elementType",
			Type: value.TypeType,
		},
	},
	ReturnType: value.TypeType,
	Body:       `return Type{"std::vector<" + elementType.id + ">"};`,
}

var PrintStringDeclaration = BuiltinFunctionDeclaration{
	Name: "printlnString",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "string",
			Type: value.StringType,
		},
	},
	ReturnType: value.NilType,
	Body:       `std::cout << string << std::endl;`,
}

var _ TopLevelDeclaration = BuiltinFunctionDeclaration{}

type BuiltinFunctionParameter struct {
	Name string
	Type value.Type
}
