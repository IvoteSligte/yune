package ast

import (
	"log"
)

type capture struct {
	name        Name
	declaration Declaration
}

type DeclarationTable struct {
	parent               *DeclarationTable
	topLevelDeclarations map[string]TopLevelDeclaration
	localDeclarations    map[string]Declaration
	captures             *[]capture
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
		captures:             &[]capture{},
	}
}

func (table *DeclarationTable) Get(name Name) (Declaration, bool) {
	if table == table.parent {
		log.Panicf("Table at address %p has itself as parent.", table)
	}
	local, isLocal := table.localDeclarations[name.String]
	if !isLocal && table.parent != nil {
		decl, found := table.parent.Get(name)
		if found {
			*table.captures = append(*table.captures, capture{name, decl})
		}
		return decl, found
	}
	if !isLocal {
		topLevel, found := table.topLevelDeclarations[name.String]
		return topLevel, found
	}
	return local, isLocal
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
