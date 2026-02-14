package ast

import (
	"log"
)

type DeclarationTable struct {
	parent               *DeclarationTable
	topLevelDeclarations map[string]TopLevelDeclaration
	localDeclarations    map[string]Declaration
	// Callback to be called when a variable cannot be found in the current scope.
	callback func(Name)
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

func (table DeclarationTable) NewScope(callback func(Name)) DeclarationTable {
	return DeclarationTable{
		parent:               &table,
		topLevelDeclarations: table.topLevelDeclarations,
		localDeclarations:    map[string]Declaration{},
		callback:             callback,
	}
}

func (table *DeclarationTable) Get(name Name) (Declaration, bool) {
	if table == table.parent {
		log.Panicf("Table at address %p has itself as parent.", table)
	}
	local, ok := table.localDeclarations[name.String]
	if !ok && table.parent != nil {
		table.callback(name)
		return table.parent.Get(name)
	}
	if !ok {
		topLevel, ok := table.topLevelDeclarations[name.String]
		return topLevel, ok
	}
	return local, ok
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
