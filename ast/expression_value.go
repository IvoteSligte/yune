package ast

import (
	"fmt"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

func lowerExpressionValue(data *fj.Value) string {
	object := data.GetObject()
	if object == nil { // primitive (Yune does not produce top-level arrays)
		integer, err := data.Int64()
		if err == nil {
			return fmt.Sprintf("%v", integer)
		}
		float, err := data.Float64()
		if err == nil {
			return fmt.Sprintf("%v", float)
		}
		boolean, err := data.Bool()
		if err == nil {
			return fmt.Sprintf("%t", boolean)
		}
		string, err := data.StringBytes()
		if err == nil {
			return fmt.Sprintf("%q", string)
		}
	}
	key, v := fjUnmarshalUnion(object)
	switch key {
	case "Closure":
		return lowerClosureValue(v)
	case "Function":
		return string(v.GetStringBytes())
	case "Box":
		return fmt.Sprintf(`box(%s)`, lowerExpressionValue(v))
	default:
		fields := ""
		v.GetObject().Visit(func(keyBytes []byte, v *fj.Value) {
			fields += fmt.Sprintf("\n    .%s = %s,", keyBytes, lowerExpressionValue(v))
		})
		return fmt.Sprintf("(ty::%s) {%s}", key, fields)
	}
}

// Lowers a JSON object representing a closure value, i.e. an instantiated closure.
// This does not include the name of the type "Closure".
func lowerClosureValue(v *fj.Value) string {
	id := string(v.GetStringBytes("id"))
	closure := registeredNodes[id].(*Closure)
	lowered := closure.Lower()
	captures := ""
	for _, capture := range v.GetArray("captures") {
		name := string(capture.GetStringBytes("name"))
		_type := UnmarshalTypeValue(capture.Get("type")).Lower()
		value := lowerExpressionValue(capture.Get("value"))
		captures += _type + " " + name + " = " + value + ";\n"
	}
	parameters := closure.LowerParameters()
	arguments := util.JoinFunction(closure.Parameters, ", ", func(p FunctionParameter) string {
		return p.Name.String
	})
	return fmt.Sprintf(`[](%s){
    %s
    return %s(%s);
}`, parameters, captures, lowered, arguments)
}
