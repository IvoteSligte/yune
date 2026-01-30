package pb

import (
	"log"

	capnp "capnproto.org/go/capnp/v3"
)

type Type = Value_Type
type TypeList = Value_Type_List

func newType() Type {
	_, s := capnp.NewSingleSegmentMessage(nil)
	t, err := NewRootValue_Type(s)
	if err != nil {
		log.Fatalf("Failed to create Capnp root: %s", err)
	}
	return t
}

func newTypeList(types ...Type) TypeList {
	_, s := capnp.NewSingleSegmentMessage(nil)
	tl, err := NewValue_Type_List(s, int32(len(types)))
	if err != nil {
		log.Fatalf("Failed to create Capnp type list: %s", err)
	}
	return tl
}

func newUnitType(f func(Type)) Type {
	t := newType()
	f(t)
	return t
}
