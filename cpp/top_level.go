package cpp

import (
	"fmt"
	"yune/util"
)

type FunctionDeclaration struct {
	Name       string
	Parameters []FunctionParameter
	ReturnType Type
	Body       Block
}

// GenHeader implements TopLevelDeclaration.
func (f FunctionDeclaration) GenHeader() string {
	return fmt.Sprintf("%s %s(%s);", f.ReturnType, f.Name, util.Join(f.Parameters, ", "))
}

func (f FunctionDeclaration) String() string {
	return fmt.Sprintf("%s %s(%s) %s", f.ReturnType, f.Name, util.Join(f.Parameters, ", "), f.Body)
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

// GenHeader implements TopLevelDeclaration.
func (c ConstantDeclaration) GenHeader() string {
	return fmt.Sprintf("extern %s %s;", c.Type, c.Name)
}

func (c ConstantDeclaration) String() string {
	return fmt.Sprintf("%s %s = %s;", c.Type, c.Name, c.Value)
}

type StructDeclaration struct {
	Name   string
	Fields []Field
}

// GenHeader implements TopLevelDeclaration.
func (s StructDeclaration) GenHeader() string {
	fields := util.Join(s.Fields, "\n")
	return fmt.Sprintf("struct %s {\n%s\n};", s.Name, fields)
}

// String implements TopLevelDeclaration.
func (s StructDeclaration) String() string {
	return "" // already declared in header
}

type Field struct {
	Name string
	Type Type
}

func (f Field) String() string {
	return fmt.Sprintf("%s %s;", f.Type, f.Name)
}

// Alias of an existing type.
type TypeAlias struct {
	Alias string
	Of    string
}

// GenHeader implements TopLevelDeclaration.
func (t TypeAlias) GenHeader() string {
	return fmt.Sprintf("typedef %s %s;", t.Of, t.Alias)
}

func (t TypeAlias) Get() Type {
	return Type(t.Alias)
}

func (t TypeAlias) String() string {
	return "" // already declared in header
}

type TopLevelDeclaration interface {
	fmt.Stringer
	GenHeader() string
}

var _ TopLevelDeclaration = FunctionDeclaration{}
var _ TopLevelDeclaration = ConstantDeclaration{}
var _ TopLevelDeclaration = StructDeclaration{}
var _ TopLevelDeclaration = TypeAlias{}
