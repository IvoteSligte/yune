package ast

var IntDeclaration = BuiltinDeclaration{
	Name:         "Int",
	InferredType: TypeType,
	Value:        InferredType{name: "Int"},
	Raw:          "typedef Int int;",
}
var FloatDeclaration = BuiltinDeclaration{
	Name:         "Float",
	InferredType: TypeType,
	Value:        InferredType{name: "Float"},
	Raw:          "typedef Float float;",
}
var BoolDeclaration = BuiltinDeclaration{
	Name:         "Bool",
	InferredType: TypeType,
	Value:        InferredType{name: "Bool"},
	Raw:          "typedef Bool bool;",
}
var NilDeclaration = BuiltinDeclaration{
	Name:         "Nil",
	InferredType: TypeType,
	Value:        InferredType{name: "Nil"},
	Raw:          "typedef Nil void;",
}

var BuiltinDeclarations = map[string]TopLevelDeclaration{
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

var TypeType = InferredType{name: "Type"}
var IntType = IntDeclaration.Value.(InferredType)
var FloatType = FloatDeclaration.Value.(InferredType)
var BoolType = BoolDeclaration.Value.(InferredType)
var NilType = NilDeclaration.Value.(InferredType)
