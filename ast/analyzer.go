package ast

import (
	"log"
	"yune/cpp"

	fj "github.com/valyala/fastjson"
)

type Analyzer struct {
	Interpreter *cpp.Interpreter
	Errors      *Errors
	Defined     map[TopLevelDeclaration]struct{}
	Table       DeclarationTable
	State       *State
}

// Returns an analyzer with only the relevant data for a top-level analysis.
func (a Analyzer) TopLevel() Analyzer {
	return Analyzer{
		Interpreter: a.Interpreter,
		Errors:      a.Errors,
		Defined:     a.Defined,
		Table: DeclarationTable{
			topLevelDeclarations: a.Table.topLevelDeclarations,
		},
		State: a.State,
	}
}

func (a Analyzer) PushError(err error) {
	*a.Errors = append(*a.Errors, err)
	log.Panicf("Analyzer error: %s", err) // TODO: only exit when needed
}

func (a Analyzer) HasErrors() bool {
	return len(*a.Errors) > 0
}

// Evaluate a lowered Expression, assuming that Expression.Analyze has already been called on it.
func (a Analyzer) Evaluate(lowered cpp.Expression) (json *fj.Value) {
	// TEMP, speeds up compilation by a lot by preventing a C++ roundtrip
	// , but assumes these are not redefined as variables
	switch lowered {
	case "String":
		fj.MustParse(`{"StringType":{}}`)
	case "Int":
		fj.MustParse(`{"IntType":{}}`)
	case "Float":
		fj.MustParse(`{"FloatType":{}}`)
	case "Bool":
		fj.MustParse(`{"BoolType":{}}`)
	}
	getType := func(name string) cpp.Type {
		span := Span{} // TODO: span
		decl, ok := a.Table.Get(Name{String: name, Span: span})
		if !ok {
			a.PushError(MacroRequestedUndefinedVariable{
				// TODO
				Macro: Variable{Name: Name{Span: Span{}, String: "<unknown>"}},
				Name:  name,
			})
		}
		return decl.GetDeclaredType().LowerValue()
	}
	json, err := a.Interpreter.Evaluate(lowered, getType)
	if err != nil {
		panic("Failed to evaluate lowered expression. Error: " + err.Error())
	}
	return
}

func (a Analyzer) NewScope() Analyzer {
	a.Table = a.Table.NewScope()
	return a
}

func (a Analyzer) GetType(name Name) (TypeValue, Flags) {
	decl, ok := a.Table.Get(name)
	if !ok {
		a.PushError(UndefinedVariable{
			Span:   name.Span,
			String: name.String,
		})
	}
	topLevel, isTopLevel := decl.(TopLevelDeclaration)
	if isTopLevel {
		_, isDone := a.Defined[topLevel]
		if !isDone {
			topLevel.Analyze(a.TopLevel())
		}
	}
	// Non-top-level declarations are analyzed in sequential order,
	// so this type should already be available.
	_type := decl.GetDeclaredType()
	if _type == nil {
		panic("Declaration.GetDeclaredType() returned nil on local declaration '" + name.String + "'")
	}
	return _type, decl.GetFlags()
}

// NOTE: probably want top-level declarations to declare their prototypes as soon as those are known,
// and their full definitions after when they have been type checked

func (a Analyzer) Declare(decl TopLevelDeclaration) {
	err := a.Interpreter.Declare(decl.LowerDeclaration(a.State))
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
	err := a.Interpreter.Declare(decl.LowerDefinition(a.State))
	if err != nil {
		panic("Failed to define declaration " + decl.GetName().String)
	}
}
