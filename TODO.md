
NOTE: is clang-repl's final result being properly waited for when there is no `main`?


runtime-free serialization of Box and List (validity can be checked with `constinit`)
broken parsing for function calls without parentheses
cycle detection
HYGIENE
macro error accountability checking

tuple is (x: Int, y: Int) -> doStuff(x, y)

remove trailing (unindented) empty lines from macro and the trailing newline that is always there (I think)


C++23 has constexpr std::string string literals and some support for constexpr std::variant

C++ runtime provides an allocator, so `new` cannot be used?

---

Removed `ty::` namespace which was preventing exporting of symbols, instead using the conventional `_t` suffix for types, `_f` for functions, and `_` for other builtins. The underscore is not allowed in a Yune symbol and therefore prevents conflicts.

---

only execute globals initialized with pure functions at compile-time

compiling as shared library so that C++ is automatically initialized even from C:
    `clang++ -shared -fPIC mylib.cpp -o libmylib.so`

not sure how to make C++ symbols C-accessible, obviously `extern "C"`, but I am currently using objects for functions so that `serialize()` is attached
    solution: get rid of objects at runtime?
    still the issue with namespaces, particularly for functions that take `ty::String`, `ty::List`, and other `ty::*`

---

A C++ shared library provides a `dynamic loader` that initialises C++ global variables and sets up the standard allocator.

The alternative is statically linking the C++ standard library and providing an initialization function.

A better alternative is simply compiling to C and using only static initialization. Except at compile-time, when C++ libraries can freely be used.

Steps to runtime-free glory [make sure to describe this in the thesis]:

[x] disable exceptions
[x] use evaluated value for runtime globals
[x] no std::vector, std::string, std::shared_ptr at global scope; implemented with wrappers that allow statics
[ ] only static initializers:
   ```C++
   constinit auto* p_ptr = [] {
       static constexpr double p = 10.0;
       return &p;
   }();
   ```

Note that List and String are immutable because they may reference immutable static global data.
