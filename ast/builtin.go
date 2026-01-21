package ast

import "yune/cpp"

// TODO: just replace with ConstantDeclaration or maybe TypeDeclaration,
// since types need to be handled specially in the Lower function as well
// due to `typedef A <expr>` not being valid C++. In general we need a way to
// evaluate expressions used for types.
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
var StringDeclaration = BuiltinTypeDeclaration{
	cpp.TypeAlias{
		Alias: "String",
		Of:    "std::string",
	},
}
var NilDeclaration = BuiltinTypeDeclaration{
	cpp.TypeAlias{
		Alias: "Nil",
		Of:    "void",
	},
}

type BuiltinFunctionDeclaration struct {
	cpp.FunctionDeclaration
}

////  String range(Int, Int):
// var StringRangeDeclaration = BuiltinFunctionDeclaration {
// 	cpp.FunctionDeclaration
// };

var BuiltinDeclarations = map[string]Declaration{
	IntDeclaration.GetName():    IntDeclaration,
	FloatDeclaration.GetName():  FloatDeclaration,
	BoolDeclaration.GetName():   BoolDeclaration,
	StringDeclaration.GetName(): StringDeclaration,
	NilDeclaration.GetName():    NilDeclaration,
}

// NOTE: main() returns int for compatibility with C++,
// though this may change in the future
var MainType = cpp.FunctionType{
	Parameter:  cpp.TupleType{},
	ReturnType: IntType,
}
var TypeType = cpp.NamedType{Name: "Type"}
var IntType = IntDeclaration.Get()
var FloatType = FloatDeclaration.Get()
var BoolType = BoolDeclaration.Get()
var StringType = StringDeclaration.Get()
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
