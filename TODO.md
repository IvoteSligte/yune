broken parsing for function calls without parentheses
cycle detection
HYGIENE
macro error accountability checking
spans in macro-generated code

tuple is (x: Int, y: Int) -> doStuff(x, y)

remove trailing (unindented) empty lines from macro and the trailing newline that is always there (I think)


C++23 has constexpr std::string string literals and some support for constexpr std::variant

C++ runtime provides an allocator, so `new` cannot be used?

---

Removed `ty::` namespace which was preventing exporting of symbols, instead using the conventional `_t` suffix for types, `_f` for functions, and `_` for other builtins. The underscore is only allowed in all-uppercase Yune symbols and therefore prevents conflicts.

---

only execute globals initialized with pure functions at compile-time

A C++ shared library provides a `dynamic loader` that initialises C++ global variables and sets up the standard allocator.

compiling as shared library with the C++ standard library linked statically so that C++ is automatically initialized even from C
    now only requires dynamically linked libc, which is reasonable since Yune interoperates with other languages via the C interface

not sure how to make C++ symbols C-accessible, obviously `extern "C"`, but I am currently using objects for functions so that `serialize()` is attached
    solution: get rid of objects at runtime?
    still the issue with namespaces, particularly for functions that take `ty::String`, `ty::List`, and other `ty::*`



    
