package main

import "testing"

func TestPrimitives(t *testing.T) {
	runModule("primitives.un", `
main(): () =
    true and false
    "string literal!#%"
    965.102
    59342168
    ()
`)
}
