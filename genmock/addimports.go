package main

import (
	"go/ast"
	"go/token"
	"path"
	"strings"
)

// addImports is an AST Vistor that adds imports to the AST.
type addImports struct {
	imports []ast.Spec
}

func (v *addImports) Visit(n ast.Node) ast.Visitor {
	if n, ok := n.(*ast.GenDecl); ok && n.Tok == token.IMPORT {
		// Found our imports. Add new ones. But first we need to
		// eliminate duplicates
		type Imp struct {
			path string
			name string
		}
		found := map[Imp]struct{}{}
		specs := []ast.Spec{}

		addWithoutDuplicates := func(list []ast.Spec) {
			for _, s := range list {
				var imp Imp
				// Extract the path and name
				i := s.(*ast.ImportSpec)
				imp.path = i.Path.Value
				if i.Name != nil {
					// No point in adding a name if it is the same as the base of the path
					if i.Name.Name != path.Base(strings.Replace(imp.path, `"`, "", -1)) {
						imp.name = i.Name.Name
					}
				}

				// Have we seen this before
				_, ok := found[imp]
				if !ok {
					// No we haven't
					specs = append(specs, s)
					found[imp] = struct{}{}
				}
			}
		}

		addWithoutDuplicates(n.Specs)
		addWithoutDuplicates(v.imports)

		n.Specs = specs
		return nil
	}
	return v
}
