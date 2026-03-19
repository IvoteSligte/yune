package main

import (
	"fmt"
	"testing"
)

func assert(condition bool) {
	if !condition {
		panic("Assertion failed.")
	}
}

func assertEq[T comparable](left T, right T) {
	if left != right {
		panic(fmt.Sprintf(`Assertion failed. left != right.
    left: %#v
    right: %#v`, left, right))
	}
}

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

func TestFunctionDeclaration(t *testing.T) {
	stdout, _ := runModule("functionDeclaration.un", `
hello(user: String): () =
    println("Hello, " + user + "!")

main(): () =
    hello "World"
`)
	assertEq(stdout, "Hello, World!\n")
}

// Tests function call syntax
func TestFunctionCall(t *testing.T) {
	stdout, _ := runModule("functionCall.un", `
main(): () =
    v := "text"
    println v
    println(v + v)
    println v + v
    println ;true
`)
	assertEq(stdout, `text
texttext
texttext
false
`)
}

func TestPrecedence(t *testing.T) {
	stdout, _ := runModule("precedence.un", `
main(): () =
    println 1 * 2 + 3
    println true and true or false
    println 1 - 2 // could be incorrectly interpreted as println(1(-2))
`)
	assertEq(stdout, `5
true
`)
}

func TestExpressionCreation(t *testing.T) {
	runModule("expressionCreation.un", `
main(): () =
    leftLeft := stringLiteral("leftLeft")
    leftRight := functionCall(variable("toString"), variable(captureName))
    left := binaryExpression("+", leftLeft, leftRight)
    binary: Expression = binaryExpression("+", left, right)
`)
}

func TestBasic(t *testing.T) {
	stdout, _ := runModule("basic.un", `
N: Int = 12

fibonacci(n: Int): Int =
    n == 0 -> 0
    n == 1 -> 1
    fibonacci(n - 1) + fibonacci(n - 2)

add(a: Int, b: Int): Int = a + b

something(f: Fn((), Int)): Int = f()

takeTuple(t: (Int, String)): String =
    "literal"

longString(text: String, getType: Fn(String, Type)): Union[String, Expression] =
    getType("add") // test getType
    stringLiteral(text)

noArguments(): () = a: Int = 0

main(): () =
    a: Int = fibonacci(add(N, N))
    true and false
    "string galore"
    noArguments(noArguments())
    f: Fn((), Int) = ||: Int = 0
    f()

    s := longString#This is a very long,
        multi-line string.
        It contains several newlines.
        Something fancy it supports is quotes "" and even hashtags#!

    println s
`)
	assertEq(stdout, `This is a very long,
        multi-line string.
        It contains several newlines.
        Something fancy it supports is quotes "" and even hashtags#!

`) // FIXME: get rid of trailing newlines
}
