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

// Evaluate a lowered Expression, assuming that Expression.Analyze has already been called on it.
func (a Analyzer) EvaluateLowered(lowered string) (json string) {
	json, err := cpp.Repl.Evaluate(lowered)
	if err != nil {
		panic("Failed to evaluate lowered expression. Error: " + err.Error())
	}
	return
}

// Evaluate an Expression, assuming that Expression.Analyze has already been called on it.
func (a Analyzer) Evaluate(expr Expression) (json string) {
	return a.EvaluateLowered(expr.Lower())
}

func (a Analyzer) NewScope() Analyzer {
	a.Table = a.Table.NewScope()
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
	// Non-top-level declarations are analyzed in sequential order,
	// so this type should already be available.
	_type := decl.GetDeclaredType()
	if _type == nil {
		panic("Declaration.GetDeclaredType() returned nil on local declaration '" + name.String + "'")
	}
	return _type
}

// NOTE: probably want top-level declarations to declare their prototypes as soon as those are known,
// and their full definitions after when they have been type checked

func (a Analyzer) Declare(decl TopLevelDeclaration) {
	err := cpp.Repl.Declare(decl.LowerDeclaration())
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
	err := cpp.Repl.Declare(decl.LowerDefinition())
	if err != nil {
		panic("Failed to define declaration " + decl.GetName().String)
	}
}
