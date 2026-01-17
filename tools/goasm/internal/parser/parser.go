// Parser for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package parser

import (
	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/token"
)

// Operator precedence (higher = binds tighter)
var opPrecedence = map[string]int{
	"|":  1,
	"^":  2,
	"&":  3,
	"<<": 4,
	">>": 4,
	"+":  5,
	"-":  5,
	"*":  6,
	"/":  6,
	"%":  6,
}

// Parser parses tokens into an AST
type Parser struct {
	tokens []token.Token
	pos    int
}

// New creates a new parser
func New(tokens []token.Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

// Parse parses the token stream and returns the AST
func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{
		Statements: make([]ast.Statement, 0),
	}

	for !p.atEnd() {
		// Skip standalone EOL
		if p.check(token.EOL) {
			p.advance()
			continue
		}

		// Stop at EOF
		if p.check(token.EOF) {
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.AddStatement(stmt)
		}
	}

	return program, nil
}

// atEnd returns true if we've reached the end of tokens
func (p *Parser) atEnd() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Type == token.EOF
}

// peek returns the current token without advancing
func (p *Parser) peek() token.Token {
	if p.pos >= len(p.tokens) {
		return token.Token{Type: token.EOF}
	}
	return p.tokens[p.pos]
}

// advance consumes and returns the current token
func (p *Parser) advance() token.Token {
	tok := p.peek()
	if !p.atEnd() {
		p.pos++
	}
	return tok
}

// check returns true if the current token has the given type
func (p *Parser) check(typ token.Type) bool {
	return p.peek().Type == typ
}

// expect returns true if the current token matches, and advances
func (p *Parser) expect(typ token.Type) bool {
	if p.check(typ) {
		p.advance()
		return true
	}
	return false
}

// locationFromToken creates a SourceLocation from a token
func locationFromToken(tok token.Token) ast.SourceLocation {
	return ast.SourceLocation{
		Filename: tok.Filename,
		Line:     tok.Row,
		Column:   tok.Col,
	}
}
