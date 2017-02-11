package ast

import (
	"testing"
)

func TestAst(t *testing.T) {
	s := SelectStmt{
		Begin:  1,
		Select: SelectClause{},
		From:   FromClause{},
	}
	t.Logf("%T\n", s)
}
