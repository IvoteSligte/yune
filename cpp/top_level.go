package cpp

import (
	"fmt"
	"strings"
	"yune/util"
)

type Module struct {
	Declarations []TopLevelDeclaration
}

// FIXME: declarations need to be ordered properly in the header file.
// It is primarily that type declarations need to come before constant declarations that use them.

func (m Module) GenHeader() string {
	// <tuple> for std::tuple, std::apply
	// <functional> for std::function
	// <string> for std::string
	// <fstream> for std::fstream (only for evaluation right now)
	prefix := `
#include <tuple>      // std::tuple, std::apply
#include <functional> // std::function
#include <string>     // std::string
#include <vector>     // std::vector
#include <fstream>    // std::fstream
#include <iostream>   // std::cout

// TODO: declare Type via ast/builtin.go?
struct Type {
    std::string id;
};

std::ostream& operator<<(std::ostream& out, const Type& t) {
    return out << t.id;
}
`
	return prefix + strings.Join(util.Map(m.Declarations, TopLevelDeclaration.GenHeader), "\n")
}

func (m Module) String() string {
	return util.Join(m.Declarations, "\n")
}

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
