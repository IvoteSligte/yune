package ast

import (
	"log"
	"yune/cpp"
)

type Module struct {
	Declarations []TopLevelDeclaration
}

func (m Module) Lower() (lowered cpp.Module, errors Errors) {
	declarations := map[string]TopLevelDeclaration{}

	// Register builtin declarations
	for _, decl := range BuiltinDeclarations {
		declarations[decl.GetName().String] = &decl
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
	for _, decl := range declarations {
		anal.Table.Add(decl)
	}
	for _, decl := range declarations {
		_, evaluated := anal.Defined[decl]
		if !evaluated {
			decl.Analyze(anal)
		}
	}
	if len(errors) > 0 {
		return
	}
	if len(anal.Defined) != len(declarations) {
		for _, decl := range declarations {
			_, defined := anal.Defined[decl]
			if !defined {
				// FIXME: detect recursive values and types
				log.Panicf("Declaration '%s' not defined even though evaluation has finished.", decl.GetName().String)
			}
		}
		log.Panicf(
			"The number of definitions (%d) does not match the number of declarations (%d), even though the evaluation process has finished.",
			len(anal.Defined),
			len(declarations),
		)
	}
	lowered = cpp.Repl.Declared // NOTE: this should probably reset the clang-repl process so multiple calls to Lower do not break things
	return
}
