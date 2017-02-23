package token

import (
	"strings"
	"sync"
)

// Token is the set of lexical tokens of the SQL language.
type Token int

// This const block define each tokens value.
const (
	ILLEGAL Token = iota

	IDENT
	INT
	ASTA
	STRING

	keywordBeg
	SELECT
	FROM
	WHERE
	GROUP
	ORDER
	BY
	ALIAS
	CASE
	WHEN
	THEN
	ELSE
	END
	NOT // not
	AND // and
	OR  // or
	IS
	NULL
	keywordEnd

	operatorBeg
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	EQL // =
	NEQ // !=
	GTR // >
	GEQ // >=
	LSS // <
	LEQ // <=
	LPAREN
	RPAREN
	SEMICOLON
	COMMA  // ,
	PERIOD // .
	operatorEnd

	EOF
	COMMENT
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	COMMENT: "COMMENT",

	IDENT:  "IDENT",
	INT:    "INT",
	STRING: "STRING",

	SELECT: "SELECT",
	FROM:   "FROM",
	WHERE:  "WHERE",
	GROUP:  "GROUP",
	ORDER:  "ORDER",
	BY:     "BY",
	ALIAS:  "AS",
	CASE:   "CASE",
	WHEN:   "WHEN",
	THEN:   "THEN",
	ELSE:   "ELSE",
	END:    "END",
	AND:    "AND",
	OR:     "OR",
	IS:     "IS",
	NULL:   "NULL",

	ASTA:      "*",
	ADD:       "+",
	SUB:       "-",
	MUL:       "*",
	QUO:       "/",
	EQL:       "=",
	GTR:       ">",
	LSS:       "<",
	LPAREN:    "(",
	RPAREN:    ")",
	SEMICOLON: ";",
	COMMA:     ",",
	PERIOD:    ".",
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keywordBeg + 1; i < keywordEnd; i++ {
		keywords[tokens[i]] = i
	}
}

func (t Token) String() string {
	return tokens[t]
}

// Lookup maps an identifier to its keyword token or IDENT(if not a keyword).
func Lookup(ident string) Token {
	if tok, isKeyword := keywords[strings.ToUpper(ident)]; isKeyword {
		return tok
	}
	return IDENT
}

// Pos is a compact encoding of a source position within a file set.
// It can be converted into a Position for a more convenient, but much
// larger, representation.
//
// The Pos value for a given file is a number in the range [base, base+size],
// where base and size are specified when adding the file to the file set via
// AddFile.
//
type Pos int

// File is a handle for a file belonging to a FileSet.
type File struct {
	set  *FileSet
	name string
	base int
	size int

	lines []int
	infos []lineInfo
}

// Name returns the file name of file f as registered  with AddFile.
func (f *File) Name() string {
	return f.name
}

// Size retruns the size of file f as registered with AddFile.
func (f *File) Size() int {
	return f.size
}

// Pos returns the Pos value for the given file offset;
// the offset must be <= f.Size()
// f.Pos(f.Offset(p)) == p.
//
func (f *File) Pos(offset int) Pos {
	if offset > f.size {
		panic("illegal file offset")
	}
	return Pos(f.base + offset)
}

// AddLine adds the line offset for a new line.
// The line offset must be larger than the offset for the previous line
// and smaller than the file size; otherwise the line offset is igonred.
//
func (f *File) AddLine(offset int) {
	f.set.mutex.Lock()
	if i := len(f.lines); (i == 0 || f.lines[i-1] < offset) && offset < f.size {
		f.lines = append(f.lines, offset)
	}
	f.set.mutex.Unlock()
}

type lineInfo struct {
	Offset   int
	Filename string
	Line     int
}

// FileSet represents a set of source files.
type FileSet struct {
	mutex sync.RWMutex
	base  int
	files []*File
	last  *File
}

// NewFileSet creates a new file set.
func NewFileSet() *FileSet {
	return &FileSet{
		base: 1,
	}
}

// Base returns the minimum base offset that must be provided to
// AddFile when adding the next file.
//
func (s *FileSet) Base() int {
	s.mutex.RLock()
	b := s.base
	s.mutex.RUnlock()
	return b
}

// AddFile adds a new file with a given filename, base offset, and file size
// to the file set s and returns the file. Multiple files may have the same name.
func (s *FileSet) AddFile(filename string, base, size int) *File {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if base < 0 {
		base = s.base
	}
	if base < s.base || size < 0 {
		panic("illegal base or size")
	}

	f := &File{s, filename, base, size, []int{0}, nil}
	base += size + 1
	if base < 0 {
		panic("token.Pos offset overflow (> 2G of source code in file set)")
	}
	s.base = base
	s.files = append(s.files, f)
	s.last = f
	return f
}

// Position describes an arbitrary source position
// including the file, line, and column location.
// A Position is valid if the line number is > 0.
type Position struct {
	Line   int
	Column int
}

// A set of constants for precedence-based expression parsing.
// Non-operators have lowest precedence, followed by operators
// starting with precedence 1 up to unary operators. The highest
// precedence serves as "catch-all" precedence for selector,
// indexing, and other operator and delimiter tokens.
//
const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
//
func (t Token) Precedence() int {
	switch t {
	case OR:
		return 1
	case AND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB:
		return 4
	case MUL, QUO, REM:
		return 5
	}
	return LowestPrec
}
