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

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

var inline = `package generated

func unused() {
	var %s func(%s)
	{
		p := %s
		initial := *p

		panicking := false

		// Generated by escape.
		%s = func(v %s) {
			if v != initial {
				*p = v

				if !panicking {
					panicking = true
					panic(&panicking)
				}
			}
		}

		defer func() {
			if panicking {
				panicking = false

				r := recover()
				if r != &panicking {
					// It would be better if we could unrecover.
					// Or not have to use panic/recover.
					panic(r)
				}
			}
		}()
	}
}
`

func find(d *decorator.Decorator, fset *token.FileSet, info *types.Info, l []dst.Stmt) (index int, label, param, typ string) {
	index = -1

	for i, s := range l {
		assign, ok := s.(*dst.AssignStmt)
		if !ok {
			continue
		}

		if len(assign.Lhs) > 1 {
			continue
		}

		call, ok := assign.Rhs[0].(*dst.CallExpr)
		if !ok {
			continue
		}

		ident, ok := call.Fun.(*dst.Ident)
		if !ok {
			continue
		}

		if ident.Name != "escape" {
			continue
		}

		ident, ok = assign.Lhs[0].(*dst.Ident)
		if !ok {
			continue
		}

		label = ident.Name

		if len(call.Args) != 1 {
			fmt.Fprintf(os.Stderr, "escape expects 1 argument, passed %d\n", len(call.Args))
			continue
		}

		ae := d.Map.Ast.Nodes[call.Args[0]].(ast.Expr)

		t := info.Types[ae].Type

		pt, ok := t.(*types.Pointer)
		if !ok {
			fmt.Fprintf(os.Stderr, "escape expects pointer type, passed %s\n", t)
			continue
		}

		typ = pt.Elem().String()

		buf := bytes.Buffer{}
		err := format.Node(&buf, fset, ae)
		if err != nil {
			continue
		}

		param = buf.String()

		index = i

		break
	}

	return
}

func replace(d *decorator.Decorator, fset *token.FileSet, file *dst.File, info *types.Info, list []dst.Stmt) []dst.Stmt {
	i, label, param, typ := find(d, fset, info, list)
	if i == -1 {
		return list
	}

	text := fmt.Sprintf(inline, label, typ, param, label, typ)
	g, err := parser.ParseFile(fset, "", text, parser.ParseComments)
	if err != nil {
		return list
	}

	generated, err := d.DecorateFile(g)
	if err != nil {
		return list
	}

	adding := generated.Decls[0].(*dst.FuncDecl).Body.List[:2]

	adding[0] = dst.Clone(adding[0]).(dst.Stmt)
	adding[1] = dst.Clone(adding[1]).(dst.Stmt)

	l := make([]dst.Stmt, 0, len(list)+2)

	copy(l, list[:i])
	l = append(l, adding...)

	return append(l, replace(d, fset, file, info, list[i+1:])...)
}

func translate(name string) error {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
	if err != nil {
		return err //nolint:wrapcheck
	}

	d := decorator.NewDecorator(fset)

	file, err := d.DecorateFile(f)
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

	conf.Check(name, fset, []*ast.File{f}, info)

	if err != nil {
		return err //nolint:wrapcheck
	}

	_ = info

	dst.Inspect(file, func(n dst.Node) bool {
		b, ok := n.(*dst.BlockStmt)
		if !ok {
			return true
		}

		b.List = replace(d, fset, file, info, b.List)

		return true
	})

	err = decorator.Print(file)
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
