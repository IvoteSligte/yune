package ast

import (
	"yune/cpp"
)

type Module struct {
	Declarations []TopLevelDeclaration
}

func (m Module) Lower() (lowered cpp.Module, errors Errors) {
	declarations := map[string]TopLevelDeclaration{}

	// Register builtin declarations
	for _, decl := range BuiltinDeclarations {
		declarations[decl.GetName().String] = decl
	}
	// get unique mapping of name -> declaration
	for _, decl := range m.Declarations {
		name := decl.GetName()
		other, exists := declarations[name.String]

		if exists {
			errors = append(errors, DuplicateDeclaration{First: other, Second: decl})
		} else {
			declarations[name.String] = decl
		}
	}
	if len(errors) > 0 {
		return
	}
	anal := Analyzer{
		Errors:                &errors,
		EvaluatedDeclarations: map[string]TopLevelDeclaration{},
		Table: DeclarationTable{
			topLevelDeclarations: declarations,
		},
	}
	for _, decl := range BuiltinDeclarations {
		anal.Table.Add(decl)
	}
	for _, decl := range m.Declarations {
		anal.Table.Add(decl)
	}
	for _, decl := range m.Declarations {
		// FIXME: if the main function is not evaluated last then Cling evaluation breaks
		_, evaluated := anal.EvaluatedDeclarations[decl.GetName().String]
		if !evaluated {
			decl.Analyze(anal)
		}
	}
	if len(errors) > 0 {
		return
	}
	if len(anal.EvaluatedDeclarations) != len(anal.Declarations) {
		panic("The number of evaluated declarations does not match the total number of declarations, even though the evaluation process has finished.")
	}
	lowered = cpp.Cling.GetDeclared() // NOTE: this should probably reset the Cling process so multiple calls to Lower do not break things
	return
}
