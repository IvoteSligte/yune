# Yune

## Dependencies

go
clang++
clang-format
C++23
clang-repl

## Yune libraries

If a Yune source file does not contain a `main` function, then a file named `library.hpp` is produced, which can be included in any C++ project. The compiler file `cpp/pb.hpp` should be included with `-I<path_to_pb.hpp_folder>` to ensure the definitions which `library.hpp` relies on exist. The standard library should be set to at least version C++23 or GNU++23 (flag `-std=c++23` or `-std=gnu++23`).

In the case that the target project is not a C++ project, it is recommended to compile the `library.hpp` file to a dynamic library file with a static C++ standard library, which can easily be linked against from any language. `extern "C"` wrappers need to be made for the required functions.


