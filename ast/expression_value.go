package ast

import (
	"fmt"
	"yune/util"

	fj "github.com/valyala/fastjson"
)

func (state *State) lowerObject(object *fj.Object) string {
	typeName, v := fjUnmarshalStruct(object)
	switch typeName {
	case "Closure":
		return state.lowerClosureValue(v)
	case "Function":
		// the :: prefix makes sure the function refers to the globally declared one,
		// not a variable with the same name currently being declared
		//
		// // assume func is declared in global scope
		// std::Function<int, bool> func = func;   // func refers to the variable being declared
		// std::Function<int, bool> func = ::func; // func refers to the correct definition
		return "::" + UnmarshalNonEmptyString(v)
	case "Box":
		return fmt.Sprintf(`box_f(%s)`, state.lowerExpressionValue(v))
	case "Tuple":
		elements := UnmarshalArray(v, "elements")
		return fmt.Sprintf(`std::make_tuple(%s)`, util.JoinFunc(elements, ", ", state.lowerExpressionValue))
	default:
		fields := ""
		v.GetObject().Visit(func(keyBytes []byte, fieldValue *fj.Value) {
			fields += fmt.Sprintf("\n    .%s = %s,", keyBytes, state.lowerExpressionValue(fieldValue))
		})
		return fmt.Sprintf("%s_t { %s }", typeName, fields)
	}
}

func (state *State) lowerExpressionValue(data *fj.Value) string {
	switch data.Type() {
	case fj.TypeArray:
		return fmt.Sprintf(`{ %s }`, util.JoinFunc(UnmarshalArray(data), ", ", state.lowerExpressionValue))
	case fj.TypeString:
		return fmt.Sprintf("String_t(%q)", data.GetStringBytes())
	case fj.TypeObject:
		return state.lowerObject(data.GetObject())
	case fj.TypeNumber, fj.TypeTrue, fj.TypeFalse:
		return data.String()
	default:
		panic(fmt.Sprintf("unexpected fj.Type: %s", data.Type()))
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
	captureValues := map[string]string{}
	for _, capture := range UnmarshalArray(v, "captures") {
		name := string(capture.GetStringBytes("name"))
		value := state.lowerExpressionValue(capture.Get("value"))
		captureValues[name] = value
	}
	return closure.LowerComplex(state, captureValues)
}
