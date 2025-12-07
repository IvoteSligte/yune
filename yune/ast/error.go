package ast

import "fmt"

type DuplicateDeclaration struct {
	Name string
}

func (e DuplicateDeclaration) Error() string {
	return fmt.Sprintf("%s was already defined.", e.Name)
}

type UndefinedVariable struct {
	Name string
}

func (e UndefinedVariable) Error() string {
	return fmt.Sprintf("Variable %s is not defined.", e.Name)
}

type InvalidUnaryExpressionType struct {
	Op   UnaryOp
	Type Type
}

func (e InvalidUnaryExpressionType) Error() string {
	return fmt.Sprintf("Cannot apply unary operator %s to type %s.", e.Op, e.Type)
}
