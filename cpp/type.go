package cpp

import (
	"fmt"
	"log"
	"slices"
	"yune/util"
)

type Type interface {
	fmt.Stringer
	Eq(other Type) bool
}

type NamedType struct {
	Name string
}

// Eq implements Type.
func (t NamedType) Eq(other Type) bool {
	{
		other, ok := other.(NamedType)
		return ok &&
			t.Name == other.Name
	}
}

// String implements Type.
func (t NamedType) String() string {
	if len(t.Name) == 0 {
		log.Println("WARN: Found empty type name when converting to string.")
	}
	return t.Name
}

// NOTE: requires #include <functional>
type FunctionType struct {
	Parameter  Type
	ReturnType Type
}

// String implements Type.
func (t FunctionType) String() string {
	_, isTupleType := t.Parameter.(TupleType)
	if isTupleType {
		// simplifies the type of functions that only take a tuple (most Yune functions)
		// and prevents a few issues, such as std::apply not working with zero-element tuples
		parameters := util.Join(t.Parameter.(TupleType).Elements, ", ")
		return fmt.Sprintf("std::function<%s(%s)>", t.ReturnType, parameters)
	} else {
		return fmt.Sprintf("std::function<%s(%s)>", t.ReturnType, t.Parameter)
	}
}

// Eq implements Type.
func (t FunctionType) Eq(other Type) bool {
	{
		other, ok := other.(FunctionType)
		return ok &&
			t.Parameter.Eq(other.Parameter) &&
			t.ReturnType.Eq(other.ReturnType)
	}
}

// NOTE: requires #include <tuple>
type TupleType struct {
	Elements []Type
}

// Eq implements Type.
func (t TupleType) Eq(other Type) bool {
	{
		other, ok := other.(TupleType)
		return ok &&
			slices.EqualFunc(t.Elements, other.Elements, Type.Eq)
	}
}

// String implements Type.
func (t TupleType) String() string {
	return fmt.Sprintf("std::tuple<%s>", util.Join(t.Elements, ", "))
}

var _ Type = NamedType{}
var _ Type = FunctionType{}
var _ Type = TupleType{}
