
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
