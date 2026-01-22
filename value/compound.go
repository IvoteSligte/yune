package value

type Tuple []Value

func (Tuple) value() {}

type List []Value

func (List) value() {}

var _ Value = (*Tuple)(nil)
var _ Value = (*List)(nil)
