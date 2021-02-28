package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
)

var inline = `package generated

func unused() {
	var %s func(%s)
	{
		p := %s
		initial := *p

		escapePanicFlag := false
		%s = func(v %s) {
			if v != initial {
				*p = v

				if !escapePanicFlag {
					escapePanicFlag = true
					panic(&escapePanicFlag)
				}
			}
		}

		defer func() {
			if escapePanicFlag {
				escapePanicFlag = false

				r := recover()
				if r != &escapePanicFlag {
					// It would be better if we could unrecover.
					// Or not have to use panic/recover.
					panic(r)
				}
			}
		}()
	}
}
`

func find(fset *token.FileSet, info *types.Info, l []ast.Stmt) (index int, label, param, typ string) {
	index = -1

	for i, s := range l {
		assign, ok := s.(*ast.AssignStmt)
		if !ok {
			continue
		}

		if len(assign.Lhs) > 1 {
			continue
		}

		call, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			continue
		}

		ident, ok := call.Fun.(*ast.Ident)
		if !ok {
			continue
		}

		if ident.Name != "escape" {
			continue
		}

		ident, ok = assign.Lhs[0].(*ast.Ident)
		if !ok {
			continue
		}

		label = ident.Name

		if len(call.Args) != 1 {
			fmt.Fprintf(os.Stderr, "escape expects 1 argument, passed %d\n", len(call.Args))
			continue
		}

		t := info.Types[call.Args[0]].Type

		pt, ok := t.(*types.Pointer)
		if !ok {
			fmt.Fprintf(os.Stderr, "escape expects pointer type, passed %s\n", t)
			continue
		}

		typ = pt.Elem().String()

		buf := bytes.Buffer{}
		err := format.Node(&buf, fset, call.Args[0])
		if err != nil {
			continue
		}

		param = buf.String()

		index = i

		break
	}

	return
}

func replace(fset *token.FileSet, info *types.Info, list []ast.Stmt) []ast.Stmt {
	i, label, param, typ := find(fset, info, list)
	if i == -1 {
		return list
	}

	text := fmt.Sprintf(inline, label, typ, param, label, typ)
	g, err := parser.ParseFile(fset, "generated", text, parser.ParseComments)
	if err != nil {
		return list
	}

	l := make([]ast.Stmt, 0, len(list)+2)

	copy(l, list[:i])
	l = append(l, g.Decls[0].(*ast.FuncDecl).Body.List[:2]...)

	return append(l, replace(fset, info, list[i+1:])...)
}

func translate(name string) error {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
	if err != nil {
		return err //nolint:wrapcheck
	}

	conf := types.Config{Importer: importer.Default()}
	conf.Error = func(e error) {
		te, ok := e.(types.Error)
		if !ok || te.Msg != "undeclared name: escape" {
			err = e
		}
	}

	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}

	conf.Check(name, fset, []*ast.File{file}, info)

	if err != nil {
		return err //nolint:wrapcheck
	}

	_ = info

	ast.Inspect(file, func(n ast.Node) bool {
		b, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		b.List = replace(fset, info, b.List)

		return true
	})

	err = format.Node(os.Stdout, fset, file)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s FILE\n", os.Args[0])
		return
	}

	err := translate(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err.Error())
	}
}
