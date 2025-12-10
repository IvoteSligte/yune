package ast

type DeclarationMap = map[string]Declaration

type Env struct {
	parent       *Env
	declarations DeclarationMap
}

func (env *Env) Get(name string) Declaration {
	declaration, ok := env.declarations[name]
	if !ok && env.parent != nil {
		return env.parent.Get(name)
	}
	return declaration
}

type Declaration interface {
	Analyze() (queries Queries, finalizer Finalizer)
	GetSpan() Span
	GetName() string
	GetType() InferredType
}

var _ Declaration = &FunctionDeclaration{}
var _ Declaration = &FunctionParameter{}
var _ Declaration = &ConstantDeclaration{}
var _ Declaration = &VariableDeclaration{}

func (m *Module) GetDeclarationMap() (DeclarationMap, []error) {
	declarations := make(DeclarationMap, len(m.Declarations))
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
