package ast

import (
	"yune/cpp"

	fj "github.com/valyala/fastjson"
)

// TODO: macros in types

type Type struct {
	// Evaluated expression
	value      TypeValue
	Expression Expression
}

// TODO: rename to GetValue
func (t Type) Get() TypeValue {
	return t.value
}

func (t Type) Lower() cpp.Type {
	return t.value.LowerType()
}

func UnmarshalType(data *fj.Value) Type {
	return Type{
		Expression: UnmarshalExpression(data),
	}
}

func (t *Type) Analyze(anal Analyzer) TypeValue {
	if t.value != nil {
		panic("Called Type.Analyze on already-analyzed type.")
	}
	expressionType := t.Expression.Analyze(&TypeType{}, anal.TopLevel())
	// TODO: check if expressionType is part of the union TypeType rather than equal
	// (is this necessary?)
	if !expressionType.Eq(&TypeType{}) {
		anal.ReportError(UnexpectedType{
			Expected: &TypeType{},
			Found:    t.value,
			At:       t.Expression.GetSpan(),
		})
	}
	json := anal.Evaluate(t.Expression.Lower(anal.State))
	t.value = anal.State.UnmarshalTypeValue(json)
	return t.value
}
