package main

import "go/ast"
import "go/parser"
import "go/token"
import "reflect"
import "fmt"
import "flag"

type visitor struct {
    indent int
}

func (me visitor) Visit(n ast.Node) ast.Visitor {
    for i := 0; i < me.indent; i++ {
        fmt.Print("    ")
    }
    fmt.Println(reflect.TypeOf(n))
    if n != nil {
        me.indent++
    } else {
        me.indent--
    }
    return me
}

func main() {
    flag.Parse()
    if flag.NArg() != 1 {
        fmt.Println("Usage: gorduino <filename>")
        return
    }
    filename := flag.Arg(0)

    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, filename, nil, 0)
    if err != nil {
        fmt.Println("Error: ", err)
        return
    }

    var v visitor

    ast.Walk(v, file)
}
