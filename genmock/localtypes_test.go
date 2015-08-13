package main

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"testing"
)

func TestLocalTypes(t *testing.T) {
	tests := []struct {
		code  string
		exp   string
		added bool
	}{
		{
			code: `
package blah

type L1 struct {}

type I1 interface {
	f1(p L1) L1
}
`,
			exp: `package blah

type L1 struct{}

type I1 interface {
	f1(p llmock.L1,) llmock.L1
}
`,
			added: true,
		},
		{
			code: `
package blah

type L1 int

type I1 interface {
	f1(p L1) (helen L1, brian L1)
}
`,
			exp: `package blah

type L1 int

type I1 interface {
	f1(p llmock.L1,) (helen llmock.L1, brian llmock.L1,)
}
`,
			added: true,
		},
		{
			code: `
package blah

type I1 interface {
	f1(p int) (helen, brian int)
}
`,
			exp: `package blah

type I1 interface {
	f1(p int) (helen, brian int)
}
`,
			added: false,
		},
		{
			code: `
package blah

type L1 struct {}

type I1 interface {
	f1(p, q L1) L1
}
`,
			exp: `package blah

type L1 struct{}

type I1 interface {
	f1(p, q llmock.L1,) llmock.L1
}
`,
			added: true,
		},
		{
			code: `
package blah

type L1 struct {}

type I1 interface {
	f1(p *L1) *L1
}
`,
			exp: `package blah

type L1 struct{}

type I1 interface {
	f1(p *llmock.L1,) *llmock.L1
}
`,
			added: true,
		},
		{
			code: `
package blah

type L1 struct {}

type I1 interface {
	f1(p []L1) map[L1]L1
}
`,
			exp: `package blah

type L1 struct{}

type I1 interface {
	f1(p []llmock.L1,) map[llmock.L1]llmock.L1
}
`,
			added: true,
		},
		{
			code: `
package blah

type L1 struct {}

type I1 interface {
	f1(p []L1) chan L1
}
`,
			exp: `package blah

type L1 struct{}

type I1 interface {
	f1(p []llmock.L1,) chan llmock.L1
}
`,
			added: true,
		},
	}

	for i, test := range tests {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "dummy.go", test.code, 0)
		if err != nil {
			t.Fatalf("Test %d, failed to parse code. %v", i, err)
		}

		added := qualifyLocalTypes(file, "llmock")

		w := bytes.Buffer{}
		err = format.Node(&w, fset, file)

		if w.String() != test.exp {
			t.Fatalf("Test %d result not as expected. Have `%s` expected `%s`", i, w.String(), test.exp)
		}

		if added != test.added {
			t.Fatalf("Test %d, Added not as expected ", i)
		}
	}

}
