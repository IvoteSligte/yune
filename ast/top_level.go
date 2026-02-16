package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"
)

type GetId interface {
	GetId() string
}

// Stores declarations that need to be serializable from C++, such as functions.
var registeredNodes = map[string]GetId{}

func registerNode(node GetId) string {
	id := node.GetId()
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

// GetId implements TopLevelDeclaration.
func (d *FunctionDeclaration) GetId() string {
	return d.Name.String
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

// LowerDeclaration implements TopLevelDeclaration.
func (d *FunctionDeclaration) LowerDeclaration() cpp.Declaration {
	params := util.JoinFunction(d.Parameters, ", ", FunctionParameter.Lower)
	return fmt.Sprintf(`struct %s_ {
    %s operator()(%s) const;
    std::string serialize() const;
} %s;`, d.GetId(), d.ReturnType.Lower(), params, d.Name.Lower())
}

// LowerDefinition implements TopLevelDeclaration.
func (d *FunctionDeclaration) LowerDefinition() cpp.Definition {
	params := util.JoinFunction(d.Parameters, ", ", FunctionParameter.Lower)
	id := d.GetId()
	return fmt.Sprintf(`%s %s_::operator()(%s) const %s
std::string %s_::serialize() const {
    return R"({ "FnId": "%s" })";
}`, d.ReturnType.Lower(), id, params, cpp.Block(d.Body.Lower()), id, id)
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
	return d.Type.Lower() + " " + d.Name.Lower()
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

// GetId implements TopLevelDeclaration.
func (d *ConstantDeclaration) GetId() string {
	return d.Name.String
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

// LowerDeclaration implements TopLevelDeclaration.
func (d ConstantDeclaration) LowerDeclaration() cpp.Declaration {
	return fmt.Sprintf("extern %s %s;", d.Type.Lower(), d.Name.Lower())
}

// LowerDefinition implements TopLevelDeclaration.
func (d ConstantDeclaration) LowerDefinition() cpp.Definition {
	return fmt.Sprintf("%s %s = %s;", d.Type.Lower(), d.Name.Lower(), cpp.LambdaBlock(d.Body.Lower()))
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
	GetId

	// Lowers the declaration to a C++ forward declaration.
	// Assumes type checking has been performed.
	LowerDeclaration() cpp.Declaration

	// Lowers the declaration to executable C++ code.
	// Assumes LowerDeclaration has been called.
	//
	// NOTE: when the value has been computed, this function should
	// lower to a more efficient representation instead of forcing
	// the same code to run.
	LowerDefinition() cpp.Definition

	Analyze(anal Analyzer)
}

// TODO: when types and type aliases can be created, make sure that
// values are cached and aliases are properly resolved.

var _ TopLevelDeclaration = &FunctionDeclaration{}
var _ TopLevelDeclaration = &ConstantDeclaration{}
