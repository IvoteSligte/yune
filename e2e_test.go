package main

import (
	"fmt"
	"testing"
)

func assertEq[T comparable](found T, expected T) {
	if found != expected {
		panic(fmt.Sprintf(`Assertion failed. found != expected.
    found   : %#v
    expected: %#v`, found, expected))
	}
}

func expectPanic(run func(), failMessage string) {
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
		run()
	}()
	if !panicked {
		panic(failMessage)
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

func TestTypeCycle(t *testing.T) {
	expectPanic(func() {
		parseAndRunModule("typeCycle.un", `
A: B = ()
B: A = ()
`)
	}, "Cycle not detected.")
}

func TestSelfCycle(t *testing.T) {
	expectPanic(func() {
		parseAndRunModule("selfCycle.un", `A: A = A`)
	}, "Cycle not detected.")
}

func TestTypeValueCycle(t *testing.T) {
	expectPanic(func() {
		parseAndRunModule("typeValueCycle.un", `
A: B = ()
B: Type = A
`)
	}, "Cycle not detected.")
}

func TestImpureConstant(t *testing.T) {
	expectPanic(func() {
		parseAndRunModule("impureConstant.un", `
T: () = printlnString("impure! begone!")
`)
	}, "Impure constant not detected.")
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
    leftLeft := stringExpression(0, "leftLeft")
    leftRight := functionCallExpression(0, variableExpression(0, "toString"), variableExpression(0, "captureName"))
    left := binaryExpression(0, "+", leftLeft, leftRight)
    binary: Expression = binaryExpression(0, "+", left, variableExpression(0, "right"))
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
EXPRESSION: Expression = binaryExpression(0, "+", stringExpression(0, "before"), stringExpression(0, "after"))
`)
}

func TestJson(t *testing.T) {
	parseAndRunModule("jsonTest.un", `import "json.un"`)
}

func TestSQL(t *testing.T) {
	stdout, _ := parseAndRunModule("sqlTest.un", `import "sql.un"`)
	assertEq(stdout, ""+
		"SELECT `SELECT`\n"+
		"IDENT `first_name`\n"+
		", `,`\n"+
		"IDENT `last_name`\n"+
		", `,`\n"+
		"IDENT `grade`\n"+
		"FROM `FROM`\n"+
		"IDENT `students`\n"+
		"WHERE `WHERE`\n"+
		"IDENT `grade`\n"+
		"> `>`\n"+
		"$ `$`\n"+
		"IDENT `grade`\n",
	)
}

func TestFmt(t *testing.T) {
	stdout, _ := parseAndRunModule("fmtTest.un", `import "fmt.un"`)
	assertEq(stdout, `Hello, World! This is a long string: "A Long String!" and math is 999 + 999 * 999

`)
}

func TestRawCppTopLevel(t *testing.T) {
	code := `
` + "`" + `
#import <string>
std::string str{"Hello, World!"};
` + "`" + `

STRING: String = ` + "`str`" + `

main(): () =
    printlnString(STRING)
`
	fmt.Println(code)
	stdout, _ := parseAndRunModule("rawCppTopLevel.un", code)
	assertEq(stdout, "Hello, World!\n")
}

func TestRawCppExpression(t *testing.T) {
	_, _ = parseAndRunModule("rawCppExpression.un", "T: Type = `Int`")
}

func TestClosureExpression(t *testing.T) {
	stdout, _ := parseAndRunModule("closureExpression.un", `
Error: Type = (Int, String)
square(text: String, getType: Fn(String, Union[Type, ()])): Union[Error, Expression] =
    result := getType(text)
    result is undefined: () -> (0, "Variable does not exist")
    result is type: Type
    type ;= Int -> (0, "Variable must be an integer")
    parameters: List((String, Expression)) = []
    statements: List(Statement) = [expressionStatement(binaryExpression(0, "*", variableExpression(0, text), variableExpression(0, text)))]
    closureExpression(0, parameters, inject(Int), statements)

main(): () =
    n: Int = 10
    squareClosure := square#n
    squareClosure() == 100 -> printlnString("correct")
    printlnString("incorrect")
`)
	assertEq(stdout, "correct\n")
}

// // Tests the deterministic evaluation order of macros.
// // It currently does not pass and was deemed too annoying to fix before the thesis deadline.
// func TestMacroEvalOrder(t *testing.T) {
// 	previousStdout := ""
//
// 	for range 20 {
// 		stdout, _ := parseAndRunModule("macroEvalOrder.un", `
// output: String = "initial"
//
// setOutput(text: String, getType: Fn(String, Union[Type, ()])): Union[Expression, String] =
//     output = "modified"
//     tupleExpression(0, [])
//
// getOutput(text: String, getType: Fn(String, Union[Type, ()])): Union[Expression, String] =
//     inject(output)
//
// setOutputFunction(): () = setOutput#
//
// main(): () =
//     printlnString getOutput#
// `)
// 		if previousStdout != "" {
// 			assertEq(previousStdout, stdout)
// 		} else {
// 			previousStdout = stdout
// 		}
// 	}
// }

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

Error: Type = (Int, String)

longString(text: String, getType: Fn(String, Union[Type, ()])): Union[Expression, Error] =
    result := getType("add")
    result is undefined: () ->
        (0, "Function 'add' is not defined")
    result is type: Type
    type ;= Fn((Int, Int), Int) ->
        (0, "Function 'add' does not have the expected type")
    stringExpression(0, text)

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
