package ast

type Module struct {
	declarations []TopLevelDeclaration
}

type TopLevelDeclaration any
