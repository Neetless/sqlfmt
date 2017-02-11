package scanner

import (
	"fmt"
	"path/filepath"
	"unicode"
	"unicode/utf8"

	"github.com/Neetless/sqlfmt/token"
)

const bom = 0xFEFF

// ErrorHandler may be provided to Scanner.Init. If a syntax error is
// encountered and a handler was installed, the handler is called with a
// position and an error message.
type ErrorHandler func(pos token.Position, msg string)

// Mode value is a set of flags (or 0).
// They control scanner behavior.
type Mode uint

const (
	// ScanComments return comments as COMMENT tokens.
	ScanComments Mode = 1 << iota
	dontInsertSemis
)

// Scanner holds the scanner's internal state while processing
// a given text.
type Scanner struct {
	file *token.File
	dir  string
	src  []byte
	err  ErrorHandler
	mode Mode

	ch         rune
	offset     int
	rdOffset   int
	lineOffset int
	insertSemi bool

	ErrorCount int
}

// Init prepares the scanner s to tokenize the text src by setting the
// scanner at the beginning of src. The scanner uses the file set file
// for position informaiton and it adds line information for each line.
func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len(%d)", file.Size(), len(src)))
	}
	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.src = src
	s.err = err
	s.mode = mode

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.insertSemi = false
	s.ErrorCount = 0

	s.next()
	if s.ch == bom {
		s.next()
	}
}

// Scan scans the next token and returns the token position, the token,
// and its literal string if applicable.
func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	// TODO: Temporary process
	tok = token.ILLEGAL
	lit = ""

	s.skipWhitespace()

	// current token start
	pos = s.file.Pos(s.offset)

	switch ch := s.ch; {
	case isLetter(ch):
		lit = s.scanIdentifier()
		if len(lit) > 1 {
			tok = token.Lookup(lit)
		} else {
			// insertSemi = true
			tok = token.IDENT
		}
	case ch == '\'':
		tok = token.STRING
		lit = s.scanString()
	default:
		s.next()
		switch ch {
		case -1:
			tok = token.EOF
		case '*':
			tok = token.MUL
			lit = "*"
		case '(':
			tok = token.LPAREN
			lit = "("
		case ')':
			tok = token.RPAREN
			lit = ")"
		case ';':
			tok = token.SEMICOLON
			lit = ";"
		case ',':
			tok = token.COMMA
			lit = ","
		case '=':
			tok = token.EQL
			lit = "="
		case '>':
			// TODO peak next ch and check >= case.
			tok = token.GTR
			lit = ">"
		case '.':
			tok = token.PERIOD
			lit = "."
		}
	}
	return
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' && !s.insertSemi || s.ch == '\r' {
		s.next()
	}
}

func (s *Scanner) scanString() string {
	offs := s.offset
	s.next()
	for s.ch != '\'' {
		if s.ch == -1 {
			panic("right single quote couldn't be found while scanning string")
		}
		s.next()
	}
	s.next()
	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

// Read the next Unicode char into s.ch.
// s.ch < 0 means end-of-file.
//
func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		s.rdOffset += w
		s.ch = r
	} else {
		// Set offset as last position.
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = -1
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}
