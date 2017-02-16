package scanner

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Neetless/sqlfmt/token"
)

func defaultInit(src []byte) Scanner {
	ef := func(_ token.Position, msg string) {
		log.Printf("error handler called (msg = %s)", msg)
	}
	var s Scanner
	fset := token.NewFileSet()
	filename := "test.sql"
	s.Init(fset.AddFile(filename, fset.Base(), len(src)), src, ef, ScanComments|dontInsertSemis)
	return s
}

type scanSet struct {
	tok token.Token
	pos token.Pos
	lit string
}

type testSet struct {
	given  []byte
	expect []scanSet
}

func TestScan(t *testing.T) {
	ts := []testSet{
		testSet{given: []byte("1 11"), expect: []scanSet{
			scanSet{tok: token.INT, pos: 1, lit: "1"},
			scanSet{tok: token.INT, pos: 3, lit: "11"},
		}},
		testSet{given: []byte("CASE WHEN THEN ELSE END"), expect: []scanSet{
			scanSet{tok: token.CASE, pos: 1, lit: "CASE"},
			scanSet{tok: token.WHEN, pos: 6, lit: "WHEN"},
			scanSet{tok: token.THEN, pos: 11, lit: "THEN"},
			scanSet{tok: token.ELSE, pos: 16, lit: "ELSE"},
			scanSet{tok: token.END, pos: 21, lit: "END"},
		}},
		testSet{given: []byte("()"), expect: []scanSet{
			scanSet{tok: token.LPAREN, pos: 1, lit: "("},
			scanSet{tok: token.RPAREN, pos: 2, lit: ")"},
		}},
		testSet{given: []byte("> ="), expect: []scanSet{
			scanSet{tok: token.GTR, pos: 1, lit: ">"},
			scanSet{tok: token.EQL, pos: 3, lit: "="},
		}},
		testSet{given: []byte("tbl.col"), expect: []scanSet{
			scanSet{tok: token.IDENT, pos: 1, lit: "tbl"},
			scanSet{tok: token.PERIOD, pos: 4, lit: "."},
			scanSet{tok: token.IDENT, pos: 5, lit: "col"},
		}},
		testSet{given: []byte("and or"), expect: []scanSet{
			scanSet{tok: token.AND, pos: 1, lit: "and"},
			scanSet{tok: token.OR, pos: 5, lit: "or"},
		}},
		testSet{given: []byte("'2015-11-11'"), expect: []scanSet{
			scanSet{tok: token.STRING, pos: 1, lit: "'2015-11-11'"},
		}},
		testSet{given: []byte(", ."), expect: []scanSet{
			scanSet{tok: token.COMMA, pos: 1, lit: ","},
			scanSet{tok: token.PERIOD, pos: 3, lit: "."},
		}},
	}

	for ix, test := range ts {
		actual := []scanSet{}
		s := defaultInit(test.given)
		for {
			var ss scanSet
			ss.pos, ss.tok, ss.lit = s.Scan()
			if ss.tok == token.EOF {
				break
			}
			actual = append(actual, ss)
		}
		if err := isSameScanSetSlice(actual, test.expect); err != nil {
			t.Fatalf("%dth test set failed. given: %s. result: %s", ix, test.given, err.Error())
		}
	}
}

func isSameScanSetSlice(actual, expect []scanSet) error {
	if len(actual) != len(expect) {
		return fmt.Errorf("# of scanned is different with expected. actual: %v, expected: %v", actual, expect)
	}

	for ix, a := range actual {
		e := expect[ix]

		if a.pos != e.pos || a.tok != e.tok || a.lit != e.lit {
			msg := fmt.Sprintf("%dth scan is failed. actual: %v, expected: %v", ix, a, e)
			return fmt.Errorf(msg)
		}

	}

	return nil
}

func TestScanFromFile(t *testing.T) {
	ef := func(_ token.Position, msg string) {
		t.Errorf("error handler called (msg = %s)", msg)
	}

	var s Scanner
	fset := token.NewFileSet()
	file, err := os.Open("./testdata/test.sql")
	if err != nil {
		t.Errorf("Couldn't open file. %s\n", err)
	}
	source, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("Cannot read file. err: %s", err)
	}
	file.Close()

	s.Init(fset.AddFile("./testdata/test.sql", fset.Base(), len(source)), source, ef, ScanComments|dontInsertSemis)
	for {
		pos, tok, lit := s.Scan()
		if tok == token.SELECT {
			if strings.ToUpper(lit) != "SELECT" || pos != 1 {
				t.Errorf("Not correct Pos(%d) and Lit(%s) fro SELECT", pos, lit)
			}
			t.Log("Got SELECT.")
		}

		if tok == token.MUL {
			if lit != "*" || (pos != 17 && pos != 14) {
				t.Errorf("Not correct Pos(%d) and Lit(%s) for *", pos, lit)
			}
			t.Log("Got MUL.")
		}

		if tok == token.INT {
			if lit != "100" || pos != 19 {
				t.Errorf("Not correct Pos(%d) and Lit(%s) for *", pos, lit)
			}
			t.Log("Got INT.")
		}

		if tok == token.FROM {
			if lit != "from" || pos != 23 {
				t.Errorf("Not correct Pos(%d) and Lit(%s) for FROM", pos, lit)
			}
			t.Log("Got FROM")
		}

		if tok == token.LPAREN {
			if lit != "(" || pos != 13 {
				t.Errorf("Not correct Pos(%d) and Lit(%s) for LPAREN", pos, lit)
			}
			t.Log("Got LPAREN")
		}
		if tok == token.RPAREN {
			if lit != ")" || pos != 15 {
				t.Errorf("Not correct Pos(%d) and Lit(%s) for RPAREN", pos, lit)
			}
			t.Log("Got RPAREN")
		}

		if tok == token.IDENT {
			switch lit {
			case "count":
				if pos != 8 {
					t.Errorf("Not correct Pos(%d) for %s", pos, lit)
				}
			case "mastertbl":
				if pos != 28 {
					t.Errorf("Not correct Pos(%d) for %s", pos, lit)
				}
			default:
				t.Errorf("Not correct lit for IDENT %s", lit)
			}
			t.Logf("Got IDENT: %s", lit)
		}

		if tok == token.SEMICOLON {
			if lit != ";" || pos != 37 {
				t.Errorf("Not correct Pos(%d) and Lit(%s) for SEMICOLON", pos, lit)
			}
			t.Log("Got SEMICOLON")
		}
		if tok == token.EOF {
			t.Log("Got EOF.")
			break
		}
	}
}
