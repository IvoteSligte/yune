package value

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"yune/util"
)

var LispTagMap map[string]reflect.Type = map[string]reflect.Type{}

type Lisper interface {
	LispTag() string
}

func Serialize(v any) (serialized string) {
	switch t := v.(type) {
	case int, float64, bool, string:
		bytes, _ := json.Marshal(t) // handles character escapes and such
		return string(bytes)
	}
	r := reflect.ValueOf(v)
	t := r.Type()
	tag := t.String()
	if l, ok := v.(Lisper); ok {
		tag = l.LispTag()
	}
	LispTagMap[tag] = t
	serialized = "(" + tag
	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < r.NumField(); i++ {
			field := Serialize(r.Field(i))
			serialized += " " + field
		}
	default:
		panic("Tried to serialize unsupported type kind: " + t.Kind().String())
	}
	serialized += ")"
	return
}

func deserializeFirst(s string) (any, string) {
	if s[0] == ')' {
		return nil, ""
	}
	if s[0] == '(' {
		s = s[1 : len(s)-1] // strip ()
		tagEnd := strings.IndexRune(s, ' ')
		tag := s[:tagEnd]
		s = s[tagEnd+1:]
		sublisps := []any{}
		for {
			var sublisp any
			sublisp, s = deserializeFirst(s)
			if sublisp == nil {
				break
			}
			sublisps = append(sublisps, sublisp)
		}
		t := LispTagMap[tag]
		return delispers[tag](sublisps), ""
	}
	literalEnd := strings.IndexRune(s, ' ')
	literal := s[:literalEnd]
	s = s[literalEnd+1:]
	var v any
	err := json.Unmarshal([]byte(literal), &v)
	if err != nil {
		panic("JSON error deserializing Lisp literal: " + err.Error())
	}
	return v, s
}

func Deserialize(s string) any {
	v, s := deserializeFirst(s)
	if s != "" {
		panic("More than one Lisp expression in string. Remaining: " + s)
	}
	return v
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
