package ast

import (
	"fmt"
	"yune/cpp"
	"yune/util"
)

type TopLevelDeclaration interface {
	Declaration

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
	var argument TypeValue = &TupleType{Elements: paramTypes}
	if len(parameters) == 1 {
		argument = paramTypes[0]
	}
	return &FnType{Argument: argument, Return: returnType.Get()}
}

type FunctionDeclaration struct {
	Span       Span
	Name       Name
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
}

func (d *FunctionDeclaration) GetSpan() Span {
	return d.Name.GetSpan()
}

// TypeCheckBody implements Declaration.
func (d *FunctionDeclaration) Analyze(anal Analyzer) {
	if d.ReturnType.Get() != nil {
		return // already (being) analyzed
	}
	anal = anal.NewScope()
	if err := anal.Table.Add(d); err != nil {
		panic("Duplicate declaration error in new scope: " + err.Error())
	}
	analyzeFunctionHeader(anal, d.Parameters, &d.ReturnType)
	anal.Declare(d)
	analyzeFunctionBody(anal, d.ReturnType.Get(), d.Body)
	declaredType := d.GetDeclaredType()
	if d.GetName().String == "main" && !declaredType.Eq(MainType) {
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
} %s;`, d.Name.String, d.ReturnType.Lower(), params, d.Name.Lower())
}

// LowerDefinition implements TopLevelDeclaration.
func (d *FunctionDeclaration) LowerDefinition() cpp.Definition {
	params := util.JoinFunction(d.Parameters, ", ", FunctionParameter.Lower)
	return fmt.Sprintf(`%s %s_::operator()(%s) const %s
std::string %s_::serialize() const {
    return R"({ "Function": "%s" })";
}`, d.ReturnType.Lower(), d.Name.String, params, cpp.Block(d.Body.Lower()), d.Name.String, d.Name.String)
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
	Span        Span
	Name        Name
	Type        Type
	Body        Block
	HasCaptures bool
	IsBuiltin   bool
}

// GetSpan implements TopLevelDeclaration.
func (d *ConstantDeclaration) GetSpan() Span {
	return d.Name.GetSpan()
}

// Analyze implements TopLevelDeclaration.
func (d *ConstantDeclaration) Analyze(anal Analyzer) {
	// The definition of the builtin "Type" is recursive, which is normally not allowed.
	if d.Name.String == "Type" {
		anal.Declare(d)
		anal.Define(d)
		return
	}
	if d.Type.Get() != nil {
		return // already (being) analyzed
	}
	declaredType := d.Type.Analyze(anal)
	scope := anal.NewScope()
	bodyType := d.Body.Analyze(declaredType, scope)

	if !declaredType.Eq(bodyType) {
		anal.PushError(ConstantTypeMismatch{
			Expected: declaredType,
			Found:    bodyType,
			At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
		})
	}
	d.HasCaptures = len(*scope.Table.captures) > 0
	anal.Define(d)
}

// LowerDeclaration implements TopLevelDeclaration.
func (d ConstantDeclaration) LowerDeclaration() cpp.Declaration {
	return fmt.Sprintf("extern %s %s;", d.Type.Lower(), d.Name.Lower())
}

// LowerDefinition implements TopLevelDeclaration.
func (d ConstantDeclaration) LowerDefinition() cpp.Definition {
	return fmt.Sprintf("%s %s = %s;", d.Type.Lower(), d.Name.Lower(), cpp.LambdaBlock(d.Body.Lower(), d.HasCaptures))
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

// TODO: when types and type aliases can be created, make sure that
// values are cached and aliases are properly resolved.

var _ TopLevelDeclaration = &FunctionDeclaration{}
var _ TopLevelDeclaration = &ConstantDeclaration{}
