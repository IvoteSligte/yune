package ast

import "yune/cpp"

type BuiltinDeclaration struct {
	cpp.Type
	Name  string
	Value cpp.Node
}

var IntDeclaration = BuiltinDeclaration{
	Name:  "Int",
	Value: cpp.Type{Name: "int"},
	Type:  TypeType,
}
var FloatDeclaration = BuiltinDeclaration{
	Name:  "Float",
	Type:  TypeType,
	Value: cpp.Type{Name: "float"},
}
var BoolDeclaration = BuiltinDeclaration{
	Name:  "Bool",
	Type:  TypeType,
	Value: cpp.Type{Name: "bool"},
}
var NilDeclaration = BuiltinDeclaration{
	Name:  "Nil",
	Type:  TypeType,
	Value: cpp.Type{Name: "void"},
}

var BuiltinDeclarations = map[string]Declaration{
	"Int":   IntDeclaration,
	"Float": FloatDeclaration,
	"Bool":  BoolDeclaration,
	"Nil":   NilDeclaration,
}

var BuiltinNames = []string{
	IntDeclaration.Name,
	FloatDeclaration.Name,
	BoolDeclaration.Name,
	NilDeclaration.Name,
}

var TypeType = cpp.Type{Name: "Type"}
var IntType = IntDeclaration.Value.(cpp.Type)
var FloatType = FloatDeclaration.Value.(cpp.Type)
var BoolType = BoolDeclaration.Value.(cpp.Type)
var NilType = NilDeclaration.Value.(cpp.Type)

// GetType implements Declaration.
func (d BuiltinDeclaration) GetType() cpp.Type {
	return d.Type
}

// GetValue implements TopLevelDeclaration.
func (d BuiltinDeclaration) GetValue() cpp.TopLevelDeclaration {
	return d.Value
}

// Lower implements TopLevelDeclaration.
func (d BuiltinDeclaration) Lower() cpp.TopLevelDeclaration {
	return d.Value
}

// CalcType implements Declaration.
func (d BuiltinDeclaration) CalcType(deps DeclarationTable) (errors Errors) {
	return
}

// GetTypeDependencies implements Declaration.
func (d BuiltinDeclaration) GetTypeDependencies() (deps []string) {
	return
}

// GetValueDependencies implements Declaration.
func (d BuiltinDeclaration) GetValueDependencies() (deps []string) {
	return
}

// TypeCheckBody implements Declaration.
func (d BuiltinDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	return
}

func (d BuiltinDeclaration) GetName() string {
	return d.Name
}

func (d BuiltinDeclaration) GetSpan() Span {
	return Span{}
}

func (d BuiltinDeclaration) InferType(DeclarationTable) (errors Errors) {
	return
}
