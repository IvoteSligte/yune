# Yune

## Dependencies

go
clang++
clang-format
C++23
clang-repl

## Linking a Yune library

The Yune compiler creates an object file named `library.o` for code without a `main` function. In order to use this from languages that do not natively support C++ interoperation, the C++ standard library must be linked using the `-lstdc++` linker flag.
