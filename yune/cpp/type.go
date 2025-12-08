package cpp

import (
	"fmt"
	"strings"
	"yune/util"
)

type Type struct {
	Name     string
	Generics []Type
}

func (t Type) String() string {
	return fmt.Sprintf("%s<%s>", t.Name, strings.Join(util.Map(t.Generics, Type.String), ", "))
}
