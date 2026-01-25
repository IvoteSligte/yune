package ast

import (
	"fmt"
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
	IntTypeDeclaration,
	FloatTypeDeclaration,
	BoolTypeDeclaration,
	StringTypeDeclaration,
	NilTypeDeclaration,
	ListTypeDeclaration,
	FnTypeDeclaration,
	ListDeclaration,
	FnDeclaration,
}

// NOTE: main() returns int for compatibility with C++,
// though this may change in the future
var MainType = value.Type("std::function<Int()>")
var TypeType = value.Type("Type")
var IntType = IntDeclaration.GetType()
var FloatType = FloatDeclaration.GetType()
var BoolType = BoolDeclaration.GetType()
var StringType = StringDeclaration.GetType()
var NilType = NilDeclaration.GetType()

type BuiltinStructDeclaration struct {
	Name   string
	Fields []BuiltinFieldDeclaration
}

// GetName implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetName() string {
	return b.Name
}

// GetSpan implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetSpan() Span {
	return Span{}
}

// GetType implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetType() value.Type {
	return TypeType
}

// GetTypeDependencies implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetTypeDependencies() []*Type {
	return []*Type{}
}

// GetValueDependencies implements TopLevelDeclaration.
func (b BuiltinStructDeclaration) GetValueDependencies() []string {
	return []string{}
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

var TypeDeclaration = BuiltinStructDeclaration{
	Name: "Type",
}
var IntTypeDeclaration = BuiltinStructDeclaration{
	Name: "IntType",
}
var FloatTypeDeclaration = BuiltinStructDeclaration{
	Name: "FloatType",
}
var BoolTypeDeclaration = BuiltinStructDeclaration{
	Name: "BoolType",
}
var StringTypeDeclaration = BuiltinStructDeclaration{
	Name: "StringType",
}
var NilTypeDeclaration = BuiltinStructDeclaration{
	Name: "NilType",
}
var ListTypeDeclaration = BuiltinStructDeclaration{
	Name: "ListType",
	Fields: []BuiltinFieldDeclaration{
		{Name: "elementType", Type: "Type"},
	},
}
var FnTypeDeclaration = BuiltinStructDeclaration{
	Name: "FnType",
	Fields: []BuiltinFieldDeclaration{
		{Name: "argumentType", Type: "Type"},
		{Name: "returnType", Type: "Type"},
	},
}

type BuiltinConstantDeclaration struct {
	Name  string
	Type  string
	Value string
}

// GetName implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetName() string {
	return b.Name
}

// GetSpan implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetSpan() Span {
	return Span{}
}

// GetType implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetType() value.Type {
	return value.Type(b.Type)
}

// GetTypeDependencies implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetTypeDependencies() []*Type {
	return []*Type{}
}

// GetValueDependencies implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) GetValueDependencies() []string {
	return []string{}
}

// Lower implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.ConstantDeclaration{
		Name:  b.Name,
		Type:  cpp.Type(b.Type),
		Value: cpp.RawCpp(b.Value),
	}
}

// TypeCheckBody implements TopLevelDeclaration.
func (b BuiltinConstantDeclaration) TypeCheckBody(deps DeclarationTable) (errors []error) {
	return
}

var _ TopLevelDeclaration = BuiltinConstantDeclaration{}

var IntDeclaration = BuiltinConstantDeclaration{
	Name:  "Int",
	Type:  TypeDeclaration.Name,
	Value: "IntType{}",
}
var FloatDeclaration = BuiltinConstantDeclaration{
	Name:  "Float",
	Type:  TypeDeclaration.Name,
	Value: "FloatType{}",
}
var BoolDeclaration = BuiltinConstantDeclaration{
	Name:  "Bool",
	Type:  TypeDeclaration.Name,
	Value: "BoolType{}",
}
var StringDeclaration = BuiltinConstantDeclaration{
	Name:  "String",
	Type:  TypeDeclaration.Name,
	Value: "StringType{}",
}
var NilDeclaration = BuiltinConstantDeclaration{
	Name:  "Nil",
	Type:  TypeDeclaration.Name,
	Value: "NilType{}",
}

type BuiltinFunctionDeclaration struct {
	Name       string
	Parameters []BuiltinFunctionParameter
	ReturnType string
	Body       string
}

// GetName implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetName() string {
	return b.Name
}

// GetSpan implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetSpan() Span {
	return Span{}
}

// GetType implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetType() value.Type {
	params := util.JoinFunction(b.Parameters, ", ", func(p BuiltinFunctionParameter) string { return p.Type })
	return value.Type(fmt.Sprintf("std::function<%s(%s)>", b.ReturnType, params))
}

// GetTypeDependencies implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetTypeDependencies() []*Type {
	return []*Type{}
}

// GetValueDependencies implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) GetValueDependencies() []string {
	return []string{}
}

// Lower implements TopLevelDeclaration.
func (b BuiltinFunctionDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.FunctionDeclaration{
		Name: b.Name,
		Parameters: util.Map(b.Parameters, func(p BuiltinFunctionParameter) cpp.FunctionParameter {
			return cpp.FunctionParameter{
				Name: p.Name,
				Type: cpp.Type(p.Type),
			}
		}),
		ReturnType: cpp.Type(b.ReturnType),
		Body: cpp.Block(util.Map(strings.Split(b.Body, "\n"), func(s string) cpp.Statement {
			return cpp.Statement(cpp.RawCpp(s))
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
			Type: "Type",
		},
		{
			Name: "returnType",
			Type: "Type",
		},
	},
	ReturnType: "FnType",
	Body:       `return FnType{argumentType, returnType};`,
}

var ListDeclaration = BuiltinFunctionDeclaration{
	Name: "List",
	Parameters: []BuiltinFunctionParameter{
		{
			Name: "elementType",
			Type: "Type",
		},
	},
	ReturnType: "ListType",
	Body:       `return ListType{elementType};`,
}

var _ TopLevelDeclaration = BuiltinFunctionDeclaration{}

type BuiltinFunctionParameter struct {
	Name string
	Type string
}
