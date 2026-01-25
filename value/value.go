package value

import (
	"fmt"
	"regexp"
	"strings"
	"yune/util"
)

type Value string

var functionTypeRegexp = regexp.MustCompile(`^std::function<(.*?)\((.*?)\)>$`)
var tupleTypeRegexp = regexp.MustCompile(`^std::tuple<(.*?)>$`)

type Type string

func (t Type) Eq(other Type) bool {
	return string(t) == string(other)
}

func (t Type) IsSubType(other Type) bool {
	if t.Eq(other) {
		return true
	}
	if string(other) == "Type" {
		return true // t is also a type
	}
	tElements, tIsTuple := t.ToTuple()
	otherElements, otherIsTuple := other.ToTuple()
	if tIsTuple && otherIsTuple {
		if len(tElements) != len(otherElements) {
			return false
		}
		for i, tElem := range tElements {
			if !tElem.IsSubType(otherElements[i]) {
				return false
			}
		}
		return true
	}
	return false
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

func (t Type) ToFunction() (argumentType Type, returnType Type, isFunction bool) {
	matches := functionTypeRegexp.FindStringSubmatch(string(t))
	if len(matches) == 0 {
		return // not a function
	}
	isFunction = true
	_ = matches[0] // full string
	returnType = Type(matches[1])
	argumentType = Type(fmt.Sprintf("std::tuple<%s>", matches[2]))
	return
}

func (t Type) ToTuple() (elements []Type, isTuple bool) {
	matches := tupleTypeRegexp.FindStringSubmatch(string(t))
	if len(matches) == 0 {
		return // not a tuple
	}
	isTuple = true
	_ = matches[0] // full string
	if len(matches[1]) > 0 {
		for elem := range strings.SplitSeq(matches[1], ", ") {
			elements = append(elements, Type(elem))
		}
	}
	return
}

func (t Type) IsTuple() bool {
	_, isTuple := t.ToTuple()
	return isTuple
}

func (t Type) IsEmptyTuple() bool {
	return string(t) == "std::tuple<>"
}

func (t Type) IsTypeType() bool {
	if string(t) == "Type" {
		return true
	}
	// Unlike for other values, the type of a tuple of elements never becomes "Type" itself.
	// (0, "string") -> (Int, String) -> (Type, Type) -> (Type, Type) -> ...
	elements, isTuple := t.ToTuple()
	if isTuple && util.All(Type.IsTypeType, elements...) {
		return true
	}
	return false
}
