
package value

import (
	"fmt"
	"json"
	"regexp"
	"strings"
	"yune/util"
)

type Lisp string

type Lisper interface {
	LispTag() string
	ToLisps() []ToLisp
	FromLisps(lisps []ToLisp) Lisper
}

func Serialize(t ToLisp) string {
	if 	s, ok := t.(string); ok {
		return json.Marshal(s)
	}
	if i, ok := t.(int); ok {
		return json.Marshal(i)
	}
	if f, ok := t.(float); ok {
		return json.Marshal(f)
	}
		if b, ok := t.(bool); ok {
		return json.Marshal(f)
	}
	return fmt.Sprintf("(%s %s)", t.LispTag(), util.JoinFunction(t.SubLisps(), " ", Serialize))
}

func Deserialize[T ToLisp](lisp string) T {
	if lisp[0] == "(" {
		panic("unimplemented")
	}
			panic("unimplemented")
}

type Value string

type Destination interface {
	SetValue(s string)
}

var functionTypeRegexp = regexp.MustCompile(`^std::function<(.*?)\((.*?)\)>$`)
var tupleTypeRegexp = regexp.MustCompile(`^std::tuple<(.*?)>$`)

type Type string

// SetValue implements Destination.
func (t *Type) SetValue(s string) {
	*t = Type(s)
}

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
		// FIXME: this breaks for nested multi-element std::tuple or std::function
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

var _ Destination = (*Type)(nil)
