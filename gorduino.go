package main

import "go/ast"
import "go/parser"
import "go/token"
import "reflect"
import "fmt"
import "flag"
import "strings"

type NodeChannel chan ast.Node

func (me NodeChannel) astAssignStmt(n *ast.AssignStmt) string {
    x := <-me
    ident := me.astIdent(x.(*ast.Ident))
    x = <-me
    expr := me.astExpr(x.(ast.Expr))
    me.eatNil()
    return fmt.Sprintf("%s = %s;", ident, expr)
}

func (me NodeChannel) astBasicLit(n *ast.BasicLit) string {
    me.eatNil()
    return n.Value
}

func (me NodeChannel) astBinaryExpr(n *ast.BinaryExpr) string {
    x  := me.astExpr((<-me).(ast.Expr))
    y  := me.astExpr((<-me).(ast.Expr))
    op := n.Op.String()

    me.eatNil()
    return fmt.Sprintf("%s %s %s", x, op, y)
}

func (me NodeChannel) astBlockStmt(n *ast.BlockStmt) string {
    stmts := make([]string, 0)

    x := <-me
    for x != nil {
        stmts = append(stmts, me.astStmt(x.(ast.Stmt)))
        x = <-me
    }
    return fmt.Sprintf("{\n%s\n}", strings.Join(stmts, "\n"))
}

func (me NodeChannel) astCallExpr(n *ast.CallExpr) string {
    x := <-me
    sub := me.astIdent(x.(*ast.Ident))

    args := make([]string, 0)
    x = <-me
    for x != nil {
        args = append(args, me.astExpr(x.(ast.Expr)))
        x = <-me
    }

    return fmt.Sprintf("%s(%s)", sub, strings.Join(args, ", "))
}

func (me NodeChannel) astConstSpec(n *ast.ValueSpec) string {
    name := me.astIdent((<-me).(*ast.Ident))
    val  := me.astBasicLit((<-me).(*ast.BasicLit))
    me.eatNil()
    return fmt.Sprintf("#define %s %s", name, val)
}

func (me NodeChannel) astDeclStmt(n *ast.DeclStmt) (ret string) {
    x := <-me
    switch x.(type) {
    case *ast.GenDecl:
        ret = me.astGenDecl(x.(*ast.GenDecl))
    case *ast.FuncDecl:
        ret = me.astFuncDecl(x.(*ast.FuncDecl))
    default:
        panic(fmt.Sprint("Unknown declaration: ", reflect.TypeOf(x)))
    }
    me.eatNil()
    return
}

func (me NodeChannel) astExpr(n ast.Expr) (ret string) {
    switch n.(type) {
    case *ast.CallExpr:
        ret = me.astCallExpr(n.(*ast.CallExpr))
    case *ast.BasicLit:
        ret = me.astBasicLit(n.(*ast.BasicLit))
    case *ast.Ident:
        ret = me.astIdent(n.(*ast.Ident))
    case *ast.BinaryExpr:
        ret = me.astBinaryExpr(n.(*ast.BinaryExpr))
    case *ast.UnaryExpr:
        ret = me.astUnaryExpr(n.(*ast.UnaryExpr))
    default:
        panic(fmt.Sprint("Unknown expression: ", reflect.TypeOf(n)))
    }
    return
}

func (me NodeChannel) astExprStmt(n *ast.ExprStmt) string {
    expr := me.astExpr((<-me).(ast.Expr))
    me.eatNil()
    return fmt.Sprint(expr, ";")
}

func (me NodeChannel) astFile(n *ast.File) string {
    x := <-me
    ident := me.astIdent(x.(*ast.Ident))

    decls := make([]string, 1)
    decls = append(decls, fmt.Sprint("// package ", ident))

    x = <-me
    for x != nil {
        switch x.(type) {
        case *ast.GenDecl:
            decls = append(decls, me.astGenDecl(x.(*ast.GenDecl)))
        case *ast.FuncDecl:
            decls = append(decls, me.astFuncDecl(x.(*ast.FuncDecl)))
        default:
            panic(fmt.Sprint("Unknown declaration type: ",
                             reflect.TypeOf(x)))
        }
        x = <-me
    }

    return strings.Join(decls, "\n")
}

func (me NodeChannel) astField(n *ast.Field) string {
    x := <-me
    name := me.astIdent(x.(*ast.Ident))

    x = <-me
    if x == nil {
        me.eatNil()
        return name
    }
    typ := me.astIdent(x.(*ast.Ident))

    me.eatNil()

    return fmt.Sprintf("%s %s", typ, name)
}

func (me NodeChannel) astFieldList(n *ast.FieldList) []string {
    fields := make([]string, 0)
    x := <-me
    for x != nil {
        fields = append(fields, me.astField(x.(*ast.Field)))
        x = <-me
    }

    return fields
}

func (me NodeChannel) astForStmt(n *ast.ForStmt) string {
    args := make([]string, 0)
    init, cond, iter := "", "", ""
    var block string

    x := <-me
    for x != nil {
        switch x.(type) {
        case ast.Expr:
            args = append(args, me.astExpr(x.(ast.Expr)))
        case *ast.ExprStmt:
            stmt := me.astExprStmt(x.(*ast.ExprStmt))
            stmt = stmt[0:len(stmt)-1] // remove trailing ;
            args = append(args, stmt)
        case *ast.BlockStmt:
            block = me.astBlockStmt(x.(*ast.BlockStmt))
        case ast.Stmt:
            stmt := me.astStmt(x.(ast.Stmt))
            stmt = stmt[0:len(stmt)-1] // remove trailing ;
            args = append(args, stmt)
        default:
            panic(fmt.Sprint("Unknown for argument: ", reflect.TypeOf(x)))
        }
        x = <-me
    }

    if len(args) == 0 {
        // nothing to do
    } else if len(args) == 1 {
        cond = args[0]
    } else if len(args) == 3 {
        init, cond, iter = args[0], args[1], args[2]
    } else {
        panic(fmt.Sprintf("Don't know how to handle %d-arg for loop",
                          len(args)))
    }

    return fmt.Sprintf("for (%s; %s; %s;) %s", init, cond, iter, block)
}

func (me NodeChannel) astFuncDecl(n *ast.FuncDecl) string {
    x := <-me
    name := me.astIdent(x.(*ast.Ident))

    x = <-me
    rettype, params := me.astFuncType(x.(*ast.FuncType))

    x = <-me
    block := me.astBlockStmt(x.(*ast.BlockStmt))

    me.eatNil()

    return fmt.Sprintf("%s %s%s\n%s", rettype, name, params, block)
}

func (me NodeChannel) astFuncType(n *ast.FuncType) (rettype string,
                                                    signature string) {
    x := <-me
    params := me.astFieldList(x.(*ast.FieldList))

    rettype = "void"
    x = <-me
    if x != nil {
        returns := me.astFieldList(x.(*ast.FieldList))
        if len(returns) > 1 {
            panic("Functions can only return one value")
        }
        rettype = returns[0]
    }

    signature = fmt.Sprintf("(%s)", strings.Join(params, ", "))

    return
}

func (me NodeChannel) astGenDecl(n *ast.GenDecl) (ret string) {
    if n.Tok == token.IMPORT {
        imp := <-me
        imports := make([]string, 1)
        for imp != nil {
            imports = append(imports,
                             me.astImportSpec(imp.(*ast.ImportSpec)))
            imp = <-me
        }
        ret = strings.Join(imports, "\n")
    } else if n.Tok == token.CONST {
        con := <-me
        consts := make([]string, 1)
        for con != nil {
            consts = append(consts,
                            me.astConstSpec(con.(*ast.ValueSpec)))
            con = <-me
        }
        ret = strings.Join(consts, "\n")
    } else if n.Tok == token.VAR {
        vr := <-me
        vars := make([]string, 1)
        for vr != nil {
            vars = append(vars, me.astValueSpec(vr.(*ast.ValueSpec)))
            vr = <-me
        }
        ret = strings.Join(vars, "\n")
    } else {
        panic(fmt.Sprint("Unsupported GenDecl: ", n.Tok))
    }
    return
}

func (me NodeChannel) astIfStmt(n *ast.IfStmt) string {
    x := <-me
    cond := me.astExpr(x.(ast.Expr))

    x = <-me
    block := me.astBlockStmt(x.(*ast.BlockStmt))

    elses := make([]string, 0)

    x = <-me
    for x != nil {
        switch x.(type) {
        case *ast.IfStmt:
            elses = append(elses, me.astIfStmt(x.(*ast.IfStmt)))
        case *ast.BlockStmt:
            elses = append(elses, me.astBlockStmt(x.(*ast.BlockStmt)))
            break
        default:
            panic("wtf, how could this happen?")
        }
        x = <-me
    }

    var elses_str string
    if len(elses) > 0 {
        elses_str = fmt.Sprintf("else %s",strings.Join(elses, " else "))
    }

    return fmt.Sprintf("if (%s) %s %s", cond, block, elses_str)
}

func (me NodeChannel) astIdent(n *ast.Ident) string {
    me.eatNil()
    return n.Name
}

func (me NodeChannel) astImportSpec(n *ast.ImportSpec) string {
    x    := <-me
    path := me.astBasicLit(x.(*ast.BasicLit))
    me.eatNil()
    path = path[1:len(path)-1] // remove sorrounding ""'s
    return fmt.Sprintf("#include <%s>", path)
}

func (me NodeChannel) astIncDecStmt(n *ast.IncDecStmt) string {
    x  := me.astExpr((<-me).(ast.Expr))
    op := "++"
    if n.Tok == token.DEC {
        op = "--"
    }

    me.eatNil()
    return fmt.Sprintf("%s%s;", x, op)
}

func (me NodeChannel) astReturnStmt(n *ast.ReturnStmt) string {
    x := <-me
    lit := me.astBasicLit(x.(*ast.BasicLit))
    me.eatNil()
    return fmt.Sprintf("return %s;", lit)
}

func (me NodeChannel) astStmt(n ast.Stmt) (ret string) {
    switch n.(type) {
    case *ast.ExprStmt:
        ret = me.astExprStmt(n.(*ast.ExprStmt))
    case *ast.ReturnStmt:
        ret = me.astReturnStmt(n.(*ast.ReturnStmt))
    case *ast.ForStmt:
        ret = me.astForStmt(n.(*ast.ForStmt))
    case *ast.IfStmt:
        ret = me.astIfStmt(n.(*ast.IfStmt))
    case *ast.DeclStmt:
        ret = me.astDeclStmt(n.(*ast.DeclStmt))
    case *ast.AssignStmt:
        ret = me.astAssignStmt(n.(*ast.AssignStmt))
    case *ast.IncDecStmt:
        ret = me.astIncDecStmt(n.(*ast.IncDecStmt))
    default:
        panic(fmt.Sprint("Unsupported statement: ", reflect.TypeOf(n)))
    }
    return
}

func (me NodeChannel) astUnaryExpr(n *ast.UnaryExpr) string {
    x  := me.astExpr((<-me).(ast.Expr))
    op := n.Op.String()

    me.eatNil()
    return fmt.Sprintf("%s%s", op, x)
}

func (me NodeChannel) astValueSpec(n *ast.ValueSpec) string {
    if n.Type == nil {
        panic("Type inference NYI")
    }
    if len(n.Names) > 1 {
        panic("Multiple assignment unsupported")
    }
    x := <-me
    name := me.astIdent(x.(*ast.Ident))
    x = <-me
    typ := me.astIdent(x.(*ast.Ident))
    x = <-me
    var val string
    if x != nil {
        val = fmt.Sprint(" = ", me.astExpr(x.(ast.Expr)))
        me.eatNil()
    }
    return fmt.Sprintf("%s %s%s;", typ, name, val)
}

func (me NodeChannel) eatNil() {
    if x := <-me; x != nil {
        panic(fmt.Sprint("Nil expected, got ", reflect.TypeOf(x)))
    }
}

func (me NodeChannel) Parse(out chan string) {
    for n := range me {
        switch n.(type) {
        case *ast.File:
            out <- me.astFile(n.(*ast.File))
        default:
            panic(fmt.Sprint("Unknown node: ", reflect.TypeOf(n)))
        }
    }
    close(out)
}

func (me NodeChannel) Visit(n ast.Node) ast.Visitor {
    me <- n
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

    v := make(NodeChannel)

    code := make(chan string)
    go v.Parse(code)

    ast.Walk(v, file)

    str := <-code
    fmt.Printf("// Code generated from %s\n", filename)
    fmt.Println(str)
}
