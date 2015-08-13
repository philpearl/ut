package main

import (
	"bytes"
	"fmt"
	"go/ast"
)

/*
We want to find any types used in interface methods that are defined in the
interface package, and so aren't qualified by a package name.

We may want to add a package name to these types.

We're looking for Fields within Fieldlists (either for parameters or returns)

A local interface will show up as a Field with a Type which is an Ident. The
Name of the Ident will be the interface name. The Obj of the Ident will be of
Kind type

A local struct is similar to a local interface

An imported interface shows up with a Type which is a SelectorExpr. An imported
struct is similar

A base type shows with a Type that is an Ident with no Obj
*/

func qualifyLocalTypes(n ast.Node, localPkgName string) bool {
	v := &QualifyLocalTypesVisitor{
		pkg: ast.NewIdent(localPkgName),
	}

	ast.Walk(v, n)
	return v.added
}

type QualifyLocalTypesVisitor struct {
	// This is the local package selector
	pkg   *ast.Ident
	added bool
}

func (q *QualifyLocalTypesVisitor) Visit(n ast.Node) ast.Visitor {
	// We're looking for fields within field lists within Params or Results of a FuncType
	switch n := n.(type) {
	case *ast.FuncType:
		to := &TypeObjVistor{
			q:         q,
			ancestors: []ast.Node{n},
		}

		return to
	}
	return q
}

type TypeObjVistor struct {
	q         *QualifyLocalTypesVisitor
	ancestors []ast.Node
}

func (to *TypeObjVistor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.Ident:
		if n.Obj != nil && n.Obj.Kind == ast.Typ {
			p := to.ancestors[len(to.ancestors)-1]
			switch p := p.(type) {
			case *ast.Field:
				p.Type = to.buildSelector(n)
			case *ast.StarExpr:
				p.X = to.buildSelector(n)
			case *ast.ArrayType:
				p.Elt = to.buildSelector(n)
			case *ast.MapType:
				if p.Key == n {
					p.Key = to.buildSelector(n)
				} else if p.Value == n {
					p.Value = to.buildSelector(n)
				}
			case *ast.ChanType:
				p.Value = to.buildSelector(n)
			default:
				fmt.Printf("Unexpected type %T\n", p)
				printNode(p)
			}
			return nil
		}
	}

	// We track ancestor nodes so we always know this node's immediate parent
	if n == nil {
		to.ancestors = to.ancestors[:len(to.ancestors)-1]
		// fmt.Printf("shrink to %d\n", len(to.ancestors))
	} else {
		to.ancestors = append(to.ancestors, n)
		// fmt.Printf("grow to %d %T\n", len(to.ancestors), n)
	}
	return to
}

func (to *TypeObjVistor) buildSelector(n *ast.Ident) *ast.SelectorExpr {
	to.q.added = true
	return &ast.SelectorExpr{
		X:   to.q.pkg,
		Sel: ast.NewIdent(n.Name),
	}
}

func printNode(n ast.Node) {
	var w bytes.Buffer
	ast.Fprint(&w, nil, n, nil)
	fmt.Printf(w.String())
}
