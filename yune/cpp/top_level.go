package cpp

import (
	"fmt"
	"yune/util"
)

type Module struct {
	Declarations []Declaration
}

func (m Module) String() string {
	return util.SeparatedBy(m.Declarations, "\n\n")
}

type FunctionDeclaration struct {
	Name       string
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
}

func (f FunctionDeclaration) String() string {
	return fmt.Sprintf("%s %s(%s) %s", f.ReturnType, f.Name, util.SeparatedBy(f.Parameters, ", "), f.Body)
}

type FunctionParameter struct {
	Name string
	Type Type
}

func (p FunctionParameter) String() string {
	return fmt.Sprintf("%s %s", p.Type, p.Name)
}

type ConstantDeclaration struct {
	Name  string
	Type  Type
	Value Expression
}

func (c ConstantDeclaration) String() string {
	return fmt.Sprintf("%s %s = %s;", c.Type, c.Name, c.Value)
}

type BuiltinDeclaration struct {
	Raw string
}

func (b BuiltinDeclaration) String() string {
	return b.Raw
}

var BuiltinDeclarations []Declaration = []Declaration{
	BuiltinDeclaration{Raw: "typedef Int int;"},
	BuiltinDeclaration{Raw: "typedef Float float;"},
	BuiltinDeclaration{Raw: "typedef Bool bool;"},
	BuiltinDeclaration{Raw: "typedef Nil void;"},
}

type TopLevelDeclaration fmt.Stringer
