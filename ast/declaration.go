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
	if table.declarations == nil {
		table.declarations = map[string]Declaration{}
		table.declarations[decl.GetName()] = decl
	}
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
	GetName() string

	// --- compilation stage 1 ---

	// Queries the names of types used in this declaration, including in the body.
	GetTypeDependencies() []string
	// Calculates the declaration's type, but does not touch the body.
	CalcType(deps DeclarationTable) Errors
	// Returns the calculated type.
	GetType() cpp.Type

	// --- compilation stage 2 ---

	// Queries the names of constants such as global variables and functions
	// used in this declaration's body.
	GetValueDependencies() []string
	// Type checks the declaration's body, possibly resulting in errors.
	// Assumes the declaration's type has been calculated.
	TypeCheckBody(deps DeclarationTable) (errors Errors)
}

var _ Declaration = &FunctionDeclaration{}
var _ Declaration = &FunctionParameter{}
var _ Declaration = &ConstantDeclaration{}
var _ Declaration = &VariableDeclaration{}
var _ Declaration = BuiltinTypeDeclaration{}
