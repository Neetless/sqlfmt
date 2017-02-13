package ast

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/Neetless/sqlfmt/ast"
	"github.com/Neetless/sqlfmt/token"
)

type printer struct {
	Config
	fset   *token.FileSet
	indent int

	output []byte

	outputPos token.Position
}

// Fprint "pretty-prints" an AST node to Fprint.
func Fprint(out io.Writer, fset *token.FileSet, node interface{}) error {
	var p printer

	// default values for Config.
	p.ImpliedSemi = true
	p.IndentWidth = 4
	p.NewlineChar = []byte("\n")

	// set printer fields.
	p.fset = fset
	p.outputPos = token.Position{Line: 1, Column: 1}

	if err := p.printNode(node); err != nil {
		return err
	}

	if _, err := out.Write(p.output); err != nil {
		return err
	}
	return nil
}

func (p *printer) printNode(node interface{}) error {
	switch n := node.(type) {
	case ast.SelectStmt:
		p.selectStmt(n)
		return nil
	default:
		return fmt.Errorf("gofmt/ast: unsupported node type %T", node)
	}
}

func (p *printer) selectStmt(node ast.SelectStmt) {
	p.selectClause(node.Select)

	p.fromClause(node.From)

	p.insertSemi()
}

func (p *printer) selectClause(node ast.SelectClause) {
	// Write SELECT keyword
	p.output = append(p.output, []byte(token.SELECT.String())...)
	p.indent++
	p.appendNewline()

	p.columnList(node.Cols)

}

func (p *printer) fromClause(node ast.FromClause) {
	p.output = append(
		p.output,
		[]byte(token.FROM.String())...,
	)
	p.indent++
	p.appendNewline()

	p.tableList(node.Tables)

}

func (p *printer) columnList(node []*ast.Column) {
	for i, v := range node {
		switch n := v.Value.(type) {
		case ast.BasicLit:
			expr := []byte(n.Kind.String())
			p.output = append(
				p.output,
				expr...,
			)
			p.outputPos.Column += utf8.RuneCount(expr)

			p.alias(v.Alias)
		}

		// when there are columns and v in this loop is not last, add camma.
		if i < len(node)-1 {
			p.output = append(p.output, []byte(",")...)
		} else if i == len(node)-1 { // when v is last column, adjust indent.

			p.indent--
		}
		p.appendNewline()
	}
}

func (p *printer) alias(name string) {
	if name != "" {
		// add alias string
	}
}

func (p *printer) tableList(tables []*ast.Table) {
	for i, v := range tables {
		switch n := v.Value.(type) {
		case ast.TableBasicLit:
			p.output = append(
				p.output,
				[]byte(n.Name)...,
			)
			p.alias(v.Alias)
		}
		// when there are columns and v in this loop is not last, add camma.
		if i < len(tables)-1 {
			p.output = append(p.output, []byte(",")...)
		} else if i == len(tables)-1 { // when v is last column, adjust indent.

			p.indent--
		}
		p.appendNewline()
	}
}

func (p *printer) appendNewline() {
	p.output = append(p.output, p.NewlineChar...)
	p.outputPos.Line++
	p.outputPos.Column = 1 + p.indent*p.IndentWidth

	// add whitespace if indented
	p.output = append(
		p.output,
		[]byte(strings.Repeat(" ", p.outputPos.Column-1))...,
	)
}

func (p *printer) insertSemi() {
	if p.ImpliedSemi {
		p.output = append(p.output, []byte(";")...)
		p.outputPos.Column++
	}
}

// Config control the output
type Config struct {
	ImpliedSemi bool // ImpliedSemi control end of statement semicolon.
	IndentWidth int
	NewlineChar []byte
}
