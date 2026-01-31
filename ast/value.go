package ast

type Value interface {
	value()
}

func Deserialize(jsonBytes []byte) (v Value) {
	unmarshaler := oneof.UnmarshalFunc(valueOptions, nil)
	err := json.Unmarshal(jsonBytes, &v, json.WithUnmarshalers(unmarshaler))
	if err != nil {
		panic("Failed to deserialize JSON: " + err.Error())
	}
	return
}
