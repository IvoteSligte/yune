package ast

import "log"

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

type Declaration interface {
	Node
	GetName() string
	GetType() InferredType
}

var _ Declaration = &FunctionDeclaration{}
var _ Declaration = &FunctionParameter{}
var _ Declaration = &ConstantDeclaration{}
var _ Declaration = &VariableDeclaration{}
