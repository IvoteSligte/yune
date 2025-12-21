package ast

var IntDeclaration Declaration = BuiltinDeclaration{Name: "Int", InferredType: TypeType}
var FloatDeclaration Declaration = BuiltinDeclaration{Name: "Float", InferredType: TypeType}
var BoolDeclaration Declaration = BuiltinDeclaration{Name: "Bool", InferredType: TypeType}
var NilDeclaration Declaration = BuiltinDeclaration{Name: "Nil", InferredType: TypeType}

var BuiltinDeclarations map[string]Declaration = map[string]Declaration{
	"Int":   IntDeclaration,
	"Float": FloatDeclaration,
	"Bool":  BoolDeclaration,
	"Nil":   NilDeclaration,
}

var BuiltinNames []string = []string{
	"Int",
	"Float",
	"Bool",
	"Nil",
}

var TypeType InferredType = InferredType{name: "Type"}
var IntType InferredType = InferredType{name: "Int"}
var FloatType InferredType = InferredType{name: "Float"}
var BoolType InferredType = InferredType{name: "Bool"}
var NilType InferredType = InferredType{name: "Nil"}
