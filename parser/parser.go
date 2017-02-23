package parser

import (
	"fmt"
	"io/ioutil"
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
	p.next()

	switch p.tok {
	case token.SELECT:
		stmt := ast.SelectStmt{
			Begin: p.pos,
		}
		slctstmt := p.parseSelect()
		stmt.Select = slctstmt

		from := p.parseFrom()
		stmt.From = from

		where := p.parseWhere()
		stmt.Where = where

		groupby := p.parseGroupby()
		stmt.Groupby = groupby

		orderby := p.parseOrderby()
		stmt.Orderby = orderby

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
		case token.WHERE, token.EOF, token.SEMICOLON, token.GROUP, token.ORDER:
			break L
		default:
			tbl := p.parseTable()
			tables = append(tables, &tbl)
		}
	}
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
		p.next()
	}
	return ast.Table{Value: expr, Alias: alias, EndPos: endPos}
}

func (p *parser) parseTableExpr() ast.TableExpr {
	switch p.tok {
	case token.IDENT:
		begin := p.pos
		kind := p.tok
		name := p.lit
		p.next()
		return ast.TableBasicLit{Begin: begin, Kind: kind, Name: name}
	default:
		// TODO mock return
		p.next()
		return ast.TableBasicLit{}
	}
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

func (p *parser) parseGroupby() ast.GroupbyClause {
	pos := p.pos
	exist := p.expect(token.GROUP)
	if !exist {
		return ast.GroupbyClause{Exists: false}
	}
	if !p.expect(token.BY) {
		panic("parser expect BY token. but got " + p.tok.String())
	}
	clus := ast.GroupbyClause{Begin: pos, ByPos: p.pos, Exists: true}
	var groups []ast.Expr
L:
	for {
		switch p.tok {
		case token.EOF, token.ORDER:
			p.next()
			break L
		case token.COMMA:
			p.next()
			continue
		default:
			groups = append(groups, p.parseExpr())
		}
	}

	clus.Groups = groups
	return clus
}

func (p *parser) parseOrderby() ast.OrderbyClause {
	pos := p.pos
	exist := p.expect(token.ORDER)
	if !exist {
		return ast.OrderbyClause{Exists: false}
	}
	if !p.expect(token.BY) {
		panic("parser expect BY token. but got " + p.tok.String())
	}
	clus := ast.OrderbyClause{Begin: pos, ByPos: p.pos, Exists: true}
	var orders []ast.Expr
L:
	for {
		switch p.tok {
		case token.EOF:
			p.next()
			break L
		case token.COMMA:
			p.next()
			continue
		default:
			orders = append(orders, p.parseExpr())
		}
	}

	clus.Orders = orders
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
	maybeIsPos := p.pos
	if p.expect(token.IS) {
		nullPos := p.pos
		if p.expect(token.NULL) {
			x = ast.IsNullExpr{Value: x, IsPos: maybeIsPos, NullPos: nullPos}
		} else {
			panic("parser expecte NULL token but got " + p.tok.String())
		}
	}
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
	case token.CASE:
		return p.parseCaseExpr()
	}
	return p.parsePrimaryExpr()
}

func (p *parser) parseCaseExpr() ast.CaseExpr {
	begin := p.pos
	if !p.expect(token.CASE) {
		panic("parser expects CASE token. but got " + p.tok.String())
	}

	var key ast.Expr
	switchKeyExists := false
	if p.tok != token.WHEN {
		switchKeyExists = true
		key = p.parseExpr()
	}

	var whens []*ast.WhenClause
L:
	for {
		switch p.tok {
		case token.WHEN:
			w := p.parseWhenClause()
			whens = append(whens, w)
			// whens = append(whens, p.parseWhenClause())
		case token.ELSE, token.END:
			break L
		default:
			panic("paraser expect WHEN, ELSE, END token. but got " + p.tok.String())
		}
	}

	elseClus := ast.ElseClause{Exists: false}
	maybeElsePos := p.pos
	if p.expect(token.ELSE) {
		elseClus.Begin = maybeElsePos
		elseClus.Exists = true
		elseClus.ResultExpr = p.parseExpr()
	}

	endPos := p.pos
	if !p.expect(token.END) {
		panic("parser expects END token. but got " + p.tok.String())
	}

	return ast.CaseExpr{Begin: begin, HasSwitchKey: switchKeyExists, SwitchKey: key, Whens: whens, Else: elseClus, EndPos: endPos}
}

func (p *parser) parseWhenClause() *ast.WhenClause {
	begin := p.pos
	if !p.expect(token.WHEN) {
		panic("parser expects WHEN token. but got " + p.tok.String())
	}

	cond := p.parseExpr()

	thenPos := p.pos
	if !p.expect(token.THEN) {
		panic("parser expects THEN token. but got " + p.tok.String())
	}

	result := p.parseExpr()

	return &ast.WhenClause{Begin: begin, CondExpr: cond, ThenPos: thenPos, ResultExpr: result}
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
			// Maybe function name
		} else if p.tok == token.LPAREN {
			lparen := p.pos
			var rparen token.Pos
			var args []ast.Expr
			p.next()
		L:
			for {
				switch p.tok {
				case token.RPAREN:
					rparen = p.pos
					p.next()
					break L
				case token.COMMA:
					p.next()
					continue
				case token.EOF:
					panic("while parsing CallExpr, got EOF.")
				default:
					args = append(args, p.parseExpr())
				}

			}

			return ast.CallExpr{Begin: pos, FuncName: lit, Lparen: lparen, Args: args, Rparen: rparen}
		}

		return ast.Ident{TblName: tbl, LitPos: pos, Kind: kind, Lit: lit}
	case token.STRING, token.INT, token.REAL:
		blit := ast.BasicLit{Begin: p.pos, Value: p.lit, Kind: p.tok}
		p.next()
		return blit
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
