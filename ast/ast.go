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
	Begin   token.Pos
	Select  SelectClause
	From    FromClause
	Where   WhereClause
	Groupby GroupbyClause
	Orderby OrderbyClause
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
	switch {
	case s.Groupby.Exists:
		return s.Groupby.End()
	case s.Where.Exists:
		return s.Where.End()
	default:
		return s.From.End()
	}
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

// GroupbyClause represents where clause node.
type GroupbyClause struct {
	Node
	Begin  token.Pos
	ByPos  token.Pos
	Groups []Expr
	Exists bool
}

func (g GroupbyClause) clauseNode() {}

// Pos is implementation of Node interface.
func (g GroupbyClause) Pos() token.Pos {
	if !g.Exists {
		return 0
	}
	return g.Begin
}

// End is implementation of Node interface.
func (g GroupbyClause) End() token.Pos {
	if !g.Exists {
		return 0
	}
	if len(g.Groups) == 0 {
		panic("Groupby must have 1 or more groups.")
	}
	return g.Groups[len(g.Groups)-1].End()
}

// OrderbyClause represents where clause node.
type OrderbyClause struct {
	Node
	Begin  token.Pos
	ByPos  token.Pos
	Orders []Expr
	Exists bool
}

func (o OrderbyClause) clauseNode() {}

// Pos is implementation of Node interface.
func (o OrderbyClause) Pos() token.Pos {
	if !o.Exists {
		return 0
	}
	return o.Begin
}

// End is implementation of Node interface.
func (o OrderbyClause) End() token.Pos {
	if !o.Exists {
		return 0
	}
	if len(o.Orders) == 0 {
		panic("Orderby must have 1 or more groups.")
	}
	return o.Orders[len(o.Orders)-1].End()
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

// IsNullExpr represent is null expression.
type IsNullExpr struct {
	Value   Expr
	IsPos   token.Pos
	NullPos token.Pos
}

func (i IsNullExpr) exprNode() {}

// Pos returns initial position.
func (i IsNullExpr) Pos() token.Pos {
	return i.Value.Pos()
}

// End returns last position.
func (i IsNullExpr) End() token.Pos {
	return i.NullPos + token.Pos(len(token.NULL.String()))
}

// CaseExpr represent case expression.
// select case code when '0' then '1' else '2' end from tbl
type CaseExpr struct {
	Begin        token.Pos
	HasSwitchKey bool
	SwitchKey    Expr
	Whens        []*WhenClause
	Else         ElseClause
	EndPos       token.Pos
}

func (c CaseExpr) exprNode() {}

// Pos returns initial position.
func (c CaseExpr) Pos() token.Pos {
	return c.Begin
}

// End returns last position.
func (c CaseExpr) End() token.Pos {
	return c.EndPos + token.Pos(len(token.END.String()))
}

// WhenClause represents when part of case expression.
type WhenClause struct {
	Begin      token.Pos
	CondExpr   Expr
	ThenPos    token.Pos
	ResultExpr Expr
}

// Pos returns initial position.
func (w WhenClause) Pos() token.Pos {
	return w.Begin
}

// End returns last position.
func (w WhenClause) End() token.Pos {
	return w.ResultExpr.End()
}

// ElseClause represents else part of case expression.
type ElseClause struct {
	Begin      token.Pos
	ResultExpr Expr
	Exists     bool
}

// Pos returns initial position.
func (e ElseClause) Pos() token.Pos {
	if !e.Exists {
		return 0
	}
	return e.Begin
}

// End returns last position.
func (e ElseClause) End() token.Pos {
	if !e.Exists {
		return 0
	}
	return e.ResultExpr.End()
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
