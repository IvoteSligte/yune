package ast

import (
	"encoding/json"
	"reflect"
	"yune/util"
)

var typeToTag = map[reflect.Type]string{
	// Expressions
	reflect.TypeFor[Integer]():          "Integer",
	reflect.TypeFor[Float]():            "Float",
	reflect.TypeFor[Bool]():             "Bool",
	reflect.TypeFor[String]():           "String",
	reflect.TypeFor[Variable]():         "Variable",
	reflect.TypeFor[FunctionCall]():     "FunctionCall",
	reflect.TypeFor[Tuple]():            "Tuple",
	reflect.TypeFor[UnaryExpression]():  "UnaryExpression",
	reflect.TypeFor[BinaryExpression](): "BinaryExpression",
}

var tagToType = util.MapMap(typeToTag, func(_type reflect.Type, tag string) (string, reflect.Type) {
	return tag, _type
})

func Marshal(v any) (_ []byte, err error) {
	r := reflect.ValueOf(v)
	t := r.Type()
	switch t.Kind() {
	case reflect.Interface, reflect.Pointer:
		return Marshal(r.Elem())
	case reflect.Struct:
		tag, ok := typeToTag[t]
		if !ok {
			panic("Tried to serialize struct type " + t.String() + " that does not have known tag.")
		}
		m := map[string][]byte{}
		m["$tag"] = []byte(tag)
		for i := range t.NumField() {
			field := t.Field(i)
			m[field.Name], err = Marshal(r.Field(i))
			if err != nil {
				return
			}
		}
		return json.Marshal(m)
	case reflect.Int, reflect.Float64, reflect.Bool, reflect.String:
		return json.Marshal(v)
	default:
		panic("Tried to serialize unserializable type " + t.String())
	}
}

func Unmarshal(data []byte) (v any, err error) {
	var tagged struct {
		Tag string `json:"$tag"`
	}
	err = json.Unmarshal(data, &tagged)
	if err != nil {
		return
	}
	if tagged.Tag == "" {
		// assume unambiguous
		err = json.Unmarshal(data, &v)
		return
	}
	t, ok := tagToType[tagged.Tag]
	if !ok {
		panic("Tried to deserialize JSON with unknown tag " + tagged.Tag)
	}
	m := map[string]json.RawMessage{}
	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	var r = reflect.Zero(t)
	for i := range t.NumField() {
		field := r.Field(i)
		v, err = Unmarshal(m[t.Field(i).Name])
		if err != nil {
			return
		}
		field.Set(reflect.ValueOf(v))
	}
	v = r.Interface()
	return
}
