package ast

import (
	"github.com/go-json-experiment/json"
)

type ValueOptions struct {
	TypeOptions
	ExpressionOptions
}

func (o ValueOptions) GetNonNil() Value {
	if t := o.TypeOptions.GetNonNil(); t != nil {
		return t
	}
	if e := o.ExpressionOptions.GetNonNil(); e != nil {
		return e
	}
	return nil
}

type TypeOptions struct {
	TypeType   *TypeType
	IntType    *IntType
	FloatType  *FloatType
	BoolType   *BoolType
	StringType *StringType
	NilType    *NilType
	TupleType  *TupleType
	ListType   *ListType
	FnType     *FnType
	StructType *StructType
}

func (o TypeOptions) GetNonNil() TypeValue {
	switch {
	case o.TypeType != nil:
		return o.TypeType
	case o.IntType != nil:
		return o.IntType
	case o.FloatType != nil:
		return o.FloatType
	case o.BoolType != nil:
		return o.BoolType
	case o.StringType != nil:
		return o.StringType
	case o.NilType != nil:
		return o.NilType
	case o.TupleType != nil:
		return o.TupleType
	case o.ListType != nil:
		return o.ListType
	case o.FnType != nil:
		return o.FnType
	case o.StructType != nil:
		return o.StructType
	default:
		return nil
	}
}

type ExpressionOptions struct {
	Integer          *Integer
	Float            *Float
	Bool             *Bool
	String           *String
	Variable         *Variable
	FunctionCall     *FunctionCall
	Tuple            *Tuple
	Macro            *Macro
	UnaryExpression  *UnaryExpression
	BinaryExpression *BinaryExpression
}

func (o ExpressionOptions) GetNonNil() Expression {
	switch {
	case o.Integer != nil:
		return *o.Integer
	case o.Float != nil:
		return *o.Float
	case o.Bool != nil:
		return *o.Bool
	case o.String != nil:
		return *o.String
	case o.Variable != nil:
		return o.Variable
	case o.FunctionCall != nil:
		return o.FunctionCall
	case o.Tuple != nil:
		return o.Tuple
	case o.Macro != nil:
		return o.Macro
	case o.UnaryExpression != nil:
		return o.UnaryExpression
	case o.BinaryExpression != nil:
		return o.BinaryExpression
	default:
		return nil
	}
}

type Value interface {
	value()
}

func Deserialize(jsonBytes []byte) (vs []Value) {
	var unmarshalers *json.Unmarshalers
	expressionUnmarshalers := json.UnmarshalFunc(func(jsonBytes []byte, e *Expression) error {
		options := ExpressionOptions{}
		println("HI Expression: " + string(jsonBytes))
		if err := json.Unmarshal(jsonBytes, &options, json.WithUnmarshalers(unmarshalers)); err != nil {
			return err
		}
		*e = options.GetNonNil()
		if *e == nil {
			panic("Only nil options when unmarshaling Expression")
		}
		return nil
	})
	// NOTE: only gets called with non-pointer type so we can't change the interface implementor
	typeUnmarshalers := json.UnmarshalFunc(func(jsonBytes []byte, t TypeValue) error {
		println("HI TypeValue: " + string(jsonBytes))
		return nil
		// options := TypeOptions{}
		// if err := json.Unmarshal(jsonBytes, &options, json.WithUnmarshalers(unmarshalers)); err != nil {
		// 	return err
		// }
		// *t = options.GetNonNil()
		// if *t == nil {
		// 	panic("Only nil options when unmarshaling TypeValue")
		// }
		// return nil
	})
	valueUnmarshalers := json.UnmarshalFunc(func(jsonBytes []byte, v *Value) error {
		options := ValueOptions{}
		println("HI Value: " + string(jsonBytes))
		if err := json.Unmarshal(jsonBytes, &options, json.WithUnmarshalers(unmarshalers)); err != nil {
			return err
		}
		*v = options.GetNonNil()
		if *v == nil {
			panic("Only nil options when unmarshaling Value")
		}
		return nil
	})
	unmarshalers = json.JoinUnmarshalers(expressionUnmarshalers, typeUnmarshalers, valueUnmarshalers)

	// NOTE: do we need an unmarshaler for both Value and Expression?
	err := json.Unmarshal(jsonBytes, &vs, json.WithUnmarshalers(unmarshalers))
	if err != nil {
		panic("Failed to deserialize JSON: " + err.Error())
	}
	return
}

type Destination interface {
	SetValue(v Value)
}

type SetType struct {
	Type *TypeValue
}

func (s SetType) SetValue(v Value) {
	*s.Type = v.(TypeValue)
}
