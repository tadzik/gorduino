# Gorduino: Go to C compiler #

Gorduino takes Go code and produces C code. It was meant to be used for
arduino development, but can be used for anything else as well.

Gorduino is written in Go, and uses the go/parser library for parsing
Go code. The semantics of Go code are not checked: it's possible to write
illegal Go code which will be translated to legal (or not) C.

There's no type checking of any sort. Types are not converted in any way,
which may result in incorrect (or illegal) code.

The resulting C code may or may not work. Don't expect miracles.

## Features ##

 * import   (becomes #include)
 * const    (becomes #define)
 * variabes (no type inference)
 * if, else if, else
 * for (all forms)
 * function declarations
 * assignments
 * function calls
 * other things that may accidentally work

### Missing features ###

 * Arrays and structs
 * Type inference

Most of Go features (multiple assignments, go statement, channels etc)
will probably never get implemted, either because they're impossible to
implement, or will be wrong more often than right, or just make no sense
in arduino scope, hence I'll not be interested in implementing them.

# Usage #

    make
    make test # requires Perl

    ./gorduino    file.go (code goes to STDOUT)
    ./treeprinter file.go (pretty-prints the AST (useful for debugging))
