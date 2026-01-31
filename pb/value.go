package pb

import (
	"fmt"
	"yune/cpp"
	"yune/util"
)

type SwigArray[T any] interface {
	Size() int64
	Get(i int) T
}

func ToSlice[T any](s SwigArray[T]) (elements []T) {
	for i := range s.Size() {
		elements = append(elements, s.Get(int(i)))
	}
	return
}

type Destination interface {
	SetValue(v Value)
}

type SetType struct {
	Type *Type
}

// SetValue implements Destination
func (s SetType) SetValue(v Value) {
	*s.Type = v.(Type)
}

// Uninitialized type
var UninitType Type = nil

var EmptyTupleType = NewTupleType(NewTypeVector())

// NOTE: main() returns int for compatibility with C++,
// though this may change in the future
var MainType = NewFnType(EmptyTupleType, NewIntType())

// AST types
var ExpressionType = NewStructType("Expression")

// TODO: complete macro return type with diagnostics
var MacroReturnType = NewTupleType(NewTypeVector(NewStringType(), ExpressionType))

var _ Destination = (*SetType)(nil)

func LowerType(t Type) cpp.Type {
	switch t := t.(type) {
	case BoolType:
		return cpp.Type("bool")
	case FloatType:
		return cpp.Type("float")
	case FnType:
		returnType := t.GetReturnType()
		var args string
		arg := t.GetArgument()
		tupleArg, ok := arg.(TupleType)
		if ok {
			elements := tupleArg.GetElements()
			args = util.JoinFunction(ToSlice(elements), ", ", func(t Type) string {
				return LowerType(t).String()
			})
		} else {
			args = LowerType(arg).String()
		}
		return cpp.Type(fmt.Sprintf("std::function<%s(%s)>", LowerType(returnType), args))
	case IntType:
		return cpp.Type("int")
	case ListType:
		element := t.GetElement()
		return cpp.Type(fmt.Sprintf("std::vector<%s>", LowerType(element)))
	case NilType:
		return cpp.Type("void")
	case StringType:
		return cpp.Type("std::string")
	case StructType:
		name := t.GetName()
		return cpp.Type(fmt.Sprintf("%s_type_", name))
	case TupleType:
		elements := util.JoinFunction(ToSlice(t.GetElements()), ", ", func(t Type) string {
			return LowerType(t).String()
		})
		return cpp.Type(fmt.Sprintf("std::tuple<%s>", elements))
	case TypeType:
		return cpp.Type("Type")
	default:
		panic("unexpected pb.Type_Which")
	}
}
