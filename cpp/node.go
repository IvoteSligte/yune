package cpp

import "fmt"

type Node = fmt.Stringer

type Raw string

func (r Raw) String() string {
	return string(r)
}
