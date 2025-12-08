package ast

import "yune/util"

type Module struct {
	Declarations []TopLevelDeclaration
}

type TopLevelDeclaration interface {
	Declaration
	topLevelDeclaration()
}

type FunctionDeclaration struct {
	Name       string
	Parameters []FunctionParameter
	ReturnType Type
	Body       []Statement
}

func (FunctionDeclaration) topLevelDeclaration() {}

func (d FunctionDeclaration) GetName() string {
	return d.Name
}

func (d FunctionDeclaration) GetDeclarationType() Type {
	return Type{
		Name:     "Fn",
		Generics: append(util.Map(d.Parameters, FunctionParameter.GetDeclarationType), d.ReturnType),
	}
}

type FunctionParameter struct {
	Name string
	Type Type
}

func (d FunctionParameter) GetName() string {
	return d.Name
}

func (d FunctionParameter) GetDeclarationType() Type {
	return d.Type
}

type ConstantDeclaration struct {
	Name string
	Type Type
	Body []Statement
}

func (ConstantDeclaration) topLevelDeclaration() {}

func (d ConstantDeclaration) GetName() string {
	return d.Name
}

func (d ConstantDeclaration) GetDeclarationType() Type {
	return d.Type
}
