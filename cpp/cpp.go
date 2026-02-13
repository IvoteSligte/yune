package cpp

import (
	"strings"
)

type Type = string

// A code block in the form of a Lambda function that is immediately invoked.
// This is a way to allow code blocks to be used where expressions can be used.
func LambdaBlock(b []Statement) string {
	return "[](){" + strings.Join(b, "") + "}()"
}

func String(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `std::string("` + s + `")`
}

type Expression = string

func Block(b []Statement) string {
	return "{\n" + strings.Join(b, "\n") + "\n}"
}

// Only the implementation part of a declaration.
// Useful for local declarations (i.e. declarations within functions).
type Definition = string

type Field struct {
	Name string
	Type Type
}

type Statement = string

type Module = string
