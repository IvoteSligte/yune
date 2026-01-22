package value

type Int int64

func (Int) value() {}

type Float float64

func (Float) value() {}

type Bool bool

func (Bool) value() {}

type String string

func (String) value() {}

var _ Value = (*Int)(nil)
var _ Value = (*Float)(nil)
var _ Value = (*Bool)(nil)
var _ Value = (*String)(nil)
