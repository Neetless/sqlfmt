package ast

import (
	"bytes"

	"github.com/Neetless/sqlfmt/parser"
	"github.com/Neetless/sqlfmt/token"

	"testing"
)

type testSQLSet struct {
	input  []byte
	expect string
	actual string
}

func TestFprint(t *testing.T) {
	testSet := []testSQLSet{
		testSQLSet{},
	}
	for i := range testSet {
		t.Errorf("not implemented yet %d", i)
	}
}

func TestFprintFromFile(t *testing.T) {
	// preparation
	fset := token.NewFileSet()
	stmt, err := parser.ParseFile(fset, "../parser/testdata/select_test.sql", nil)
	if err != nil {
		t.Fatal(err)
	}

	expect := `SELECT
    *
FROM
    table1
;`
	var buf []byte
	out := bytes.NewBuffer(buf)

	// test
	if err := Fprint(out, fset, stmt); err != nil {
		t.Fatal(err)
	}
	if out.String() != expect {
		t.Errorf("Fprint from File failed. expect:\n%s\nactual:\n%s", expect, out.String())
	}
}
