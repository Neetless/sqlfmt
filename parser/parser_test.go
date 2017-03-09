package parser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Neetless/sqlfmt/ast"
	"github.com/Neetless/sqlfmt/token"
)

type testData struct {
	testSQL string
	expect  ast.Stmt
}

func setTestData() []testData {
	var testSet []testData
	testSet = []testData{
		testData{testSQL: `-- comment 1
select c from t`,
			expect: ast.SelectStmt{Begin: 14, Select: ast.SelectClause{Begin: 14, Cols: []*ast.Column{&ast.Column{Value: ast.Ident{TblName: "", LitPos: 21, Kind: token.IDENT, Lit: "c"}, Alias: "", EndPos: 22}}},
				From:     ast.FromClause{Begin: 23, Tables: []*ast.Table{&ast.Table{Value: ast.TableBasicLit{Begin: 28, Kind: token.IDENT, Name: "t"}, Alias: "", EndPos: 29}}},
				Where:    ast.WhereClause{Exists: false},
				Groupby:  ast.GroupbyClause{Exists: false},
				Orderby:  ast.OrderbyClause{Exists: false},
				Comments: []*ast.CommentGroup{&ast.CommentGroup{List: []*ast.Comment{&ast.Comment{Begin: 1, Text: "-- comment 1"}}}},
			},
		},
		/* TODO test joined table parser
		testData{testSQL: `select * from a natural left outer join b`,
			expect: ast.SelectStmt{},
		},
		testData{testSQL: `select * from a join b`,
			expect: ast.SelectStmt{Begin: 1, Select: ast.SelectClause{Begin: 1, Cols: []*ast.Column{&ast.Column{Value: ast.BasicLit{Begin: 8, Value: "*", Kind: token.ASTA}, Alias: "", EndPos: 9}}},
				From: ast.FromClause{Begin: 10, Tables: []*ast.Table{&ast.Table{
					Value: ast.JoinedTable{Left: ast.TableBasicLit{Begin: 15, Kind: token.IDENT, Name: "a"},
						Kind: token.NILL, JoinType: token.NILL, OuterJoinType: token.NILL,
						JoinPos: 999,
						Right:   ast.TableBasicLit{Begin: 15, Kind: token.IDENT, Name: "b"},
					},
					Alias:  "",
					EndPos: 23,
				}}},
				Where:   ast.WhereClause{Exists: false},
				Groupby: ast.GroupbyClause{Exists: false},
				Orderby: ast.OrderbyClause{Exists: false},
			},
		},
		*/
		testData{testSQL: `select c from t where t.v is null`,
			expect: ast.SelectStmt{Begin: 1, Select: ast.SelectClause{Begin: 1, Cols: []*ast.Column{&ast.Column{Value: ast.Ident{TblName: "", LitPos: 8, Kind: token.IDENT, Lit: "c"}, Alias: "", EndPos: 9}}},
				From:    ast.FromClause{Begin: 10, Tables: []*ast.Table{&ast.Table{Value: ast.TableBasicLit{Begin: 15, Kind: token.IDENT, Name: "t"}, Alias: "", EndPos: 16}}},
				Where:   ast.WhereClause{Begin: 17, CondExpr: ast.IsNullExpr{Value: ast.Ident{TblName: "t", LitPos: 23, Kind: token.IDENT, Lit: "v"}, IsPos: 27, NullPos: 30}, Exists: true},
				Groupby: ast.GroupbyClause{Exists: false},
				Orderby: ast.OrderbyClause{Exists: false},
			},
		},
		testData{testSQL: `select case
when code > 1 then '1'
when code < 2 then '2' end from tbl`,
			expect: ast.SelectStmt{
				Begin: 1,
				Select: ast.SelectClause{Begin: 1, Cols: []*ast.Column{
					&ast.Column{
						Value: ast.CaseExpr{
							Begin:        8,
							HasSwitchKey: false,
							Whens: []*ast.WhenClause{
								&ast.WhenClause{
									Begin:      13,
									CondExpr:   ast.BinaryExpr{X: ast.Ident{TblName: "", LitPos: 18, Kind: token.IDENT, Lit: "code"}, OpPos: 22, Op: token.GTR, Y: ast.BasicLit{Begin: 25, Value: "1", Kind: token.INT}},
									ThenPos:    27,
									ResultExpr: ast.BasicLit{Begin: 32, Value: "'1'", Kind: token.STRING},
								},
								&ast.WhenClause{
									Begin:      36,
									CondExpr:   ast.BinaryExpr{X: ast.Ident{TblName: "", LitPos: 41, Kind: token.IDENT, Lit: "code"}, OpPos: 46, Op: token.LSS, Y: ast.BasicLit{Begin: 48, Value: "2", Kind: token.INT}},
									ThenPos:    50,
									ResultExpr: ast.BasicLit{Begin: 55, Value: "'2'", Kind: token.STRING},
								},
							},
							EndPos: 59,
						},
						Alias:  "",
						EndPos: 62,
					},
				}},
				From:    ast.FromClause{Begin: 63, Tables: []*ast.Table{&ast.Table{Value: ast.TableBasicLit{Begin: 68, Kind: token.IDENT, Name: "tbl"}, Alias: "", EndPos: 71}}},
				Where:   ast.WhereClause{Exists: false},
				Groupby: ast.GroupbyClause{Exists: false},
				Orderby: ast.OrderbyClause{Exists: false},
			},
		},
		testData{testSQL: `select case code when '0' then '1' else '2' end from tbl`,
			expect: ast.SelectStmt{
				Begin: 1,
				Select: ast.SelectClause{Begin: 1, Cols: []*ast.Column{
					&ast.Column{
						Value: ast.CaseExpr{
							Begin:        8,
							HasSwitchKey: true,
							SwitchKey:    ast.Ident{TblName: "", LitPos: 13, Kind: token.IDENT, Lit: "code"},
							Whens: []*ast.WhenClause{&ast.WhenClause{
								Begin:      18,
								CondExpr:   ast.BasicLit{Begin: 23, Value: "'0'", Kind: token.STRING},
								ThenPos:    27,
								ResultExpr: ast.BasicLit{Begin: 32, Value: "'1'", Kind: token.STRING},
							}},
							Else:   ast.ElseClause{Begin: 36, ResultExpr: ast.BasicLit{Begin: 41, Value: "'2'", Kind: token.STRING}, Exists: true},
							EndPos: 45,
						},
						Alias:  "",
						EndPos: 48,
					},
				}},
				From:    ast.FromClause{Begin: 49, Tables: []*ast.Table{&ast.Table{Value: ast.TableBasicLit{Begin: 54, Kind: token.IDENT, Name: "tbl"}, Alias: "", EndPos: 57}}},
				Where:   ast.WhereClause{Exists: false},
				Groupby: ast.GroupbyClause{Exists: false},
				Orderby: ast.OrderbyClause{Exists: false},
			},
		},
		testData{testSQL: `select score from tbl order by score`,
			expect: ast.SelectStmt{
				Begin:   1,
				Select:  ast.SelectClause{Begin: 1, Cols: []*ast.Column{&ast.Column{Value: ast.Ident{LitPos: 8, Lit: "score", Kind: token.IDENT}, Alias: "", EndPos: 13}}},
				From:    ast.FromClause{Begin: 14, Tables: []*ast.Table{&ast.Table{Value: ast.TableBasicLit{Begin: 19, Kind: token.IDENT, Name: "tbl"}, Alias: "", EndPos: 22}}},
				Where:   ast.WhereClause{Exists: false},
				Groupby: ast.GroupbyClause{Exists: false},
				Orderby: ast.OrderbyClause{Begin: 29, ByPos: 35, Exists: true, Orders: []ast.Expr{ast.Ident{LitPos: 38, Lit: "score", Kind: token.IDENT}}},
			},
		},
		testData{testSQL: `select key from tbl GROUP BY key`,
			expect: ast.SelectStmt{
				Begin: 1,
				Select: ast.SelectClause{Begin: 1, Cols: []*ast.Column{&ast.Column{
					Value: ast.Ident{LitPos: 8, Lit: "key", Kind: token.IDENT}, Alias: "", EndPos: 11}}},
				From: ast.FromClause{Begin: 12, Tables: []*ast.Table{&ast.Table{
					Value: ast.TableBasicLit{Begin: 17, Kind: token.IDENT, Name: "tbl"}, Alias: "", EndPos: 20}}},
				Where:   ast.WhereClause{Exists: false},
				Groupby: ast.GroupbyClause{Begin: 21, ByPos: 27, Exists: true, Groups: []ast.Expr{ast.Ident{LitPos: 30, Kind: token.IDENT, Lit: "key"}}}},
		},
		testData{testSQL: `select count(*) from tbl;`,
			expect: ast.SelectStmt{
				Begin: 1,
				Select: ast.SelectClause{Begin: 1, Cols: []*ast.Column{&ast.Column{
					Value: ast.CallExpr{Begin: 8,
						FuncName: "count",
						Lparen:   13,
						Args:     []ast.Expr{ast.BasicLit{Begin: 14, Value: "*", Kind: token.ASTA}},
						Rparen:   15},
					Alias:  "",
					EndPos: 16}}},
				From: ast.FromClause{Begin: 17, Tables: []*ast.Table{&ast.Table{
					Value: ast.TableBasicLit{Begin: 22, Kind: token.IDENT, Name: "tbl"}, Alias: "", EndPos: 25}}},
				Where: ast.WhereClause{Exists: false}},
		},
		testData{
			testSQL: `  select id, username from id_mst, user_mst where id_mst.id = user_mst.id and user_mst.dt > '2015-12-01';`,
			expect: ast.SelectStmt{
				Begin: 3,
				Select: ast.SelectClause{
					Begin: 3,
					Cols: []*ast.Column{
						&ast.Column{
							Value:  ast.Ident{LitPos: 10, TblName: "", Kind: token.IDENT, Lit: "id"},
							Alias:  "",
							EndPos: 12,
						},
						&ast.Column{
							Value:  ast.Ident{LitPos: 14, TblName: "", Kind: token.IDENT, Lit: "username"},
							Alias:  "",
							EndPos: 22,
						},
					},
				},
				From: ast.FromClause{
					Begin: 23,
					Tables: []*ast.Table{
						&ast.Table{Value: ast.TableBasicLit{Begin: 28, Kind: token.IDENT, Name: "id_mst"},
							Alias:  "",
							EndPos: 34,
						},
						&ast.Table{
							Value:  ast.TableBasicLit{Begin: 36, Kind: token.IDENT, Name: "user_mst"},
							Alias:  "",
							EndPos: 44,
						},
					},
				},
				Where: ast.WhereClause{
					Exists: true,
					Begin:  45,
					CondExpr: ast.BinaryExpr{
						X: ast.BinaryExpr{
							X: ast.Ident{
								LitPos:  51,
								TblName: "id_mst",
								Lit:     "id",
								Kind:    token.IDENT,
							},
							OpPos: 61,
							Op:    token.EQL,
							Y: ast.Ident{
								LitPos:  63,
								TblName: "user_mst",
								Lit:     "id",
								Kind:    token.IDENT,
							},
						},
						OpPos: 75,
						Op:    token.AND,
						Y: ast.BinaryExpr{
							X: ast.Ident{
								LitPos:  79,
								TblName: "user_mst",
								Lit:     "dt",
								Kind:    token.IDENT,
							},
							OpPos: 91,
							Op:    token.GTR,
							Y: ast.BasicLit{
								Begin: 93,
								Value: "'2015-12-01'",
								Kind:  token.STRING,
							},
						},
					},
				},
			},
		},
		testData{
			// test alias and multi columns.
			testSQL: `select col1 as id, col2 as name from tbl1 as user, tbl2 as item;`,
			expect: ast.SelectStmt{
				Begin: 1,
				Select: ast.SelectClause{
					Begin: 1,
					Cols: []*ast.Column{
						&ast.Column{
							Value:  ast.Ident{LitPos: 8, TblName: "", Kind: token.IDENT, Lit: "col1"},
							Alias:  "id",
							EndPos: 18,
						},
						&ast.Column{
							Value:  ast.Ident{LitPos: 20, TblName: "", Kind: token.IDENT, Lit: "col2"},
							Alias:  "name",
							EndPos: 32,
						},
					},
				},
				From: ast.FromClause{
					Begin: 33,
					Tables: []*ast.Table{
						&ast.Table{
							Value: ast.TableBasicLit{
								Begin: 38,
								Kind:  token.IDENT,
								Name:  "tbl1",
							},
							Alias:  "user",
							EndPos: 50,
						},
						&ast.Table{
							Value: ast.TableBasicLit{
								Begin: 52,
								Kind:  token.IDENT,
								Name:  "tbl2",
							},
							Alias:  "item",
							EndPos: 64,
						},
					},
				},
				Where: ast.WhereClause{Exists: false},
			},
		},
	}
	return testSet
}

func TestParseFileWithSrc(t *testing.T) {
	// preparation
	ts := setTestData()
	fs := token.NewFileSet()

	// test
	for _, v := range ts {
		stmt, err := ParseFile(fs, "test.sql", v.testSQL)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("given SQL: %s", v.testSQL)
		nodeEqualTest(stmt, v.expect, t)
	}

}

// Test parsing select statement sql file.
func TestParseFile(t *testing.T) {
	fs := token.NewFileSet()
	stmt, err := ParseFile(fs, "testdata/select_test.sql", nil)

	if err != nil {
		t.Error(err)
	}

	expectStmt := ast.SelectStmt{
		Begin: 1,
		Select: ast.SelectClause{
			Begin: 1,
			Cols: []*ast.Column{
				&ast.Column{
					Value:  ast.BasicLit{Begin: 8, Kind: token.ASTA, Value: "*"},
					Alias:  "",
					EndPos: 9,
				},
			},
		},
		From: ast.FromClause{
			Begin: 10,
			Tables: []*ast.Table{
				&ast.Table{
					Value: ast.TableBasicLit{
						Begin: 15,
						Kind:  token.IDENT,
						Name:  "table1",
					},
					Alias:  "",
					EndPos: 21,
				},
			},
		},
		Where: ast.WhereClause{Exists: false},
	}
	if _, ok := stmt.(ast.SelectStmt); !ok {
		t.Error("Parsed statement is not SELECT statement.")
	}

	nodeEqualTest(stmt, expectStmt, t)
}

func nodeEqualTest(actual, expect ast.Node, t *testing.T) {
	t.Log("Node pos/end check.")
	posEqualTest(actual, expect, t)
	switch expectStmt := expect.(type) {
	case ast.SelectStmt:
		actualStmt, ok := actual.(ast.SelectStmt)
		if !ok {
			t.Errorf("actual type is not SelectStmt, is %T.", actual)
		}

		clauseEqualTest(actualStmt.Select, expectStmt.Select, t)
		clauseEqualTest(actualStmt.From, expectStmt.From, t)
		clauseEqualTest(actualStmt.Where, expectStmt.Where, t)
		commentGroupSliceEqualTest(actualStmt.Comments, expectStmt.Comments, t)
	default:
		t.Errorf("Expect's top level type is not correct.")
	}

}

func clauseEqualTest(actual, expect ast.Clause, t *testing.T) {
	switch expectClause := expect.(type) {
	case ast.SelectClause:
		actualClause, ok := actual.(ast.SelectClause)
		if !ok {
			t.Fatalf("actual type is not SelectClause.actual: %T.", actual)
		}

		t.Log("SelectClause pos/end check")
		posEqualTest(actualClause, expectClause, t)
		columnsEqualTest(actualClause.Cols, expectClause.Cols, t)

	case ast.FromClause:
		actualClause, ok := actual.(ast.FromClause)
		if !ok {
			t.Fatalf("actual type is not FromClause. actual: %T.", actual)
		}
		t.Log("FromClause pos/end check")
		posEqualTest(actualClause, expectClause, t)
		tablesEqualTest(actualClause.Tables, expectClause.Tables, t)

	case ast.WhereClause:
		actualClause, ok := actual.(ast.WhereClause)
		if !ok {
			t.Fatalf("actual type is not WhereClause. actual: %T.", actual)
		}

		if actualClause.Exists != expectClause.Exists {
			t.Fatalf(
				"WhereClause exist check fail. actual: %v, expected: %v.",
				actualClause.Exists,
				expectClause.Exists,
			)
		}
		posEqualTest(actualClause, expectClause, t)
		if actualClause.Exists {
			exprEqualTest(actualClause.CondExpr, expectClause.CondExpr, t)
		}
	default:
		t.Fatalf("Expect's top level type is not correct.")
	}
}

func tablesEqualTest(actual, expect []*ast.Table, t *testing.T) {
	// size check
	if len(actual) != len(expect) {
		t.Fatalf(
			"tables sizes are different. actual: %d, expect: %d.",
			len(actual),
			len(expect),
		)
	}

	for ix, actualTbl := range actual {
		posEqualTest(actualTbl, expect[ix], t)

		if actualTbl.Alias != expect[ix].Alias {
			t.Fatalf(
				"table Alias is incorrect. actual: %s, expect: %s.",
				actualTbl.Alias,
				expect[ix].Alias,
			)
		}

		tableExprEqualTest(actualTbl.Value, expect[ix].Value, t)
	}

}

func tableExprEqualTest(actual, expect ast.TableExpr, t *testing.T) {
	t.Log("TableExpr pos/end check.")
	var actualStruct, expectStruct string
	if tp := reflect.TypeOf(actual); tp.Kind() == reflect.Ptr {
		actualStruct = "*" + tp.Elem().Name()
		expectStruct = "*" + reflect.TypeOf(expect).Elem().Name()
	} else {
		actualStruct = tp.Name()
		expectStruct = reflect.TypeOf(expect).Name()

	}
	typemsg := fmt.Sprintf("actual type %s, expect: %s", actualStruct, expectStruct)
	posEqualTest(actual, expect, t)
	switch expectExpr := expect.(type) {
	case ast.TableBasicLit:
		actualExpr, ok := actual.(ast.TableBasicLit)
		if !ok {
			t.Fatal("actual type is not ast.TableBasicLit. " + typemsg)
		}
		if actualExpr.Kind != expectExpr.Kind {
			t.Fatalf(
				"TableBasicLit kind incorrect. actual: %s, expected: %s.",
				actualExpr.Kind.String(),
				expectExpr.Kind.String(),
			)
		}
		if actualExpr.Name != expectExpr.Name {
			t.Fatalf(
				"TableBasicLit value is incorrect. actual: %s, expected: %s.",
				actualExpr.Name,
				expectExpr.Name,
			)
		}
	default:
		t.Fatal("Unexpected type the expected tableExpr has. " + typemsg)
	}

}

func columnsEqualTest(actual, expect []*ast.Column, t *testing.T) {
	// size check
	if len(actual) != len(expect) {
		t.Fatalf(
			"columns sizes are different. actual: %d, expect: %d.",
			len(actual),
			len(expect),
		)
	}

	for ix, actualCol := range actual {
		posEqualTest(actualCol, expect[ix], t)

		if actualCol.Alias != expect[ix].Alias {
			t.Fatalf(
				"column Alias is incorrect. actual: %s, expect: %s.",
				actualCol.Alias,
				expect[ix].Alias,
			)
		}

		exprEqualTest(actualCol.Value, expect[ix].Value, t)
	}

}

func exprEqualTest(actual, expect ast.Expr, t *testing.T) {
	t.Log("Expr pos/end check.")
	var actualStruct, expectStruct string
	if tp := reflect.TypeOf(actual); tp.Kind() == reflect.Ptr {
		actualStruct = "*" + tp.Elem().Name()
		expectStruct = "*" + reflect.TypeOf(expect).Elem().Name()
	} else {
		actualStruct = tp.Name()
		expectStruct = reflect.TypeOf(expect).Name()

	}
	typemsg := fmt.Sprintf("actual type %s, expect: %s", actualStruct, expectStruct)
	posEqualTest(actual, expect, t)
	switch expectExpr := expect.(type) {
	case ast.BasicLit:
		actualExpr, ok := actual.(ast.BasicLit)
		if !ok {
			t.Fatal("actual type is not ast.BasicLit. " + typemsg)
		}
		if actualExpr.Kind != expectExpr.Kind {
			t.Fatalf(
				"BasicLit kind incorrect. actual: %s, expected: %s.",
				actualExpr.Kind.String(),
				expectExpr.Kind.String(),
			)
		}
		if actualExpr.Value != expectExpr.Value {
			t.Fatalf(
				"BasicLit values is incorrect. actual: %s, expected: %s.",
				actualExpr.Value,
				expectExpr.Value,
			)
		}

	case ast.Ident:
		actualExpr, ok := actual.(ast.Ident)
		if !ok {
			t.Fatal("actual type is not ast.Ident. " + typemsg)
		}
		if actualExpr.TblName != expectExpr.TblName {
			t.Fatalf(
				"Ident table name is incorrect. actual: %s, expected: %s.",
				actualExpr.TblName,
				expectExpr.TblName,
			)
		}
		if actualExpr.Kind != expectExpr.Kind {
			t.Fatalf(
				"Ident kind is incorrect. actual: %s, expected: %s.",
				actualExpr.Kind.String(),
				expectExpr.Kind.String(),
			)
		}
		if actualExpr.Lit != expectExpr.Lit {
			t.Fatalf(
				"Ident lit is incorrect. actual: %s, expected: %s.",
				actualExpr.Lit,
				expectExpr.Lit,
			)
		}
	case ast.BinaryExpr:
		actualExpr, ok := actual.(ast.BinaryExpr)
		if !ok {
			t.Fatal("actual type is not ast.BinaryExpr. " + typemsg)
		}
		if actualExpr.Op != expectExpr.Op {
			t.Fatalf("BinaryExpr op is incorrect. actual: %s, expect: %s.", actualExpr.Op, expectExpr.Op)
		}
		exprEqualTest(actualExpr.X, expectExpr.X, t)
		exprEqualTest(actualExpr.Y, expectExpr.Y, t)
	case ast.CallExpr:
		actualExpr, ok := actual.(ast.CallExpr)
		if !ok {
			t.Fatal("actual type is not ast.BinaryExpr. " + typemsg)
		}
		if actualExpr.FuncName != expectExpr.FuncName {
			t.Fatalf("CallExpr FuncName is incorrect. actual: %s, expect: %s.", actualExpr.FuncName, expectExpr.FuncName)
		}
		if len(actualExpr.Args) != len(expectExpr.Args) {
			t.Fatalf("CallExpr Args size is incorrect. actual: %d, expect: %d.", len(actualExpr.Args), len(expectExpr.Args))
		}
		for ix, actualArg := range actualExpr.Args {
			exprEqualTest(actualArg, expectExpr.Args[ix], t)
		}
	case ast.CaseExpr:
		actualExpr, ok := actual.(ast.CaseExpr)
		if !ok {
			t.Fatal("actual type is not ast.CaseExpr. ", typemsg)
		}

		if actualExpr.HasSwitchKey != expectExpr.HasSwitchKey {
			t.Fatal("case expression switch key exisitance is incorrect. actual: ", actualExpr.HasSwitchKey, " expect: ", expectExpr.HasSwitchKey)
		}
		if actualExpr.HasSwitchKey {
			exprEqualTest(actualExpr.SwitchKey, expectExpr.SwitchKey, t)
		}

		for ix, actualWhen := range actualExpr.Whens {
			posEqualTest(actualWhen, expectExpr.Whens[ix], t)
			exprEqualTest(actualWhen.CondExpr, expectExpr.Whens[ix].CondExpr, t)
			if actualWhen.ThenPos != expectExpr.Whens[ix].ThenPos {
				t.Fatal(ix, "th actual when's then position is incorrect. ")
			}
		}

		if actualExpr.Else.Exists != expectExpr.Else.Exists {
			t.Fatal("case else exists boolean is incorrect. actual ", actualExpr.Else.Exists, " expect ", expectExpr.Else.Exists)
		}

	case ast.IsNullExpr:
		actualExpr, ok := actual.(ast.IsNullExpr)
		if !ok {
			t.Fatal("actual type is not ast.IsNullExpr. ", typemsg)
		}
		exprEqualTest(actualExpr.Value, expectExpr.Value, t)

		if actualExpr.IsPos != expectExpr.IsPos {
			t.Fatal("actual is null expression has incorrect is token position. actual ", actualExpr.IsPos, " expect ", expectExpr.IsPos)
		}

		if actualExpr.NullPos != expectExpr.NullPos {
			t.Fatal("actual is null expression has incorrect null token position. actual ", actualExpr.NullPos, " expect ", expectExpr.NullPos)
		}

	default:
		t.Fatal("Unexpected type the expected expr has. " + typemsg)
	}

}

func commentGroupSliceEqualTest(actual, expect []*ast.CommentGroup, t *testing.T) {
	if len(actual) != len(expect) {
		t.Fatalf("actual len of []CommentGroup is incorrect. actual: %d, expect: %d", len(actual), len(expect))
	}

	for ix, actualCommentGroup := range actual {
		expectCommentGroup := expect[ix]
		if len(actualCommentGroup.List) != len(expectCommentGroup.List) {
			t.Fatalf("actual len of CommentGroup.List is incorrect. actual: %d, expect: %d.", len(actual), len(expect))
		}

		posEqualTest(actualCommentGroup, expectCommentGroup, t)

		for ix, actualComment := range actualCommentGroup.List {
			expectComment := expectCommentGroup.List[ix]
			posEqualTest(actualComment, expectComment, t)

			if actualComment.Text != expectComment.Text {
				t.Fatalf("actual comment text is incorrect. actual: %s, expect: %s.", actualComment.Text, expectComment.Text)
			}
		}
	}
}

func posEqualTest(actual, expect ast.Node, t *testing.T) {
	expectType := reflect.TypeOf(expect)
	actualType := reflect.TypeOf(actual)
	var expectTypeName, actualTypeName string
	if expectType.Kind() == reflect.Ptr {
		expectTypeName = expectType.Elem().Name()
		actualTypeName = actualType.Elem().Name()
	} else {
		expectTypeName = expectType.Name()
		actualTypeName = actualType.Name()
	}
	msg := fmt.Sprintf("types are actual: %s, expect: %s", actualTypeName, expectTypeName)
	if actual.Pos() != expect.Pos() {
		t.Fatalf(
			"Node has different pos. actual: %d, expect: %d. "+msg,
			actual.Pos(),
			expect.Pos(),
		)
	}
	if actual.End() != expect.End() {
		t.Fatalf(
			"Node has different end. actual: %d, expect %d. "+msg,
			actual.End(),
			expect.End(),
		)
	}
}
