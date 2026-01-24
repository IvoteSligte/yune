package value

type Value string

type Type string

func (t Type) Eq(other Type) bool {
	return string(t) == string(other)
}

func (t Type) String() string {
	return string(t)
}
