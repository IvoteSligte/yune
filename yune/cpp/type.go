package cpp

import (
	"fmt"
	"yune/util"
)

type Type struct {
	Name     string
	Generics []Type
}

func (t Type) String() string {
	if len(t.Generics) == 0 {
		return t.Name
	} else {
		return fmt.Sprintf("%s<%s>", t.Name, util.SeparatedBy(t.Generics, ", "))
	}
}
