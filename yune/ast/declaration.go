package ast

import (
	"log"
	"yune/cpp"
)

type DeclarationTable struct {
	parent       *DeclarationTable
	declarations map[string]Declaration
}

func (table *DeclarationTable) Add(decl Declaration) {
	_, exists := table.declarations[decl.GetName()]
	if exists {
		log.Fatalf("Duplicate declaration of %s in the same scope.", decl.GetName()) // TODO: handle properly
	}
	table.declarations[decl.GetName()] = decl
}

func (table *DeclarationTable) NewScope() DeclarationTable {
	return DeclarationTable{
		parent:       table,
		declarations: map[string]Declaration{},
	}
}

func (table *DeclarationTable) Get(name string) (Declaration, bool) {
	declaration, ok := table.declarations[name]
	if !ok && table.parent != nil {
		return table.parent.Get(name)
	}
	return declaration, true
}

type BuiltinDeclaration struct {
	InferredType
	Name string
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

type Declaration interface {
	Node
	GetName() string
	GetType() InferredType
	CalcType(table DeclarationTable) (errors Errors)
	TypeCheck(table DeclarationTable) (errors Errors)
	Lower() cpp.Declaration
}

var _ Declaration = &FunctionDeclaration{}
var _ Declaration = &FunctionParameter{}
var _ Declaration = &ConstantDeclaration{}
var _ Declaration = &VariableDeclaration{}
var _ Declaration = BuiltinDeclaration{}
