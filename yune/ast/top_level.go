package ast

import (
	"yune/cpp"
	"yune/util"
)

func newDeclarationMap[T Declaration](declarations []T) (declarationMap DeclarationMap, errors Errors) {
	declarationMap = make(DeclarationMap, len(declarations))
	for _, decl := range declarations {
		previous, exists := declarationMap[decl.GetName()]

		if exists {
			errors = append(errors, DuplicateDeclaration{
				First:  previous,
				Second: decl,
			})
		} else {
			declarationMap[decl.GetName()] = decl
		}
	}
	return
}

type Module struct {
	Declarations []TopLevelDeclaration
}

func (m Module) Lower() cpp.Module {
	return cpp.Module{
		Declarations: util.Map(m.Declarations, TopLevelDeclaration.Lower),
	}
}

func (m *Module) Analyze() (queries Queries, finalizer Finalizer) {
	finalizers := make([]Finalizer, len(m.Declarations))

	for _, decl := range m.Declarations {
		_queries, _finalizer := decl.Analyze()
		queries = append(queries, _queries...)
		finalizers = append(finalizers, _finalizer)
	}
	finalizer = func(env Env) (errors Errors) {
		declarations, errors := newDeclarationMap(m.Declarations)
		env = Env{
			parent:       &env,
			declarations: declarations,
		}
		if len(errors) > 0 {
			return
		}
		for _, fin := range finalizers {
			errors = append(errors, fin(env)...)
		}
		return
	}
	return
}

type FunctionDeclaration struct {
	Span
	Name       Name
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
}

// Lower implements TopLevelDeclaration.
func (d FunctionDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.FunctionDeclaration{
		Name:       d.Name.String,
		Parameters: util.Map(d.Parameters, FunctionParameter.Lower),
		ReturnType: d.ReturnType.Lower(),
		Body:       d.Body.Lower(),
	}
}

// Analyze implements Declaration.
func (d FunctionDeclaration) Analyze() (queries Queries, finalizer Finalizer) {
	parameterFins := make([]Finalizer, len(d.Parameters)+2)
	for i := range d.Parameters {
		_queries, _finalizer := d.Parameters[i].Analyze()
		queries = append(queries, _queries...)
		parameterFins = append(parameterFins, _finalizer)
	}
	_queries, returnTypeFin := d.ReturnType.Analyze()
	queries = append(queries, _queries...)
	_queries, bodyFin := d.Body.Analyze()
	queries = append(queries, _queries...)
	finalizer = func(env Env) (errors Errors) {
		errors = returnTypeFin(env)
		declarations, _errors := newDeclarationMap(d.Parameters)
		errors = append(errors, _errors...)
		env = Env{
			parent:       &env,
			declarations: declarations,
		}
		if len(errors) > 0 {
			return
		}
		errors = bodyFin(env)
		if len(errors) > 0 {
			return
		}
		returnType := d.ReturnType.GetType()
		bodyType := d.Body.GetType()
		if !returnType.Eq(bodyType) {
			errors = append(errors, TypeMismatch{
				Expected: returnType,
				Found:    bodyType,
				At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
			})
		}
		return
	}
	return
}

func (FunctionDeclaration) topLevelDeclaration() {}

func (d FunctionDeclaration) GetName() string {
	return d.Name.String
}

func (d FunctionDeclaration) GetType() InferredType {
	return InferredType{
		name:     "Fn",
		generics: append(util.Map(d.Parameters, FunctionParameter.GetType), d.ReturnType.InferredType),
	}
}

type FunctionParameter struct {
	Span
	Name Name
	Type Type
}

func (d FunctionParameter) Lower() cpp.FunctionParameter {
	return cpp.FunctionParameter{
		Name: d.Name.String,
		Type: d.Type.Lower(),
	}
}

// Analyze implements Declaration.
func (d FunctionParameter) Analyze() (queries Queries, finalizer Finalizer) {
	queries, finalizer = d.Type.Analyze()
	return
}

func (d FunctionParameter) GetName() string {
	return d.Name.String
}

func (d FunctionParameter) GetType() InferredType {
	return d.Type.InferredType
}

type ConstantDeclaration struct {
	Span
	Name Name
	Type Type
	Body Block
}

// Lower implements TopLevelDeclaration.
func (d ConstantDeclaration) Lower() cpp.TopLevelDeclaration {
	return cpp.ConstantDeclaration{
		Name:  d.Name.String,
		Type:  d.Type.Lower(),
		Value: d.Body.Lower(),
	}
}

// Analyze implements Declaration.
func (d ConstantDeclaration) Analyze() (queries Queries, finalizer Finalizer) {
	queries, typeFin := d.Type.Analyze()
	_queries, bodyFin := d.Body.Analyze()
	queries = append(queries, _queries...)
	finalizer = func(env Env) (errors Errors) {
		errors = append(typeFin(env), bodyFin(env)...)
		return
	}
	return
}

// GetType implements Declaration.
func (d ConstantDeclaration) GetType() InferredType {
	return d.Type.InferredType
}

func (ConstantDeclaration) topLevelDeclaration() {}

func (d ConstantDeclaration) GetName() string {
	return d.Name.String
}

func (d ConstantDeclaration) GetDeclarationType() Type {
	return d.Type
}

type TopLevelDeclaration interface {
	Declaration
	topLevelDeclaration()
	Lower() cpp.TopLevelDeclaration
}

var _ TopLevelDeclaration = FunctionDeclaration{}
var _ TopLevelDeclaration = ConstantDeclaration{}
