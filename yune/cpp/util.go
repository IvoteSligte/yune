package cpp

import (
	"fmt"
	"strings"
	"yune/util"
)

func separatedBy[T fmt.Stringer](array []T, separator string) string {
	return strings.Join(util.Map(array, T.String), separator)
}
