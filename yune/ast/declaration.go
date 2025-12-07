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
	GetName() string
	GetDeclarationType() Type
}

var _ Declaration = FunctionDeclaration{}
var _ Declaration = FunctionParameter{}
var _ Declaration = ConstantDeclaration{}
var _ Declaration = VariableDeclaration{}

type TopLevelDeclarations = map[string]TopLevelDeclaration

func (m *Module) GetDeclarations() (TopLevelDeclarations, []DuplicateDeclaration) {
	declarations := make(TopLevelDeclarations, len(m.Declarations))
	errors := make([]DuplicateDeclaration, 0)

	for _, decl := range m.Declarations {
		name := decl.GetName()
		_, exists := declarations[name]

		if exists {
			errors = append(errors, DuplicateDeclaration{Name: name})
		} else {
			declarations[name] = decl
		}
	}
	return declarations, errors
}
