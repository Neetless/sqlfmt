package parser

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Neetless/sqlfmt/ast"
	"github.com/Neetless/sqlfmt/scanner"
	"github.com/Neetless/sqlfmt/token"
)

// WrongAstTypeError represents the error which says the AST type is not expected one.
type WrongAstTypeError struct {
	msg string
}

// Implementation for error interface.
func (e WrongAstTypeError) Error() string {
	return e.msg
}

type parser struct {
	scanner scanner.Scanner
	file    *token.File

	// Next token
	pos token.Pos
	tok token.Token
	lit string
}

// ParseFile parse sql statement from given file.
func ParseFile(fset *token.FileSet, filename string, src interface{}) (ast.Stmt, error) {
	var text []byte
	if src != nil {
		switch s := src.(type) {
		case string:
			text = []byte(s)
		case []byte:
			text = s
		default:
			var stmt ast.Stmt
			return stmt, fmt.Errorf("src expect string or []byte but got %T", s)
		}
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}

		src, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		text = src
		file.Close()
	}
	stmt, err := parse(text, filename)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func parse(src []byte, filename string) (ast.Stmt, error) {
	var p parser
	var s scanner.Scanner

	fs := token.NewFileSet()
	ef := func(pos token.Position, msg string) {
		fmt.Fprintf(os.Stderr, "Error occured at position: %d. Message: %s.", pos, msg)
	}
	p.file = fs.AddFile(filename, -1, len(src))

	s.Init(p.file, src, ef, scanner.ScanComments)
	p.scanner = s

	stmt, err := p.parseStmt()
	if err != nil {
		return stmt, err
	}
	return stmt, nil
}

func (p *parser) parseStmt() (ast.Stmt, error) {
	pos, tok, lit := p.scanner.Scan()
	p.pos = pos
	p.tok = tok
	p.lit = lit

	switch tok {
	case token.SELECT:
		stmt := ast.SelectStmt{
			Begin: pos,
		}
		slctstmt := p.parseSelect()
		stmt.Select = slctstmt

		from := p.parseFrom()
		stmt.From = from

		where := p.parseWhere()
		stmt.Where = where

		return stmt, nil

	default:
		var stmt ast.Stmt
		return stmt, fmt.Errorf("cannot parsed")
	}

}

func (p *parser) parseSelect() ast.SelectClause {
	pos := p.pos
	if !p.expect(token.SELECT) {
		panic("select token expected but given " + p.tok.String())
	}

	cols := p.parseColumns()
	return ast.SelectClause{Begin: pos, Cols: cols}
}

// expect check check if the current token is same as expected.
// If it's expected, scan next.
func (p *parser) expect(tok token.Token) bool {
	if tok == p.tok {
		p.next()
		return true
	}
	return false
}

func (p *parser) parseFrom() ast.FromClause {
	pos := p.pos
	if !p.expect(token.FROM) {
		panic("from keyword is expected but get " + p.tok.String())
	}

	tables := p.parseTableList()
	return ast.FromClause{Begin: pos, Tables: tables}
}

func (p *parser) parseTableList() []*ast.Table {
	var tables []*ast.Table
L:
	for {
		switch p.tok {
		case token.COMMA:
			p.next()
			continue
		// TODO implement token.GROUPBY, token.ORDERBY
		case token.WHERE, token.EOF:
			break L
		default:
			tbl := p.parseTable()
			tables = append(tables, &tbl)
		}
	}
	log.Printf("return tables size %d\n", len(tables))
	return tables
}

func (p *parser) parseTable() ast.Table {
	expr := p.parseTableExpr()
	alias := ""
	endPos := expr.End()
	// TODO implement implicit alias
	if p.expect(token.ALIAS) {
		alias = p.lit
		endPos = p.pos + token.Pos(len(alias))
	}
	return ast.Table{Value: expr, Alias: alias, EndPos: endPos}
}

func (p *parser) parseTableExpr() ast.Table {

}

func (p *parser) parseWhere() ast.WhereClause {
	pos := p.pos
	exist := p.expect(token.WHERE)
	if !exist {
		return ast.WhereClause{Exists: false}
	}
	clus := ast.WhereClause{Begin: pos, Exists: true}
	expr := p.parseExpr()

	clus.CondExpr = expr
	return clus
}

func (p *parser) parseExpr() ast.Expr {
	return p.parseBinaryExpr(token.LowestPrec + 1)
}

func (p *parser) parseColumns() []*ast.Column {
	var cols []*ast.Column

L:
	for {
		switch p.tok {
		case token.COMMA:
			p.next()
		case token.FROM:
			break L
		default:
			col := p.parseColumn()
			cols = append(cols, &col)
		}
	}
	return cols
}

func (p *parser) parseColumn() ast.Column {
	expr := p.parseExpr()
	alias := ""
	var endPos token.Pos
	if p.expect(token.ALIAS) {
		alias = p.lit
		endPos = p.pos + token.Pos(len(alias))
		p.next()
	} else {
		endPos = expr.End()
	}
	return ast.Column{Value: expr, Alias: alias, EndPos: endPos}
}

func (p *parser) parseBinaryExpr(prec1 int) ast.Expr {
	x := p.parseUnaryExpr()
	for {
		op, opPrec := p.tokPrec()
		if opPrec < prec1 {
			return x
		}
		pos := p.pos

		p.expect(op)
		y := p.parseBinaryExpr(opPrec + 1)
		x = ast.BinaryExpr{X: x, OpPos: pos, Op: op, Y: y}
	}
}

func (p *parser) parseUnaryExpr() ast.Expr {
	switch p.tok {
	case token.ADD, token.SUB:
		pos, op := p.pos, p.tok
		p.next()
		x := p.parseUnaryExpr()
		return ast.UnaryExpr{OpPos: pos, Op: op, X: x}
	case token.MUL:
		pos := p.pos
		p.next()
		return ast.BasicLit{Begin: pos, Value: "*", Kind: token.ASTA}
	}
	return p.parsePrimaryExpr()
}

func (p *parser) parsePrimaryExpr() ast.Expr {
	switch p.tok {
	case token.IDENT:
		pos := p.pos
		kind := p.tok
		lit := p.lit
		tbl := ""

		p.next()

		// Maybe table name
		if p.tok == token.PERIOD {
			p.next()
			tbl = lit
			lit = p.lit
			if !p.expect(token.IDENT) {
				panic("expect column name after table name. but got " + p.tok.String())
			}
		}

		return ast.Ident{TblName: tbl, LitPos: pos, Kind: kind, Lit: lit}
	case token.STRING:
		return ast.BasicLit{Begin: p.pos, Value: p.lit, Kind: p.tok}
	}

	// TODO mock return
	var dummy ast.Expr
	return dummy
}

func (p *parser) next() {
	p.pos, p.tok, p.lit = p.scanner.Scan()
}

func (p *parser) tokPrec() (token.Token, int) {
	tok := p.tok
	return tok, tok.Precedence()
}
