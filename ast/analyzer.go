package ast

import (
	"log"
	"yune/cpp"
)

type Analyzer struct {
	Errors          *Errors
	Defined         map[TopLevelDeclaration]struct{}
	Table           DeclarationTable
	GetTypeCallback func(Name)
}

func (a Analyzer) PushError(err error) {
	*a.Errors = append(*a.Errors, err)
	log.Fatalln("Analyzer error: ", err) // TODO: only exit when needed
}

func (a Analyzer) HasErrors() bool {
	return len(*a.Errors) > 0
}

// Evaluate an Expression, assuming that Expression.Analyze has already been called on it.
func (a Analyzer) Evaluate(expr Expression) (json string) {
	lowered := expr.Lower()
	json, err := cpp.Repl.Evaluate(lowered)
	if err != nil {
		panic("Failed to evaluate lowered expression. Error: " + err.Error())
	}
	return
}

func (a Analyzer) NewScope(callback func(Name, Declaration)) Analyzer {
	a.Table = a.Table.NewScope(callback)
	return a
}

func (a Analyzer) GetType(name Name) TypeValue {
	decl, ok := a.Table.Get(name)
	if !ok {
		panic("Unknown declaration: " + name.String)
	}
	if a.GetTypeCallback != nil {
		a.GetTypeCallback(name)
	}
	topLevel, isTopLevel := decl.(TopLevelDeclaration)
	if isTopLevel {
		_, isDone := a.Defined[topLevel]
		if !isDone {
			// Keep only the relevant data for a top-level analyzer.
			topLevel.Analyze(Analyzer{
				Errors:  a.Errors,
				Defined: a.Defined,
				Table: DeclarationTable{
					topLevelDeclarations: a.Table.topLevelDeclarations,
				},
				GetTypeCallback: nil,
			})
		}
	}
	_type := decl.GetDeclaredType()
	if _type == nil {
		panic("Declaration.GetDeclaredType() returned nil on declaration '" + name.String + "'")
	}
	return _type
}

// NOTE: probably want top-level declarations to declare their prototypes as soon as those are known,
// and their full definitions after when they have been type checked

func (a Analyzer) Declare(decl TopLevelDeclaration) {
	err := cpp.Repl.Write(decl.LowerDeclaration())
	if err != nil {
		panic("Failed to declare " + decl.GetName().String)
	}
}

func (a Analyzer) Define(decl TopLevelDeclaration) {
	_, alreadyDefined := a.Defined[decl]
	if alreadyDefined {
		panic("Redefinition of declaration " + decl.GetName().String)
	}
	a.Defined[decl] = struct{}{}
	err := cpp.Repl.Write(decl.LowerDefinition())
	if err != nil {
		panic("Failed to define declaration " + decl.GetName().String)
	}
}
