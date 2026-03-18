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

func TestFunctions(t *testing.T) {
	runModule("functions.un", `
hello(user: String): () =
    println("Hello, " + user + "!")

main(): () =
    hello "World"
`)
}
