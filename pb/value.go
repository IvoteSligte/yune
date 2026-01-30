package pb

import (
	"fmt"
	"yune/cpp"

	capnp "capnproto.org/go/capnp/v3"
)

func ToSlice[T ~capnp.StructKind](s capnp.StructList[T]) (elements []T) {
	for i := range s.Len() {
		elements = append(elements, s.At(i))
	}
	return
}

type Destination interface {
	SetValue(v Value)
}

// Uninitialized type (not an actual Yune type)
var UninitType = Type{}

var TypeType = newUnitType(Type.SetType)
var IntType = newUnitType(Type.SetInt)
var FloatType = newUnitType(Type.SetFloat)
var BoolType = newUnitType(Type.SetBool)
var StringType = newUnitType(Type.SetString_)
var NilType = newUnitType(Type.SetNil)

var EmptyTupleType = NewTupleType()

// NOTE: main() returns int for compatibility with C++,
// though this may change in the future
var MainType = NewFnType(EmptyTupleType, IntType)

// AST types
var ExpressionType = NewStructType("Expression")

// TODO: complete macro return type with diagnostics
var MacroReturnType = NewTupleType(StringType, ExpressionType)

// SetValue implements Destination.
func (t *Type) SetValue(v Value) {
	_type, err := v.Type()
	if err != nil {
		panic("SetValue type mismatch: <unknown value> is not a Type.")
	}
	*t = _type
}

var _ Destination = &Type{}

func (t Type) Lower() cpp.Type {
	switch t.Which() {
	case Value_Type_Which_bool:
		return cpp.Type("bool")
	case Value_Type_Which_float:
		return cpp.Type("float")
	case Value_Type_Which_fn:
		fn := t.Fn()
		_return, _ := fn.Return()
		var args string
		arg, _ := fn.Argument()
		if arg.HasTuple() {
			elements, _ := arg.Tuple()
			for i := range elements.Len() {
				if i > 0 {
					args += ", "
				}
				args += elements.At(i).Lower().String()
			}
		} else {
			args = arg.Lower().String()
		}
		return cpp.Type(fmt.Sprintf("std::function<%s(%s)>", _return.Lower(), args))
	case Value_Type_Which_int:
		return cpp.Type("int")
	case Value_Type_Which_list:
		element, _ := t.List()
		return cpp.Type(fmt.Sprintf("std::vector<%s>", element))
	case Value_Type_Which_nil:
		return cpp.Type("void")
	case Value_Type_Which_string_:
		return cpp.Type("std::string")
	case Value_Type_Which_struc:
		name, _ := t.Struc().Name()
		return cpp.Type(fmt.Sprintf("%s_type_", name))
	case Value_Type_Which_tuple:
		elements, _ := t.Tuple()
		s := elements.At(0).Lower().String()
		for i := 1; i < elements.Len(); i++ {
			s += ", "
			s += elements.At(i).Lower().String()
		}
		return cpp.Type(fmt.Sprintf("std::tuple<%s>", s))
	case Value_Type_Which_type:
		return cpp.Type("Type")
	default:
		panic("unexpected pb.Value_Type_Which")
	}
}

func (left Type) Eq(right Type) bool {
	eq, _ := capnp.Equal(left.ToPtr(), right.ToPtr())
	return eq
}

func (t Type) ToFunction() (fn FnType, ok bool) {
	if ok = (t.Which() == Value_Type_Which_fn); !ok {
		return
	}
	rawFn := t.Fn()
	arg, _ := rawFn.Argument()
	ret, _ := rawFn.Return()
	fn = FnType{
		ArgumentType: arg,
		ReturnType:   ret,
	}
	return
}

func (t Type) ToTuple() (tuple TupleType, ok bool) {
	if ok = (t.Which() == Value_Type_Which_tuple); !ok {
		return
	}
	rawTuple, _ := t.Tuple()
	tuple = TupleType{
		Elements: ToSlice(rawTuple),
	}
	return
}

func (t Type) IsTuple() bool {
	_, isTuple := t.ToTuple()
	return isTuple
}

type FnType struct {
	// Type of the argument that the function takes.
	// Is always represented as a tuple, even with only one argument.
	ArgumentType Type
	ReturnType   Type
}

type TupleType struct {
	Elements []Type
}

func NewFnType(argumentType Type, returnType Type) Type {
	if argumentType.Which() != Value_Type_Which_tuple {
		argumentType = NewTupleType(argumentType)
	}
	t := newType()
	t.SetFn()
	t.Fn().SetArgument(argumentType)
	t.Fn().SetReturn(returnType)
	return t
}

func NewTupleType(elementTypes ...Type) Type {
	t := newType()
	t.SetTuple(newTypeList(elementTypes...))
	return t
}

func NewListType(elementType Type) Type {
	t := newType()
	t.SetList(elementType)
	return t
}

func NewStructType(name string) Type {
	t := newType()
	t.SetStruc()
	t.Struc().SetName(name)
	return t
}
