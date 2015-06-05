package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// parseCodeBlock() parses a block of code.
//
// It works by embedding the code in a dummy go file with a function, then
// extracting the AST for the code block we're interested in
func parseCodeBlock(code string) ([]ast.Stmt, error) {
	// Embed the code in a function and add a package
	code = "package dummy\nfunc dummy() {\n" + code + "\n}\n"

	// Parse the code
	fset := token.NewFileSet()
	af, err := parser.ParseFile(fset, "dummy.go", code, 0)
	if err != nil {
		return nil, err
	}

	// Extract the statements we need
	v := &blockVisitor{}
	ast.Walk(v, af)

	return v.stmts, nil
}
