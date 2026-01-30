package pb

import (
	"log"

	capnp "capnproto.org/go/capnp/v3"
)

func (v Value) IsEmpty() bool {
	return v.Which() == Value_Which_empty
}

func newValue() Value {
	_, s := capnp.NewSingleSegmentMessage(nil)
	v, err := NewRootValue(s)
	if err != nil {
		log.Fatalf("Failed to create Capnp root: %s", err)
	}
	return v
}

func newType() Type {
	_, s := capnp.NewSingleSegmentMessage(nil)
	t, err := NewRootType(s)
	if err != nil {
		log.Fatalf("Failed to create Capnp root: %s", err)
	}
	return t
}

func newTypeList(types ...Type) Type_List {
	_, s := capnp.NewSingleSegmentMessage(nil)
	tl, err := NewType_List(s, int32(len(types)))
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
