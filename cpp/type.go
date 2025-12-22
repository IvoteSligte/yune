package cpp

import (
	"fmt"
	"log"
	"slices"
	"yune/util"
)

type Type struct {
	Name     string
	Generics []Type
}

func (t Type) IsUninit() bool {
	return t.Eq(Type{})
}

func (t Type) String() string {
	if len(t.Name) == 0 {
		log.Println("WARN: Found empty type name when converting to string.")
	}
	if len(t.Generics) == 0 {
		return t.Name
	} else {
		return fmt.Sprintf("%s<%s>", t.Name, util.SeparatedBy(t.Generics, ", "))
	}
}

func (left Type) Eq(right Type) bool {
	return left.Name == right.Name &&
		slices.EqualFunc(left.Generics, right.Generics, Type.Eq)
}

// Function type has generics[0] as argument type and generics[1] as return type.
func (t Type) IsFunction() bool {
	return t.Name == "Fn"
}

func (t Type) GetParameterType() (parameterType Type, isFunction bool) {
	isFunction = t.IsFunction()
	if !isFunction {
		return
	}
	parameterType = t.Generics[0]
	return
}

func (t Type) GetReturnType() (returnType Type, isFunction bool) {
	isFunction = t.IsFunction()
	if !isFunction {
		return
	}
	returnType = t.Generics[1]
	return
}
