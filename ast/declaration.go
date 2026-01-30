package ast

import (
	"log"
	"yune/pb"
)

type DeclarationTable struct {
	parent       *DeclarationTable
	declarations map[string]Declaration
}

func (table *DeclarationTable) Add(decl Declaration) error {
	prev, exists := table.declarations[decl.GetName().String]
	if exists {
		return DuplicateDeclaration{
			First:  prev,
			Second: decl,
		}
	}
	if table.declarations == nil {
		table.declarations = map[string]Declaration{}
	}
	table.declarations[decl.GetName().String] = decl
	return nil
}

func (table DeclarationTable) NewScope() DeclarationTable {
	return DeclarationTable{
		parent:       &table,
		declarations: map[string]Declaration{},
	}
}

func (table *DeclarationTable) Get(name string) (Declaration, bool) {
	if table == table.parent {
		log.Panicf("Table at address %p has itself as parent.", table)
	}
	declaration, ok := table.declarations[name]
	if !ok && table.parent != nil {
		return table.parent.Get(name)
	}
	return declaration, ok
}

func (table *DeclarationTable) GetTopLevel(name string) (TopLevelDeclaration, bool) {
	if table.parent != nil {
		return table.parent.GetTopLevel(name)
	}
	declaration, ok := table.declarations[name]
	return declaration.(TopLevelDeclaration), ok
}

type Declaration interface {
	Node
	GetName() Name

	GetMacroTypeDependencies() []Query
	GetMacroValueDependencies() []Name

	GetMacros() []*Macro
	// Queries the type expressions used in this declaration, including in the body.
	GetTypeDependencies() []Query
	// Queries the names of constants such as global variables and functions
	// used in this declaration's body.
	GetValueDependencies() []Name

	// Returns the type of this declaration.
	GetDeclaredType() pb.Type

	// Type checks the declaration's body, possibly resulting in errors.
	// Assumes the declaration's type has been calculated.
	TypeCheckBody(deps DeclarationTable) (errors Errors)
}

var _ Declaration = &FunctionDeclaration{}
var _ Declaration = &FunctionParameter{}
var _ Declaration = &ConstantDeclaration{}
var _ Declaration = &VariableDeclaration{}
