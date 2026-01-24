package value

import (
	"fmt"
	"regexp"
	"strings"
	"yune/util"
)

type Value string

type Type string

func (t Type) Eq(other Type) bool {
	return string(t) == string(other)
}

func (t Type) String() string {
	return string(t)
}

func NewTupleType(elements []Type) Type {
	return Type(fmt.Sprintf("std::tuple<%s>", util.Join(elements, ", ")))
}

func NewFunctionType(arguments []Type, returnType Type) Type {
	return Type(fmt.Sprintf("std::function<%s(%s)>", returnType, util.Join(arguments, ", ")))
}

func (t Type) IsFunction() bool {
	return strings.HasPrefix(string(t), "std::function")
}

func (t Type) IsTuple() bool {
	return strings.HasPrefix(string(t), "std::tuple")
}

func (t Type) IsEmptyTuple() bool {
	return string(t) == "std::tuple<>"
}

type functionType = struct {
	argumentType Type
	returnType   Type
}

// Returns (argument type, return type)
func (t Type) ToFunction() (Type, Type) {
	if !t.IsFunction() {
		panic("Called ToFunction() on non-function type.")
	}
	splitRegexp := regexp.MustCompile(`std::function<(.*?)\((.*?)\)>`)
	matches := splitRegexp.FindStringSubmatch(string(t))
	_ = matches[0] // full string
	returnType := matches[1]
	argumentType := matches[2]
	return Type(argumentType), Type(returnType)
}
