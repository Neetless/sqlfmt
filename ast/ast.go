package ast

import "github.com/Neetless/sqlfmt/token"

// Node is a base interface which gives position information.
type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

// Stmt is an interface for statement.
type Stmt interface {
	Node
	stmtNode()
}

// DataMnpltStmt represents data manipulate statement.
type DataMnpltStmt struct {
	Stmt
}

// SelectStmt represents a select statement.
type SelectStmt struct {
	Begin  token.Pos
	Select SelectClause
	From   FromClause
	Where  WhereClause
}

func (s SelectStmt) stmtNode() {
}

// Pos is implementation for Node interface.
func (s SelectStmt) Pos() token.Pos {
	return s.Begin
}

// End is implmentation for Node interface.
func (s SelectStmt) End() token.Pos {
	//TODO Last component for SelectStmt is not determined.
	if s.Where.Exists {
		return s.Where.End()
	}
	return s.From.End()
}

// Clause represents any clause node.
type Clause interface {
	Node
	clauseNode()
}

// SelectClause represents select clause for sql.
type SelectClause struct {
	Begin token.Pos
	Cols  []*Column
}

func (s SelectClause) clauseNode() {}

// Pos is implementation for Node interface.
func (s SelectClause) Pos() token.Pos {
	return s.Begin
}

// End is implementation for Node interface.
func (s SelectClause) End() token.Pos {
	if len(s.Cols) == 0 {
		panic("Empty Cols is invalid.")
	}
	return s.Cols[len(s.Cols)-1].End()
}

// FromClause represents from clause in sql.
type FromClause struct {
	Begin  token.Pos
	Tables []*Table
}

func (f FromClause) clauseNode() {}

// Pos is implementation for Node interface.
func (f FromClause) Pos() token.Pos {
	return f.Begin
}

// End is implementation for Node interface.
func (f FromClause) End() token.Pos {
	if len(f.Tables) == 0 {
		panic("from clause contains no table.")
	}
	return f.Tables[len(f.Tables)-1].End()
}

// WhereClause represents where clause node.
type WhereClause struct {
	Node
	Begin    token.Pos
	CondExpr Expr
	Exists   bool
}

func (w WhereClause) clauseNode() {}

// Pos is implementation of Node interface.
func (w WhereClause) Pos() token.Pos {
	if !w.Exists {
		return 0
	}
	return w.Begin
}

// End is implementation of Node interface.
func (w WhereClause) End() token.Pos {
	if !w.Exists {
		return 0
	}
	return w.CondExpr.End()
}

// Table contains a table factors.
type Table struct {
	Value  TableExpr
	Alias  string
	EndPos token.Pos
}

// Pos returns the first position.
func (t Table) Pos() token.Pos {
	return t.Value.Pos()
}

// End returns the last position.
func (t Table) End() token.Pos {
	return t.EndPos
}

// TableExpr represents a table expression in from clause.
type TableExpr interface {
	Node
	tableExprNode()
}

// TableBasicLit represents one table name.
type TableBasicLit struct {
	Begin token.Pos
	Kind  token.Token
	Name  string
}

func (t TableBasicLit) tableExprNode() {}

// Pos is for Nore interface implementetion.
func (t TableBasicLit) Pos() token.Pos {
	return t.Begin
}

// End is for Node interface implementation.
func (t TableBasicLit) End() token.Pos {
	return t.Begin + token.Pos(len(t.Name))
}

// Column represents a column of table.
type Column struct {
	Node
	Value  Expr
	Alias  string
	EndPos token.Pos
}

// Pos returns the position of a column.
func (c Column) Pos() token.Pos {
	return c.Value.Pos()
}

// End returns the end postion of a column.
func (c Column) End() token.Pos {
	return c.EndPos
}

// Expr represent a expression in a column.
type Expr interface {
	Node
	exprNode()
}

// CallExpr represent function call expression.
type CallExpr struct {
	Begin    token.Pos
	FuncName string
	Lparen   token.Pos
	Args     []Expr
	Rparen   token.Pos
}

func (c CallExpr) exprNode() {}

// Pos returns initial position according to offset.
func (c CallExpr) Pos() token.Pos {
	return c.Begin
}

// End returns last position according to offset.
func (c CallExpr) End() token.Pos {
	return c.Rparen + 1
}

// BinaryExpr represents a binary expression.
type BinaryExpr struct {
	X     Expr
	OpPos token.Pos
	Op    token.Token
	Y     Expr
}

func (b BinaryExpr) exprNode() {}

// Pos implements Node interface.
func (b BinaryExpr) Pos() token.Pos {
	return b.X.Pos()
}

// End implements Node interface.
func (b BinaryExpr) End() token.Pos {
	return b.Y.End()
}

// UnaryExpr represents a unary expression.
type UnaryExpr struct {
	OpPos token.Pos
	Op    token.Token
	X     Expr
}

func (u UnaryExpr) exprNode() {}

// Pos implements Node interface.
func (u UnaryExpr) Pos() token.Pos {
	return u.OpPos
}

// End implements Node interface.
func (u UnaryExpr) End() token.Pos {
	return u.X.End()
}

// BasicLit represents a basic literal expression.
type BasicLit struct {
	Begin token.Pos
	Value string
	Kind  token.Token
}

func (b BasicLit) exprNode() {}

// Pos implement Node interface.
func (b BasicLit) Pos() token.Pos {
	return b.Begin
}

// End implement Node interface.
func (b BasicLit) End() token.Pos {
	return b.Begin + token.Pos(len(b.Value))
}

// Ident represents simple identifier for where clause.
type Ident struct {
	TblName string
	LitPos  token.Pos
	Kind    token.Token
	Lit     string
}

func (i Ident) exprNode() {}

// Pos implements Node interface.
func (i Ident) Pos() token.Pos {
	return i.LitPos
}

// End implements Node interface.
func (i Ident) End() token.Pos {
	return i.LitPos + token.Pos(len(i.Lit))
}
