package value

type Type string

func (t Type) Eq(other Type) bool {
	return string(t) == string(other)
}
