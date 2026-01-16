// Lexer for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package lexer

import (
	"bufio"
	"io"
	"strings"

	"github.com/natoga32/goasm/internal/token"
)

// Lexer tokenizes assembly source code
type Lexer struct {
	input  string
	pos    int // Current position in input
	row    int // Current line number (1-based)
	col    int // Current column (1-based)
	tokens []token.Token
}

// New creates a new lexer from a string
func New(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		row:    1,
		col:    1,
		tokens: make([]token.Token, 0),
	}
}

// NewFromReader creates a new lexer from an io.Reader
func NewFromReader(r io.Reader) (*Lexer, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return New(sb.String()), nil
}

// Tokenize processes the input and returns all tokens
func (l *Lexer) Tokenize() []token.Token {
	for !l.atEnd() {
		l.scanToken()
	}
	l.tokens = append(l.tokens, token.New(token.EOF, "", l.row, l.col))
	return l.tokens
}

// atEnd returns true if we've reached the end of input
func (l *Lexer) atEnd() bool {
	return l.pos >= len(l.input)
}

// peek returns the current character without advancing
func (l *Lexer) peek() byte {
	if l.atEnd() {
		return 0
	}
	return l.input[l.pos]
}

// peekNext returns the next character without advancing
func (l *Lexer) peekNext() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

// advance consumes and returns the current character
func (l *Lexer) advance() byte {
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.row++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

// match checks if the current character matches expected, and advances if so
func (l *Lexer) match(expected byte) bool {
	if l.atEnd() || l.peek() != expected {
		return false
	}
	l.advance()
	return true
}

// addToken adds a token to the list
func (l *Lexer) addToken(typ token.Type, value string, row, col int) {
	l.tokens = append(l.tokens, token.New(typ, value, row, col))
}
