package cpp

import "fmt"

type Node = fmt.Stringer

// Raw code for builtin code that cannot be expressed in Yune.
type Raw string

func (r Raw) String() string {
	return string(r)
}
