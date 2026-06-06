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
