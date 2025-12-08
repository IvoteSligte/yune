package ast

type Environment struct {
	parent  *Environment
	symbols map[string]Declaration
}

func (env *Environment) Get(name string) (Declaration, bool) {
	declaration, ok := env.symbols[name]
	if !ok && env.parent != nil {
		return env.parent.Get(name)
	}
	return declaration, ok
}

type Declaration interface {
	GetSpan() Span
	GetName() string
	GetDeclarationType() Type
}

var _ Declaration = FunctionDeclaration{}
var _ Declaration = FunctionParameter{}
var _ Declaration = ConstantDeclaration{}
var _ Declaration = VariableDeclaration{}

type TopLevelDeclarations = map[string]TopLevelDeclaration

func (m *Module) GetDeclarations() (TopLevelDeclarations, []error) {
	declarations := make(TopLevelDeclarations, len(m.Declarations))
	errors := make([]error, 0)

	for _, decl := range m.Declarations {
		name := decl.GetName()
		first_decl, exists := declarations[name]

		if exists {
			errors = append(errors, DuplicateDeclaration{
				First:  first_decl,
				Second: decl,
			})
		} else {
			declarations[name] = decl
		}
	}
	return declarations, errors
}
