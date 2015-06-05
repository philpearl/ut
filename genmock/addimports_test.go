package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"
)

func TestAddImports(t *testing.T) {
	basefile := `
package fred

import (
    "io"
    mbase "github.com/philpearl/ut"
)
`

	tests := []struct {
		toAdd []ast.Spec
		exp   string
	}{
		{
			toAdd: []ast.Spec{buildImport("io", "")},
			exp: `package fred

import (
	"io"
	mbase "github.com/philpearl/ut"
)
`,
		},
		{
			toAdd: []ast.Spec{buildImport("fred.com/elephant", "tusk")},
			exp: `package fred

import (
	"io"
	mbase "github.com/philpearl/ut"
	tusk "fred.com/elephant"
)
`,
		},
		{
			toAdd: []ast.Spec{buildImport("fred.com/elephant", "")},
			exp: `package fred

import (
	"io"
	mbase "github.com/philpearl/ut"
	"fred.com/elephant"
)
`,
		},
		{
			toAdd: []ast.Spec{buildImport("fred.com/elephant", ""), buildImport("io", ""), buildImport("fmt", ""), buildImport("fmt", "")},
			exp: `package fred

import (
	"io"
	mbase "github.com/philpearl/ut"
	"fred.com/elephant"
	"fmt"
)
`,
		},
	}

	for i, test := range tests {
		// get a file to add imports to
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "dummy.go", basefile, 0)
		if err != nil {
			t.Fatalf("Failed to set up base ast. %v", err)
		}

		v := &addImports{test.toAdd}
		ast.Walk(v, f)

		var buf bytes.Buffer
		printer.Fprint(&buf, fset, f)

		if buf.String() != test.exp {
			t.Fatalf("Test %d output not as expected.  Have %s", i, buf.String())
		}
	}
}

func buildImport(path, name string) *ast.ImportSpec {
	is := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"" + path + "\"",
		},
	}

	if name != "" {
		is.Name = ast.NewIdent(name)
	}
	return is
}
