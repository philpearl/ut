package main

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// This generates some basic scaffold that looks like:
// type mockName struct {
//   ut.CallTracker
// }

// func NewmockName(t *testing.T) *mockName {
//   return &mockName{ut.NewCallRecords(t)}
// }
// func (m *mockName) AddCall(name string, params ...any) ut.CallTracker {
//   m.CallTracker.AddCall(name, params)
//   return m
// }
// func (m *mockName) SetReturns(params ...any) ut.CallTracker {
//   m.CallTracker.SetReturns(params)
//   return m
// }
func genBasicDecls(mockName string, methodNames []string) []ast.Decl {
	return []ast.Decl{
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(mockName),
					Type: &ast.StructType{
						Fields: &ast.FieldList{
							List: []*ast.Field{{Type: &ast.SelectorExpr{X: ast.NewIdent("ut"), Sel: ast.NewIdent("CallTracker")}}},
						},
					},
				},
			},
		},
		&ast.FuncDecl{
			Name: ast.NewIdent(constructorName(mockName)),
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{ast.NewIdent("t")},
							Type: &ast.StarExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("testing"),
									Sel: ast.NewIdent("T"),
								},
							},
						},
					},
				},
				Results: &ast.FieldList{
					List: []*ast.Field{
						{
							Type: &ast.StarExpr{
								X: ast.NewIdent(mockName),
							},
						},
					},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.UnaryExpr{
								Op: token.AND,
								X: &ast.CompositeLit{
									Type: ast.NewIdent(mockName),
									Elts: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("ut"),
												Sel: ast.NewIdent("NewCallRecords"),
											},
											Args: []ast.Expr{
												ast.NewIdent("t"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		&ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("m"),
						},
						Type: &ast.StarExpr{
							X: ast.NewIdent(mockName),
						},
					},
				},
			},
			Name: ast.NewIdent("AddCall"),
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								ast.NewIdent("name"),
							},
							Type: ast.NewIdent("string"),
						},
						{
							Names: []*ast.Ident{
								ast.NewIdent("params"),
							},
							Type: &ast.Ellipsis{
								Elt: ast.NewIdent("any"),
							},
						},
					},
				},
				Results: &ast.FieldList{
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
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.SwitchStmt{
						Tag: &ast.Ident{
							Name: "name",
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.CaseClause{
									List: stringLiteralList(methodNames),
									Body: []ast.Stmt{
										&ast.BranchStmt{
											Tok: token.BREAK,
										},
									},
								},
								&ast.CaseClause{
									List: nil, // Default case.
									Body: []ast.Stmt{
										&ast.ExprStmt{
											X: &ast.CallExpr{
												Fun: &ast.Ident{Name: "panic"},
												Args: []ast.Expr{
													&ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X:   &ast.Ident{Name: "fmt"},
															Sel: &ast.Ident{Name: "Errorf"},
														},
														Args: []ast.Expr{
															&ast.BasicLit{
																Kind:  token.STRING,
																Value: strconv.Quote("AddCall: %T has no method %s"),
															},
															&ast.Ident{Name: "m"},
															&ast.Ident{Name: "name"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("m"),
									Sel: ast.NewIdent("CallTracker"),
								},
								Sel: ast.NewIdent("AddCall"),
							},
							Args: []ast.Expr{
								ast.NewIdent("name"),
								ast.NewIdent("params..."),
							},
						},
					},
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("m"),
						},
					},
				},
			},
		},

		&ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("m"),
						},
						Type: &ast.StarExpr{
							X: ast.NewIdent(mockName),
						},
					},
				},
			},
			Name: ast.NewIdent("SetReturns"),
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								ast.NewIdent("params"),
							},
							Type: &ast.Ellipsis{
								Elt: ast.NewIdent("any"),
							},
						},
					},
				},
				Results: &ast.FieldList{
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
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("m"),
									Sel: ast.NewIdent("CallTracker"),
								},
								Sel: ast.NewIdent("SetReturns"),
							},
							Args: []ast.Expr{
								ast.NewIdent("params..."),
							},
						},
					},
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("m"),
						},
					},
				},
			},
		},
	}
}

func constructorName(typeName string) string {
	if len(typeName) > 0 && typeName[0] >= 'a' && typeName[0] <= 'z' {
		return "new" + strings.Title(typeName)
	}
	return "New" + typeName
}

func stringLiteralList(vs []string) []ast.Expr {
	expr := make([]ast.Expr, len(vs))
	for i, v := range vs {
		expr[i] = &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(v),
		}
	}
	return expr
}
