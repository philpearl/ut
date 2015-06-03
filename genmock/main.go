package main

// TODO:::
// 1 Need some imports for parameters and returns used in the mocks
// 2 A NewMockXXXX method is handy
// 2 Some routines to allow chaining of AddCall and SetReturns

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
)

// blockVisitor walks the AST and extracts the first Block Statement it finds.
// We only use it when we've generated the code ourselves so we know there is only
// one code block to look for
type blockVisitor struct {
	stmts []ast.Stmt
}

func (v *blockVisitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.BlockStmt:
		v.stmts = n.List
		return nil
	}
	return v
}

type findUsedImports struct {
	names map[string]struct{}
}

func newFindUsedImports() *findUsedImports {
	return &findUsedImports{make(map[string]struct{})}
}

func (v *findUsedImports) Visit(n ast.Node) ast.Visitor {
	sel, ok := n.(*ast.SelectorExpr)
	if ok {
		id, ok := sel.X.(*ast.Ident)
		if ok {
			v.names[id.Name] = struct{}{}
		}
	}
	return v
}

// isUsed indicates whether an import is used.
//
// Import specs can either just be a path, in which case the last
// path component is the name, so it can also have a separate name
func (v *findUsedImports) isUsed(s *ast.ImportSpec) bool {
	if s.Name != nil {
		_, ok := v.names[s.Name.Name]
		return ok
	}

	path := s.Path.Value
	if path[0] == '"' {
		path = path[1:]
	}
	if path[len(path)-1] == '"' {
		path = path[:len(path)-1]
	}
	parts := strings.Split(path, "/")

	name := parts[len(parts)-1]
	_, ok := v.names[name]
	return ok
}

// InterfaceVisitor walks the AST and finds interfaces
type InterfaceVisitor struct {
	fset        *token.FileSet
	name        string
	code        string
	packageName string
	imports     []*ast.ImportSpec
}

func (i *InterfaceVisitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.TypeSpec:
		t, ok := n.Type.(*ast.InterfaceType)
		if ok {
			// This is an interface
			if n.Name.Name != i.name {
				return nil
			}
			i.code = i.buildMockForInterface(t)
			return nil
		}
	case *ast.ImportSpec:
		i.imports = append(i.imports, n)
	}

	return i
}

func (i *InterfaceVisitor) buildMockForInterface(t *ast.InterfaceType) string {
	var buf bytes.Buffer

	fi := newFindUsedImports()

	nodes := []ast.Decl{}

	// Mock Implementation of the interface
	m := buildMockStruct(i.name)
	ast.Walk(fi, m)
	nodes = append(nodes, m)

	// Method receiver for our mock interface
	recv := buildMethodReceiver(i.name)

	for _, m := range t.Methods.List {
		t, ok := m.Type.(*ast.FuncType)
		if ok {
			// Names for return values causes problems, so remove them.
			if t.Results != nil {
				removeFieldNames(t.Results)
			}

			// We can have multiple names for a method type if multiple
			// methods are declared with the same signature
			for _, n := range m.Names {
				fd := buildMockMethod(recv, n.Name, t)
				ast.Walk(fi, fd)
				nodes = append(nodes, fd)
			}
		}
	}

	usedImports := []ast.Spec{}
	utSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"github.com/philpearl/ut\"",
		},
	}

	for _, is := range i.imports {
		if fi.isUsed(is) {
			usedImports = append(usedImports, is)
		}
	}

	usedImports = append(usedImports, utSpec)

	fmt.Printf("%d used imports", len(usedImports))

	imp := &ast.GenDecl{
		Tok:    token.IMPORT,
		Specs:  usedImports,
		Lparen: token.Pos(1), // This is a hack to get all the imports to appear
	}

	nodes = append([]ast.Decl{imp}, nodes...)

	f := &ast.File{
		Name:  ast.NewIdent(i.packageName),
		Decls: nodes,
	}

	ast.SortImports(i.fset, f)

	printer.Fprint(&buf, i.fset, f)

	return buf.String()
}

// removeFieldNames removes names from the FieldList in place.
// This is used to remove names from return values
func removeFieldNames(fl *ast.FieldList) {
	l := []*ast.Field{}
	for _, f := range fl.List {
		if f.Names == nil {
			l = append(l, f)
		} else {
			for range f.Names {
				nf := *f
				nf.Names = nil
				l = append(l, &nf)
			}
		}
	}
	fl.List = l
}

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

func buildMockStruct(name string) *ast.GenDecl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(fmt.Sprintf("Mock%s", name)),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.SelectorExpr{
									X:   ast.NewIdent("ut"),
									Sel: ast.NewIdent("CallTracker"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildMethodReceiver(name string) *ast.FieldList {
	return &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{
					ast.NewIdent("i"),
				},
				Type: &ast.StarExpr{
					X: ast.NewIdent(fmt.Sprintf("Mock%s", name)),
				},
			},
		},
	}
}

/* buildMockMethod builds the AST for the mock method.
The function body needs to look something like:

	r := ut.TrackCall("method", param1, param2)
	return r[0].(int), r[1].(thing)

... except we need to worry about types a little more for the
return values.  So instead we do

	r := ut.TrackCall("method", param1, param2)
	var r_0 int
	var r_1 thing
	if r[0] != nil { r_0 = r[0].(int) }
	if r[1] != nil { r_1 = r[1].(thing) }
	return r_0, r_1

... and we might have an ellipsis parameter so in fact we do

	params := make([]interface{}, 2)
	params[0] = param1
	params[1] = param2
	r := ut.TrackCall("method", params...)
	var r_0 int
	var r_1 thing
	if r[0] != nil { r_0 = r[0].(int) }
	if r[1] != nil { r_1 = r[1].(thing) }
	return r_0, r_1
*/
func buildMockMethod(recv *ast.FieldList, name string, t *ast.FuncType) *ast.FuncDecl {
	stmts := []ast.Stmt{}
	p, err := storeParams(t.Params)
	if err != nil {
		fmt.Printf("Failed to set up call parameters. %v", err)
	}
	stmts = append(stmts, p...)
	p, err = trackCall(t.Results.NumFields(), name)
	if err != nil {
		fmt.Printf("failed to track call. %v", err)
	}
	stmts = append(stmts, p...)

	p, err = declReturnValues(t.Results)
	if err != nil {
		fmt.Printf("failed to declare return values. %v", err)
	}
	stmts = append(stmts, p...)

	p, err = buildReturnStatement(t.Results.NumFields())
	if err != nil {
		fmt.Printf("failed to build return statement. %v", err)
	}
	if p != nil {
		stmts = append(stmts, p...)
	}

	// This is our method declaration
	return &ast.FuncDecl{
		Type: t,
		Name: ast.NewIdent(name),
		Recv: recv,
		Body: &ast.BlockStmt{
			List: stmts,
		},
	}
}

// storeParams handles parameters
//
// Because our parameters may contain an ellipsis we always need to add all the parameters
// to an interface{} array
//
//  params := []interface{}{}
//  params[0] = p1
//  params[1] = p2
//  for i, p := range ellipsisParam {
//      params[2+i]	= p
//  }
func storeParams(params *ast.FieldList) ([]ast.Stmt, error) {
	code := ""
	// Is there an ellipsis parameter?
	listlen := len(params.List)
	if listlen > 0 {
		last := params.List[len(params.List)-1]
		if _, ok := last.Type.(*ast.Ellipsis); ok {
			code += fmt.Sprintf("\tparams := make([]interface{}, %d + len(%s))\n", params.NumFields()-1, last.Names[0].Name)
		}
	}

	if code == "" {
		// No ellipsis
		code += fmt.Sprintf("\tparams := make([]interface{}, %d)\n", params.NumFields())
	}

	i := 0
	for _, f := range params.List {
		for _, n := range f.Names {
			if _, ok := f.Type.(*ast.Ellipsis); ok {
				// Ellipsis expression
				code += fmt.Sprintf(`
    for j, p := range %s {
    	params[%d+j] = p
    }
`, n.Name, i)
			} else {
				code += fmt.Sprintf("\tparams[%d] = %s\n", i, n.Name)
			}
			i++
		}
	}

	return parseCodeBlock(code)
}

// trackCall builds the ast for the call expression.
//
// The call looks like
//     r := i.TrackCall("method", params...)
//
// If there are no return values r := is omitted
func trackCall(numReturns int, methodName string) ([]ast.Stmt, error) {
	if numReturns == 0 {
		return parseCodeBlock(fmt.Sprintf("\ti.TrackCall(\"%s\", params...)", methodName))
	}
	return parseCodeBlock(fmt.Sprintf("\tr := i.TrackCall(\"%s\", params...)", methodName))
}

// declReturnValues builds the return part of the call
//
func declReturnValues(results *ast.FieldList) ([]ast.Stmt, error) {
	if results.NumFields() == 0 {
		return nil, nil
	}
	stmts := []ast.Stmt{}
	for i, f := range results.List {
		// var r_X type
		stmts = append(stmts, &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							ast.NewIdent(fmt.Sprintf("r_%d", i)),
						},
						Type: f.Type,
					},
				},
			},
		})
		// if r[X] != nil {
		//     r_X = r[X].(type)
		// }
		stmts = append(stmts, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X: &ast.IndexExpr{
					X: ast.NewIdent("r"),
					Index: &ast.BasicLit{
						Kind:  token.INT,
						Value: fmt.Sprintf("%d", i),
					},
				},
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{
							ast.NewIdent(fmt.Sprintf("r_%d", i)),
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.TypeAssertExpr{
								X: &ast.IndexExpr{
									X: ast.NewIdent("r"),
									Index: &ast.BasicLit{
										Kind:  token.INT,
										Value: fmt.Sprintf("%d", i),
									},
								},
								Type: f.Type,
							},
						},
					},
				},
			},
		})
	}

	return stmts, nil
}

// buildReturnStatement
//
// return r_0, r_1, r_2
func buildReturnStatement(count int) ([]ast.Stmt, error) {
	r := &ast.ReturnStmt{}
	for i := 0; i < count; i++ {
		r.Results = append(r.Results, ast.NewIdent(fmt.Sprintf("r_%d", i)))
	}
	return []ast.Stmt{r}, nil
}

func generateMock(f *flags) {
	fset := token.NewFileSet()
	p, err := parser.ParseFile(fset, f.gofile, nil, 0)
	if err != nil {
		panic(err)
	}

	v := &InterfaceVisitor{fset: fset, name: f.ifName, packageName: f.packageName}
	ast.Walk(v, p)

	outfile := f.outfile
	if outfile == "" {
		outfile = fmt.Sprintf("mock%s.go", strings.ToLower(f.ifName))
	}

	err = ioutil.WriteFile(outfile, []byte(v.code), 0666)
	if err != nil {
		fmt.Printf("Failed to open %s for writing", outfile)
		os.Exit(2)
	}
}

type flags struct {
	gofile      string
	ifName      string
	outfile     string
	packageName string
}

func (f *flags) validate() bool {
	if f.gofile == "" {
		return false
	}
	if f.ifName == "" {
		return false
	}
	if f.packageName == "" {
		return false
	}
	return true
}

func (f *flags) setup() {
	flag.StringVar(&f.gofile, "filename", "", "The file that contains the interface definition; Must be specified.")
	flag.StringVar(&f.ifName, "interface", "", "The interface that we should create a mock for; Must be specified.")
	flag.StringVar(&f.outfile, "outfile", "", "The file to create the mock in. By default will use mock<interface>.go in the current directory.")
	flag.StringVar(&f.packageName, "package", "", "Package name to use for the mock file; Must be specified.")
}

func main() {
	f := &flags{}
	f.setup()

	flag.Parse()

	if !f.validate() {
		flag.Usage()
		os.Exit(2)
	}

	generateMock(f)
}
