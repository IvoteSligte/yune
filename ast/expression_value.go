package ast

import (
	"fmt"
	"log"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

func (state *State) lowerExpressionValue(data *fj.Value) string {
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
		array, err := data.Array()
		if err == nil {
			return fmt.Sprintf(`{ []() constexpr {
static constexpr auto array[%d] = {%s};
return array;
}() }`, len(array), util.JoinFunc(array, ", ", state.lowerExpressionValue))

			// return fmt.Sprintf(
			// 	`{%s}`, util.JoinFunc(array, ", ", state.lowerExpressionValue),
			// )
		}
		log.Panicf("Tried to lower non-object JSON variant: %s", data)
	}
	key, v := fjUnmarshalUnion(object)
	switch key {
	case "Closure":
		return state.lowerClosureValue(v)
	case "Function":
		// the :: prefix makes sure the function refers to the globally declared one,
		// not a variable with the same name currently being declared
		//
		// // assume func is declared in global scope
		// std::Function<int, bool> func = func;   // func refers to the variable being declared
		// std::Function<int, bool> func = ::func; // func refers to the correct definition
		return "::" + string(v.GetStringBytes())
	case "Box":
		return fmt.Sprintf(`box([]() constexpr {
    static constexpr auto value = %s;
    return &value;
}())`, state.lowerExpressionValue(v))
	default:
		fields := ""
		v.GetObject().Visit(func(keyBytes []byte, v *fj.Value) {
			fields += fmt.Sprintf("\n    .%s = %s,", keyBytes, state.lowerExpressionValue(v))
		})
		return fmt.Sprintf("(ty::%s) {%s}", key, fields)
	}
}

// Lowers a JSON object representing a closure value, i.e. an instantiated closure.
// This does not include the name of the type "Closure".
func (state *State) lowerClosureValue(v *fj.Value) string {
	id := string(v.GetStringBytes("id"))
	closure := state.registeredClosures[id]
	if closure == nil {
		panic(fmt.Sprintf("Invalid closure ID: '%s'", id))
	}
	lowered := closure.Lower(state)
	captures := ""
	for _, capture := range v.GetArray("captures") {
		name := string(capture.GetStringBytes("name"))
		_type := state.UnmarshalTypeValue(capture.Get("type")).LowerType()
		value := state.lowerExpressionValue(capture.Get("value"))
		captures += _type + " " + name + " = " + value + ";\n"
	}
	parameters := closure.LowerParameters()
	arguments := util.JoinFunc(closure.Parameters, ", ", func(p FunctionParameter) string {
		return p.Name.String
	})
	return fmt.Sprintf(`[](%s){
    %s
    return %s(%s);
}`, parameters, captures, lowered, arguments)
}
