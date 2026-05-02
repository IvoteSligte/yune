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

func assertEq[T comparable](found T, expected T) {
	if found != expected {
		panic(fmt.Sprintf(`Assertion failed. found != expected.
    found   : %#v
    expected: %#v`, found, expected))
	}
}

func TestPrimitives(t *testing.T) {
	parseAndRunModule("primitives.un", `
main(): () =
    true and false
    "string literal!#%\n\t\"\\"
    965.102
    59342168
    ()
`)
}

func TestImpureConstant(t *testing.T) {
	panicked := false
	// defer runs at function exit, so we wrap this in another function to
	// not accidentally recover from the later panic
	func() {

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic:", r)
			}
			panicked = true
		}()
		parseAndRunModule("impureConstant.un", `
T: () = printlnString("impure! begone!")
`)
	}()
	if !panicked {
		panic("Impure constant not detected.")
	}
}

func TestParsing(t *testing.T) {
	parseAndRunModule("parsing.un", `

doNothing(): () = ()

function(argument: Int, another: Fn(Float, Float)): (Int, String) =
    doNothing()
    (argument + 5, "a string")

main(): () = ()
`)
}

func TestFunctionDeclaration(t *testing.T) {
	stdout, _ := parseAndRunModule("functionDeclaration.un", `
import "std.un"

hello(user: String): () =
    println("Hello, " + user + "!")

main(): () =
    hello "World"
`)
	assertEq(stdout, "Hello, World!\n")
}

// Tests function call syntax
func TestFunctionCall(t *testing.T) {
	stdout, _ := parseAndRunModule("functionCall.un", `
import "std.un"

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

func TestPrintln(t *testing.T) {
	stdout, _ := parseAndRunModule("println.un", `
import "std.un"

main(): () =
    println 0
    println 5
    println 11
    println 49329
    println(-99998)
    println "a\nstring"
    println true
    println ()
`)
	assertEq(stdout, `0
5
11
49329
-99998
a
string
true
()
`)
}

func TestPrecedence(t *testing.T) {
	stdout, _ := parseAndRunModule("precedence.un", `
import "std.un"

main(): () =
    println 4 * 2 + 3
    println true and true or false
    println 1 - 2 // could be incorrectly interpreted as println(1(-2))
`)
	assertEq(stdout, `11
true
-1
`)
}

func TestExpressionCreation(t *testing.T) {
	parseAndRunModule("expressionCreation.un", `
import "std.un"

main(): () =
    leftLeft := stringExpression("leftLeft")
    leftRight := functionCallExpression(variableExpression("toString"), variableExpression("captureName"))
    left := binaryExpression("+", leftLeft, leftRight)
    binary: Expression = binaryExpression("+", left, variableExpression("right"))
`)
}

func TestStaticInitialization(t *testing.T) {
	parseAndRunModule("staticInitialization.un", `
// Tests string serialization
STRING: String = "string" + "another"

// Tests indirect Union construction
UNION: Union[String, Int] = "chars"

// Tests serialization of List
LIST: List(Union[String, Int]) = ["alpha", "beta", "gamma", 50]

// Tests serialization of Box
EXPRESSION: Expression = binaryExpression("+", stringExpression("before"), stringExpression("after"))
`)
}

func TestRawCpp(t *testing.T) {
	_, _ = parseAndRunModule("rawCpp.un", "T: Type = `Int`")
}

func TestRecursiveType(t *testing.T) {
	_, _ = parseAndRunModule("recursiveType.un", "R: Type = List(`box_f(StructType_t{.name = \"R\"})`)")
}

func TestBasic(t *testing.T) {
	stdout, _ := parseAndRunModule("basic.un", `
import "std.un"

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
    stringExpression(text)

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


`) // TODO: get rid of (leading and) trailing newlines or whitespace in the macro parsing
}
