package ast

import (
	"math/rand/v2"
	"yune/cpp"
	"yune/util"
)

type Id = uint64

// Stores declarations that need to be serializable from C++, such as functions.
var registeredNodes = map[Id]Node{}

// TODO: free stored declarations use human-readable strings as Id,
// and only add nodes to the map once (instead of every call to Lower)
func registerNode(node Node) Id {
	id := rand.Uint64()
	registeredNodes[id] = node
	return id
}

// Assumes that the analyzer is in the function's body scope.
func analyzeFunctionHeader(anal Analyzer, parameters []FunctionParameter, returnType *Type) {
	// check for duplicate parameters
	for i := range parameters {
		param := &parameters[i]
		if err := anal.Table.Add(param); err != nil {
			anal.PushError(err)
		}
		param.Analyze(anal)
	}
	returnType.Analyze(anal)
}

// Assumes that the analyzer is in the function's body scope and parameters have been declared.
func analyzeFunctionBody(anal Analyzer, returnType TypeValue, body Block) {
	bodyType := body.Analyze(returnType, anal)
	if !returnType.Eq(bodyType) {
		anal.PushError(ReturnTypeMismatch{
			Expected: returnType,
			Found:    bodyType,
			At:       body.Statements[len(body.Statements)-1].GetSpan(),
		})
	}
}

func getFunctionType(parameters []FunctionParameter, returnType Type) TypeValue {
	if returnType.Get() == nil {
		return nil
	}
	paramTypes := []TypeValue{}
	for _, param := range parameters {
		paramType := param.GetDeclaredType()
		if paramType == nil {
			return nil
		}
		paramTypes = append(paramTypes, paramType)
	}
	var argument TypeValue = NewTupleType(paramTypes...)
	if len(parameters) == 1 {
		argument = paramTypes[0]
	}
	return FnType{Argument: argument, Return: returnType.Get()}
}

type FunctionDeclaration struct {
	Span       Span
	Name       Name
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
}

// GetSpan implements TopLevelDeclaration.
func (d *FunctionDeclaration) GetSpan() Span {
	return d.Name.GetSpan()
}

// TypeCheckBody implements Declaration.
func (d *FunctionDeclaration) Analyze(anal Analyzer) {
	if d.ReturnType.Get() != nil {
		return // already (being) analyzed
	}
	anal = anal.NewScope(nil)

	if err := anal.Table.Add(d); err != nil {
		panic("Duplicate declaration error in new scope: " + err.Error())
	}
	analyzeFunctionHeader(anal, d.Parameters, &d.ReturnType)
	anal.Declare(d)
	analyzeFunctionBody(anal, d.ReturnType.Get(), d.Body)
	declaredType := d.GetDeclaredType()
	if d.GetName().String == "main" && declaredType != nil && !declaredType.Eq(MainType) {
		anal.PushError(InvalidMainSignature{
			Found: d.GetDeclaredType(),
			At:    d.Name.GetSpan(),
		})
	}
	anal.Define(d)
}

// Lower implements Declaration.
func (d *FunctionDeclaration) Lower() cpp.Declaration {
	return cpp.FunctionDeclaration(
		registerNode(d),
		d.Name.Lower(),
		util.Map(d.Parameters, FunctionParameter.Lower),
		d.ReturnType.Lower(),
		cpp.Block(d.Body.Lower()),
	)
}

func (d FunctionDeclaration) GetName() Name {
	return d.Name
}

func (d FunctionDeclaration) GetDeclaredType() TypeValue {
	return getFunctionType(d.Parameters, d.ReturnType)
}

type FunctionParameter struct {
	Span
	Name Name
	Type Type
}

func (d *FunctionParameter) Analyze(anal Analyzer) {
	if d.Type.Get() != nil {
		panic("Re-analyzing function parameter '" + d.Name.String + "'")
	}
	d.Type.Analyze(anal)
}

func (d FunctionParameter) Lower() cpp.FunctionParameter {
	return d.Type.Lower() + " " + d.Name.String
}

// GetName implements Declaration
func (d FunctionParameter) GetName() Name {
	return d.Name
}

// GetDeclaredType implements Declaration
func (d FunctionParameter) GetDeclaredType() TypeValue {
	return d.Type.Get()
}

type ConstantDeclaration struct {
	Span Span
	Name Name
	Type Type
	Body Block
}

// GetSpan implements TopLevelDeclaration.
func (d *ConstantDeclaration) GetSpan() Span {
	return d.Name.GetSpan()
}

// Analyze implements TopLevelDeclaration.
func (d *ConstantDeclaration) Analyze(anal Analyzer) {
	if d.Type.Get() != nil {
		return // already (being) analyzed
	}
	declaredType := d.Type.Analyze(anal)
	bodyType := d.Body.Analyze(declaredType, anal.NewScope(nil))

	if !declaredType.Eq(bodyType) {
		anal.PushError(ConstantTypeMismatch{
			Expected: declaredType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
		})
	}
	anal.Define(d)
}

// Lower implements Declaration.
func (d ConstantDeclaration) Lower() cpp.Declaration {
	return cpp.ConstantDeclaration(
		d.Name.Lower(),
		d.Type.Lower(),
		cpp.LambdaBlock(d.Body.Lower()),
	)
}

// GetType implements Declaration.
func (d ConstantDeclaration) GetDeclaredType() TypeValue {
	return d.Type.Get()
}

func (d ConstantDeclaration) GetName() Name {
	return d.Name
}

func (d ConstantDeclaration) GetDeclarationType() Type {
	return d.Type
}

type TopLevelDeclaration interface {
	Declaration
	// Lowers the declaration to executable C++ code.
	// Assumes type checking has been performed.
	//
	// NOTE: when the value has been computed, this function should
	// lower to a more efficient representation instead of forcing
	// the same code to run.
	Lower() cpp.Declaration // NOTE: should this be split into separate functions for lowering to declaration and definition

	Analyze(anal Analyzer)
}

// TODO: when types and type aliases can be created, make sure that
// values are cached and aliases are properly resolved.

var _ TopLevelDeclaration = &FunctionDeclaration{}
var _ TopLevelDeclaration = &ConstantDeclaration{}
