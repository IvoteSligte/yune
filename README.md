# Yune

## Introduction

Yune is a programming language made for my Bachelor's thesis. It has a powerful metaprogramming system and the ability to flawlessly interoperate with C++, which it compiles to. This `README` only contains information related to running the code. For the theory, read [my thesis](thesis.pdf).

## Running the code

The compiler can theoretically be run on Linux, Windows, and MacOS, but it has only been tested on Linux (Fedora). The following executables must to be in your `PATH`: `go`, `clang++`, and `clang-repl`. `clang++` must support at least `C++23`. The code has been tested using `go1.25.10` and `LLVM/clang` version `21.1.8`.

A Yune file can be compiled using `go . -- <file.un>`. Other files that this file imports are automatically loaded. Note that files are imported by path, there is no standard location for libraries. The standard library [`std.un`](std.un) is a regular file.

A simple example:
```
#import "std.un"

main(): () =
    name := "World"
    println("Hello, " + name + "!")
```

## Features

Yune supports tagged unions through the `Union[A, B, C, ...]` type (`std::variant` in C++) and tuples using `(A, B, C)` (`std::tuple` in C++). The language has primitive types `Int`, `Float`, `Bool`, and `String`, which translate to `int`, `float`, `bool`, and `std::string` in C++.

The `is` operator can be used to check which variant a `Union` is. Example:
```
someUnion is num: Int -> doStuffIfTrue()
doStuffIfFalse()

// panics if someUnion is not an Int
someUnion is num: Int
doStuff()
```

Other types can be introduced using C++ interoperation as follows:
```
`
// This quoted section is C++ code.
// C++ types must end in '_t'.
struct NewType_t {};
`
// A C++ type must be declared in Yune like this. The quotes indicate a C++ expression.
// Yune partially supports structs. They can be used as opaque types.
NewType: Type = `box_f(StructType_t{.name = "NewType"})`
```

See [`json.un`](json.un), [`sql.un`](sql.un), and [`fmt.un`](fmt.un) for more complex examples of Yune code, including its macro system. Tests for the compiler can also be seen in [`e2e_test.go`](e2e_test.go).

## Yune libraries

If a Yune source file does not contain a `main` function, then a file named `library.hpp` is produced, which can be included in any C++ project. The compiler file `cpp/pb.hpp` should be included with `-I<path_to_pb.hpp_folder>` to ensure the definitions which `library.hpp` relies on exist. The standard library should be set to at least version C++23 or GNU++23 (flag `-std=c++23` or `-std=gnu++23`).

In the case that the target project is not a C++ project, it is recommended to compile the `library.hpp` file to a dynamic library file with a static C++ standard library, which can easily be linked against from any language. `extern "C"` wrappers need to be made for the required functions.

## IDE Tools

Yune currently only has an Emacs mode that provides syntax highlighting in [`yune-mode.el`](yune-mode.el).

## Builtins

```
Primitive types: Type, Int, Float, Bool, String, List(T), Fn(Arg/TupleOfArgs, Return), Union(List(Type)), Expression, Statement

Functions for creating Expressions:
- integerExpression(location: Int, value: Int)
- floatExpression(Int, Float)
- boolExpression(Int, Bool)
- stringExpression(Int, String)
- variableExpression(Int, name: String)
- unaryExpression(Int, op: String, Expression)
- binaryExpression(Int, op: String, left: Expression, right: Expression)
- functionCallExpression(Int, function: Expression, arg: Expression)
- closureExpression(Int, params: List((name: String, type: Expression)), returnType: Expression, body: List(Statement))
- macroExpression(Int, name: String, text: String)
- listExpression(Int, elements: List(Expression))
- tupleExpression(Int, elements: List(Expression))
- inject(T)

Functions for creating Statements:
- variableDeclaration(name: String, type: Expression, body: List(Statement))
- assignStatement(target: String, body: List(Statement))
- branchStatement(cond: Expression, if: List(Statement), else: List(Statement))
- isBranchStatement(value: Expression, name: String, type: Expression, then: List(Statement), else: List(Statement))
- expressionStatement(Expression)

Miscellaneous functions:
- toFloat(Int): Float
- panic(String): Union[]
- printlnString(String): ()
- len(String): Int
- append(List(T), T): List(T)
- subString(String, Int, Int): String
- get(List(T), Int): T
- set(List(T), Int, T): ()

Reserved names: 'Box', primitives followed by 'Type' as suffix (e.g. ListType, TupleType), and the names of all functions named above in PascalCase.
```
