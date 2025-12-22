package ast

import "yune/cpp"

type BuiltinTypeDeclaration struct {
	cpp.TypeAlias
}

var IntDeclaration = BuiltinTypeDeclaration{
	cpp.TypeAlias{
		Alias: "Int",
		Of:    "int",
	},
}
var FloatDeclaration = BuiltinTypeDeclaration{
	cpp.TypeAlias{
		Alias: "Float",
		Of:    "float",
	},
}
var BoolDeclaration = BuiltinTypeDeclaration{
	cpp.TypeAlias{
		Alias: "Bool",
		Of:    "bool",
	},
}
var NilDeclaration = BuiltinTypeDeclaration{
	cpp.TypeAlias{
		Alias: "Nil",
		Of:    "void",
	},
}

var BuiltinDeclarations = map[string]Declaration{
	IntDeclaration.GetName():   IntDeclaration,
	FloatDeclaration.GetName(): FloatDeclaration,
	BoolDeclaration.GetName():  BoolDeclaration,
	NilDeclaration.GetName():   NilDeclaration,
}

var TypeType = cpp.Type{Name: "Type"}
var IntType = IntDeclaration.Get()
var FloatType = FloatDeclaration.Get()
var BoolType = BoolDeclaration.Get()
var NilType = NilDeclaration.Get()

// GetType implements Declaration.
func (d BuiltinTypeDeclaration) GetType() cpp.Type {
	return TypeType
}

// Lower implements TopLevelDeclaration.
func (d BuiltinTypeDeclaration) Lower() cpp.TopLevelDeclaration {
	return d.TypeAlias
}

// CalcType implements Declaration.
func (d BuiltinTypeDeclaration) CalcType(deps DeclarationTable) (errors Errors) {
	return
}

// GetTypeDependencies implements Declaration.
func (d BuiltinTypeDeclaration) GetTypeDependencies() (deps []string) {
	return
}

// GetValueDependencies implements Declaration.
func (d BuiltinTypeDeclaration) GetValueDependencies() (deps []string) {
	return
}

// TypeCheckBody implements Declaration.
func (d BuiltinTypeDeclaration) TypeCheckBody(deps DeclarationTable) (errors Errors) {
	return
}

func (d BuiltinTypeDeclaration) GetName() string {
	return d.Alias
}

func (d BuiltinTypeDeclaration) GetSpan() Span {
	return Span{}
}

func (d BuiltinTypeDeclaration) InferType(DeclarationTable) (errors Errors) {
	return
}
