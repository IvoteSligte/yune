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

func TestExpressions(t *testing.T) {
	runModule("expressions.un", `
main(): () =
    leftLeft := stringLiteral("leftLeft")
    leftRight := functionCall(variable("toString"), variable(captureName))
    left := binaryExpression("+", leftLeft, leftRight)
    binary: Expression = binaryExpression("+", left, right)
`)
}

func TestBasic(t *testing.T) {
	runModule("basic.un", `
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
}
