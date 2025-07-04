# Yune
Compiler for the Yune programming language

The Yune programming language promotes usage of Domain Specific Languages through the use of macros rather than trying to be a single monolithic language that has every feature imaginable. Yune compiles to C and has memory safety based on Rust's. All data in Yune itself is abstract by default, which allows for low-cost interactions with other languages and interfaces for free.

Yune uses the ".un" (pronounced "dot yune") file extension.

## WIP
Very Work-In-Progress. The language's syntax itself is not yet stable, but it is converging. The parser is the current focus of development.

## Compiler Stages

- [x] Lexer
    <details>
    Tokenize a file into a sequence of tokens
    </details>
- [ ] Parser
    <details>
    Construct a Parse Tree from a lexed file.
    The Parse Tree intuitively contains everything needed to output a formatted version of the same code.
    </details>
- [ ] HIR
    <details>
    - [ ] Convert string identifiers to abstract integer identifiers.
    - [ ] Type checking.
    </details>
- [ ] THIR
    <details>
    Fully typed HIR.
    </details>
- [ ] MIR
    <details>
    - [ ] Lower multi-RHS binary expressions.
    - [ ] Lambda lifting.
    - [ ] Tagged unions to concrete C-types.
    - [ ] Explicit Typing
    </details>
- [ ] C
    <details>
    Convert into C code.
    </details>

## Planned Features

- Operator overloading

- Floating point values

- Type narrowing (for unions)
    <details>
    Determines the narrowed type of a variable after a branch has been applied to it.
    Example:
        ```
        square(n: Int | Float): Int | Float =
            n is int: Int -> n * n
            // n is now known to be a Float due to type narrowing
            n * n
        ```
    Considerations:
        While it is easy to implement for unions, more complicated operations can also benefit from type narrowing for better static analysis. This requires subtyping, though.
    </details>

- Type inference
    <details>
    Type inference is the process of determining a variable's type from the context, allowing the user to omit the type in a lot of cases.
    </details>

- Borrow checking

- Language Server
    <details>
    A language server provides syntax highlighting and in-code-editor error messages, as well as other less useful features.
    </details>

## Possible Features

These are mere ideas and may not be implemented in future versions. Even if they are, it will take a long time because the compiler itself is a lot of work.

### Type specifications on non-variables
Currently only variables' types can be specified, but this is quite limited.
Example: `println (randomVariable: Int)`

### Macros
A macro takes a raw string representation of a block of code in any language and parses it into Yune code. This allows embedding any [Domain Specific Language](https://wikipedia.com/wiki/domain_specific_language) (DSL). Yune does not try to be a language that can do everything. Without massive financial backing that is impossible regardless. Furthermore, many incredible languages with specialized purposes have already been developed. It is not necessary to reinvent the wheel. Integration with other languages - which is made easier by Yune's nature supporting abstract types and compilation to C - allows users to use pre-existing tools. Indentation-sensitivity ensures that any language can be embedded because there is no need to indicate the end of the macro by anything but a de-indentation. The macro receives a string without indentation regardless of the indentation of the macro.

#### Example
```
let filename = "long-file.txt"
// Executes an inline bash script that grabs the name of a file
// from the Yune `filename` variable, loads the file, applies
// a substitution to strings in the file, and then returns the
// resulting character sequence.
let modifiedFile = Fn'call bash#
    echo "Hello from bash!"
    cat $filename | sed 's/Hello, Mars!/Hello, World!/g'

println modifiedFile
```

#### Considerations
    Macros need access to variables and constants defined in its context. What macros can interact with, how these interactions work and what values macros can produce needs to be determined.
    It is difficult to have proper LSP<->macro interactions such as variable renaming. If macros are allowed to implicitly borrow and use variables from the code context then it can be more complicated than a simple string renaming task because variables can have different names in the embedded language. Syntax highlighting at the very least should be supported for readability.
    Domain Specific Languages can need temporary files for compilation, so figuring out an API for those should help with macro simplicity.
    Macros are run by the Yune compiler, which means that they have the same permissions as it. Sandboxing to prevent filesystem modification might be necessary. Particularly if the LSP needs to run macro code (e.g. for diagnostics within a DSL).

### Subtypes
A subtype is a type that is defined by another type and a filter function on values of the other type. In combination with advanced type narrowing, subtypes can provide a way of checking whether a value has certain properties at compile time.

#### Example 1:
```
Even = sub Int, |n| n.mod(2) == 0
```

#### Example 2:
```
Alphabetic = sub String, |s| s.chars.all Char'isAlpha
Positive = sub String, |s| s.chars.all Char'isDigit
```

#### Considerations

Certain functions need to be able to tell if two types intersect. For example, the union operator `A | B` requires that `A` and `B` are disjoint.

This is not viable for arbitrary functions provided to subtypes. Determining if the following subtypes are disjoint is impossible.

```
Random1 = sub Int, |_| remote.getRandomBool
Random2 = sub Int, |_| ;remote.getRandomBool
```

A simple solution is moving the burden to the programmer and providing an attribute `@disjoint`, but a good static analysis process is still required, otherwise it defeats the point of having subtypes in the first place. Which is better compiler understanding of the code (and as a consequence to have some unions without storage overhead).

There is an even worse issue, though, and that is that membership tests are inconsistent. To prevent this, [referential transparency](https://wikipedia.com/wiki/referential_transparency) is required, which in turn requires a compiler step that can determine if a function is referentially transparent. Though just verifying that a function does not modify its arguments might be enough if the assumption is that the user knows what they are doing. An unlikely assumption, but perhaps sensible since Yune is not a language for writing mathematical proofs.

