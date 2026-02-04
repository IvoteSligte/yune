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

func FunctionDeclaration(name string, parameters []FunctionParameter, returnType Type, body Body) Declaration {
	prefix := fmt.Sprintf(`%s %s(%s)`, returnType, name, strings.Join(parameters, ", "))
	return Declaration{
		Header:         prefix + ";",
		Implementation: prefix + body,
	}
}

type FunctionParameter = string

func ConstantDeclaration(name string, _type Type, value Expression) Declaration {
	return Declaration{
		Header:         fmt.Sprintf("extern %s %s;", _type, name),
		Implementation: fmt.Sprintf("%s %s = %s;", _type, name, value),
	}
}

func StructDeclaration(name string, fields []string) Declaration {
	return Declaration{
		Header:         fmt.Sprintf("struct %s %s;", name, Block(fields)),
		Implementation: "", // already declared in the header
	}
}

func Field(name string, _type Type) string {
	return fmt.Sprintf("%s %s;", _type, name)
}
