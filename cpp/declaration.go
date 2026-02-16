package cpp

import (
	"fmt"
	"strings"
)

type Body = string

type Declaration struct {
	Header         string
	Implementation string
}

func NewDeclaration(header string, implementation string) Declaration {
	if len(header) > 0 && header[0] == '\n' {
		header = header[1:]
	}
	if len(implementation) > 0 && implementation[0] == '\n' {
		implementation = implementation[1:]
	}
	return Declaration{Header: header, Implementation: implementation}
}

func FunctionDeclaration(id string, name string, parameters []FunctionParameter, returnType Type, body Body) Declaration {
	params := strings.Join(parameters, ", ")
	return NewDeclaration(
		/* Header */ fmt.Sprintf(`
struct %s_ {
    %s operator()(%s) const;
    std::string serialize() const;
} %s;`, id, returnType, params, name),
		/* Implementation */ fmt.Sprintf(`
%s %s_::operator()(%s) const %s
std::string %s_::serialize() const {
    return R"({ "FnId": "%s" })";
}`, returnType, id, params, body, id, id),
	)
}

type FunctionParameter = string

func ConstantDeclaration(name string, _type Type, value Expression) Declaration {
	return NewDeclaration(
		/* Header: */ fmt.Sprintf("extern %s %s;", _type, name),
		/* Implementation:*/ fmt.Sprintf("%s %s = %s;", _type, name, value),
	)
}

func StructDeclaration(name string, fields []string) Declaration {
	return NewDeclaration(
		/* Header: */ fmt.Sprintf("struct %s %s;", name, Block(fields)),
		/* Implementation: */ "", // already declared in the header
	)
}

func NewField(name string, _type Type) string {
	return fmt.Sprintf("%s %s;", _type, name)
}
