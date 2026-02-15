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
		Errors:  &errors,
		Defined: map[TopLevelDeclaration]struct{}{},
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
		// FIXME: if the main function is not evaluated last then clang-repl evaluation breaks
		_, evaluated := anal.Defined[decl]
		if !evaluated {
			println("analyzing top-level", decl.GetName().String)
			decl.Analyze(anal)
		}
	}
	if len(errors) > 0 {
		return
	}
	if len(anal.Defined) != len(declarations) {
		panic("The number of definitions does not match the number of declarations, even though the evaluation process has finished.")
	}
	lowered = cpp.Repl.GetDeclared() // NOTE: this should probably reset the clang-repl process so multiple calls to Lower do not break things
	return
}
