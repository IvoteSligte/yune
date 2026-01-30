package value

import (
	"fmt"
	"yune/cpp"
	"yune/pb"
	"yune/util"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Destination interface {
	SetValue(v *anypb.Any)
}

type Type struct {
	pb *pb.Type
}

// Uninitialized type (not an actual Yune type)
var UninitType = Type{}

var TypeType = Type{&pb.Type{Kind: pb.Type_TYPE}}
var IntType = Type{&pb.Type{Kind: pb.Type_INT}}
var FloatType = Type{&pb.Type{Kind: pb.Type_FLOAT}}
var BoolType = Type{&pb.Type{Kind: pb.Type_BOOL}}
var StringType = Type{&pb.Type{Kind: pb.Type_STRING}}
var NilType = Type{&pb.Type{Kind: pb.Type_NIL}}

var EmptyTupleType = NewTupleType([]Type{})

// NOTE: main() returns int for compatibility with C++,
// though this may change in the future
var MainType = NewFnType(EmptyTupleType, IntType)

// AST types
var ExpressionType = NewStructType("Expression")

// TODO: complete macro return type with diagnostics
var MacroReturnType = NewTupleType([]Type{StringType, ExpressionType})

// SetValue implements Destination.
func (t *Type) SetValue(v *anypb.Any) {
	if err := v.UnmarshalTo(t.pb); err != nil {
		panic("Protobuf unmarshalling error: " + err.Error())
	}
}

var _ Destination = &Type{}

func (t Type) Lower() cpp.Type {
	switch t.pb.Kind {
	case pb.Type_BOOL:
		return cpp.Type("bool")
	case pb.Type_FLOAT:
		return cpp.Type("float")
	case pb.Type_FN:
		fn := t.pb.GetFn()
		_return := Type{fn.ReturnType}.Lower()
		var args string
		if fn.ArgumentType.Kind == pb.Type_TUPLE {
			elements := fn.ArgumentType.GetTuple().Elements
			args = util.JoinFunction(elements, ", ", func(elem *pb.Type) string {
				return Type{elem}.String()
			})
		} else {
			args = Type{fn.ArgumentType}.Lower().String()
		}
		return cpp.Type(fmt.Sprintf("std::function<%s(%s)>", _return, args))
	case pb.Type_INT:
		return cpp.Type("int")
	case pb.Type_LIST:
		return cpp.Type(fmt.Sprintf("std::vector<%s>", t.pb.GetList().GetElement()))
	case pb.Type_STRING:
		return cpp.Type("std::string")
	case pb.Type_TUPLE:
		elements := util.JoinFunction(t.pb.GetTuple().GetElements(), ", ", func(elem *pb.Type) string {
			return Type{elem}.Lower().String()
		})
		return cpp.Type(fmt.Sprintf("std::tuple<%s>", elements))
	case pb.Type_TYPE:
		return cpp.Type("Type")
	case pb.Type_NIL:
		return cpp.Type("void")
	case pb.Type_STRUCT:
		return cpp.Type(fmt.Sprintf("%s_type_", t.pb.GetStruct().Name))
	default:
		panic(fmt.Sprintf("unexpected pb.Type_Kind: %#v", t.pb.Kind))
	}
}

func (t Type) String() string {
	switch t.pb.Kind {
	case pb.Type_BOOL:
		return "bool"
	case pb.Type_FLOAT:
		return "float"
	case pb.Type_FN:
		fn := t.pb.GetFn()
		return fmt.Sprintf("Fn(%s, %s)", Type{fn.ArgumentType}, Type{fn.ReturnType})
	case pb.Type_INT:
		return "Int"
	case pb.Type_LIST:
		return fmt.Sprintf("List(%s)", Type{t.pb.GetList().Element})
	case pb.Type_STRING:
		return "String"
	case pb.Type_TUPLE:
		elements := util.JoinFunction(t.pb.GetTuple().GetElements(), ", ", func(elem *pb.Type) string {
			return Type{elem}.String()
		})
		return fmt.Sprintf("(%s)", elements)
	case pb.Type_TYPE:
		return "Type"
	case pb.Type_NIL:
		return "Nil"
	case pb.Type_STRUCT:
		return t.pb.GetStruct().Name
	default:
		panic(fmt.Sprintf("unexpected pb.Type_Kind: %#v", t.pb.Kind))
	}

}

func (left Type) Eq(right Type) bool {
	return proto.Equal(left.pb, right.pb)
}

func (t Type) ToFunction() (fn FnType, ok bool) {
	if ok = (t.pb.Kind == pb.Type_FN); !ok {
		return
	}
	rawFn := t.pb.Detail.(*pb.Type_Fn_).Fn
	fn = FnType{
		ArgumentType: Type{rawFn.ArgumentType},
		ReturnType:   Type{rawFn.ReturnType},
	}
	return
}

func (t Type) ToTuple() (tuple TupleType, ok bool) {
	if ok = (t.pb.Kind == pb.Type_TUPLE); !ok {
		return
	}
	rawTuple := t.pb.Detail.(*pb.Type_Tuple_).Tuple
	tuple = TupleType{
		Elements: util.Map(rawTuple.Elements, func(elem *pb.Type) Type {
			return Type{elem}
		}),
	}
	return
}

func (t Type) IsTuple() bool {
	_, isTuple := t.ToTuple()
	return isTuple
}

type FnType struct {
	ArgumentType Type
	ReturnType   Type
}

type TupleType struct {
	Elements []Type
}

func NewFnType(argumentType Type, returnType Type) Type {
	return Type{&pb.Type{
		Kind: pb.Type_FN,
		Detail: &pb.Type_Fn_{Fn: &pb.Type_Fn{
			ArgumentType: argumentType.pb,
			ReturnType:   returnType.pb,
		}},
	}}
}

func NewTupleType(elementTypes []Type) Type {
	return Type{&pb.Type{
		Kind: pb.Type_TUPLE,
		Detail: &pb.Type_Tuple_{Tuple: &pb.Type_Tuple{
			Elements: util.Map(elementTypes, func(t Type) *pb.Type {
				return t.pb
			}),
		}},
	}}
}

func NewListType(elementType Type) Type {
	return Type{&pb.Type{
		Kind: pb.Type_LIST,
		Detail: &pb.Type_List_{List: &pb.Type_List{
			Element: elementType.pb,
		}},
	}}
}

func NewStructType(name string) Type {
	return Type{&pb.Type{
		Kind: pb.Type_STRUCT,
		Detail: &pb.Type_Struct_{Struct: &pb.Type_Struct{
			Name: name,
		}},
	}}
}
