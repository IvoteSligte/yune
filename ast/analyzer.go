package ast

import "yune/cpp"

type Analyzer struct {
	Errors                *Errors
	Declarations          map[string]TopLevelDeclaration
	EvaluatedDeclarations map[string]TopLevelDeclaration
	Table                 DeclarationTable
}

func (a Analyzer) PushError(err error) {
	*a.Errors = append(*a.Errors, err)
}

func (a Analyzer) AppendErrors(errors ...error) {
	*a.Errors = append(*a.Errors, errors...)
}

func (a Analyzer) HasErrors() bool {
	return len(*a.Errors) > 0
}

func SubAnalyze[T any](a Analyzer, f func(Analyzer) T) T {
	sub := Analyzer{
		Errors: a.Errors,
		// Evaluation cannot depend on variable declarations in the outer scope
		// as the values of these are not known.
		Table: a.Table.TopLevel(),
	}
	return f(sub)
}

// Evaluate an Expression, assuming that Expression.Analyze has already been called on it.
func (a Analyzer) Evaluate(expr Expression) (json string) {
	definitions := []cpp.Definition{}
	lowered := expr.Lower(&definitions)
	// braces group the definitions and expression together into a single "transaction"
	json, err := cpp.Cling.Evaluate("{\n" + defString(definitions) + lowered + "}\n")
	if err != nil {
		panic("Failed to evaluate lowered expression. Error: " + err.Error())
	}
	return
}

func (a Analyzer) NewScope() Analyzer {
	a.Table = a.Table.NewScope()
	return a
}

func (a Analyzer) GetType(name Name) TypeValue {
	decl, ok := a.Table.Get(name.String)
	if !ok {
		panic("Unknown declaration: " + name.String)
	}
	return decl.GetDeclaredType()
}
