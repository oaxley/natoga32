// Lexer for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package lexer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/natoga32/goasm/internal/token"
)

// Lexer tokenizes assembly source code
type Lexer struct {
	input    string
	filename string
	pos      int // Current position in input
	row      int // Current line number (1-based)
	col      int // Current column (1-based)
	tokens   []token.Token
	err      error // first error found by Tokenize
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

// NewWithFile creates a new lexer from a string with a filename
func NewWithFile(input, filename string) *Lexer {
	return &Lexer{
		input:    input,
		filename: filename,
		pos:      0,
		row:      1,
		col:      1,
		tokens:   make([]token.Token, 0),
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

// NewFromReaderWithFile creates a new lexer from an io.Reader with a filename
func NewFromReaderWithFile(r io.Reader, filename string) (*Lexer, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return NewWithFile(sb.String(), filename), nil
}

// Tokenize processes the input and returns all tokens
func (l *Lexer) Tokenize() []token.Token {
	for !l.atEnd() {
		l.scanToken()
	}
	l.tokens = append(l.tokens, token.NewWithFile(token.EOF, "", l.filename, l.row, l.col))
	return l.tokens
}

// Err returns the first error found by Tokenize (an unknown character), or nil
func (l *Lexer) Err() error {
	return l.err
}

// unexpected records an unknown character that starts at byte pos-1 (the
// first error only), and skips the rest of a multi-byte UTF-8 character
func (l *Lexer) unexpected(row, col int) {
	r, size := utf8.DecodeRuneInString(l.input[l.pos-1:])
	for i := 1; i < size; i++ {
		l.advance()
	}
	if l.err != nil {
		return
	}
	loc := fmt.Sprintf("%d:%d", row, col)
	if l.filename != "" {
		loc = l.filename + ":" + loc
	}
	if r == utf8.RuneError && size <= 1 {
		l.err = fmt.Errorf("%s: error: unexpected byte 0x%02X", loc, l.input[l.pos-1])
		return
	}
	l.err = fmt.Errorf("%s: error: unexpected character '%c'", loc, r)
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
	l.tokens = append(l.tokens, token.NewWithFile(typ, value, l.filename, row, col))
}
