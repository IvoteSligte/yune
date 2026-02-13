package ast

import (
	"log"
)

type DeclarationTable struct {
	parent               *DeclarationTable
	topLevelDeclarations map[string]TopLevelDeclaration
	localDeclarations    map[string]Declaration
}

func (table *DeclarationTable) Add(decl Declaration) error {
	prev, exists := table.localDeclarations[decl.GetName().String]
	if exists {
		return DuplicateDeclaration{
			First:  prev,
			Second: decl,
		}
	}
	if table.localDeclarations == nil {
		table.localDeclarations = map[string]Declaration{}
	}
	table.localDeclarations[decl.GetName().String] = decl
	return nil
}

func (table DeclarationTable) NewScope() DeclarationTable {
	return DeclarationTable{
		parent:               &table,
		topLevelDeclarations: table.topLevelDeclarations,
		localDeclarations:    map[string]Declaration{},
	}
}

func (table *DeclarationTable) Get(name string) (Declaration, bool) {
	if table == table.parent {
		log.Panicf("Table at address %p has itself as parent.", table)
	}
	local, ok := table.localDeclarations[name]
	if !ok && table.parent != nil {
		return table.parent.Get(name)
	}
	if !ok {
		topLevel, ok := table.topLevelDeclarations[name]
		return topLevel, ok
	}
	return local, ok
}

func (table DeclarationTable) GetTopLevel(name string) (topLevel TopLevelDeclaration, ok bool) {
	topLevel, ok = table.topLevelDeclarations[name]
	return
}

func (table DeclarationTable) TopLevel() DeclarationTable {
	if table.parent != nil {
		return table.parent.TopLevel()
	}
	return table
}

type Declaration interface {
	Node
	GetName() Name

	// Returns the type of this declaration.
	GetDeclaredType() TypeValue
}

var _ Declaration = &FunctionDeclaration{}
var _ Declaration = &FunctionParameter{}
var _ Declaration = &ConstantDeclaration{}
var _ Declaration = &VariableDeclaration{}
