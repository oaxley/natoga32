package parser

import (
	"fmt"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/helpers"
	"github.com/natoga32/goasm/internal/token"
)

// parsePrimary parses a primary expression
func (p *Parser) parsePrimary() (ast.Expression, error) {
	tok := p.advance()
	loc := locationFromToken(tok)

	switch tok.Type {
	case token.EOF:
		return nil, fmt.Errorf("unexpected end of file in expression")

	case token.NUMBER:
		val, err := helpers.ParseNumber(tok.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q: %v", tok.Value, err)
		}
		return &ast.Number{Value: val, Location: loc}, nil

	case token.IDENT:
		return &ast.Identifier{Name: tok.Value, Location: loc}, nil

	case token.DOLLAR:
		return &ast.CurrentPC{Location: loc}, nil

	case token.STRING:
		// Strip quotes
		s := tok.Value
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
		}
		return &ast.StringLiteral{Text: s, Location: loc}, nil

	case token.CHAR:
		// Strip quotes and get character
		s := tok.Value
		if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
			s = s[1 : len(s)-1]
		}
		var ch rune = 0
		if len(s) > 0 {
			if s[0] == '\\' && len(s) > 1 {
				// Handle escape sequences
				switch s[1] {
				case 'n':
					ch = '\n'
				case 't':
					ch = '\t'
				case 'r':
					ch = '\r'
				case '0':
					ch = 0
				case '\\':
					ch = '\\'
				case '\'':
					ch = '\''
				default:
					ch = rune(s[1])
				}
			} else {
				ch = rune(s[0])
			}
		}
		return &ast.CharLiteral{Char: ch, Location: loc}, nil

	case token.ABS_LO, token.ABS_HI, token.PCREL_LO, token.PCREL_HI:
		// Relocation modifier: %lo(expr), %hi(expr), etc.
		if !p.check(token.LPAREN) {
			return nil, fmt.Errorf("expected '(' after relocation modifier")
		}
		p.advance() // consume '('

		expr, err := p.parseExpression(1)
		if err != nil {
			return nil, err
		}

		if p.check(token.RPAREN) {
			p.advance()
		}

		switch tok.Type {
		case token.ABS_LO:
			return &ast.LoRel{Expr: expr, Location: loc}, nil
		case token.ABS_HI:
			return &ast.HiRel{Expr: expr, Location: loc}, nil
		case token.PCREL_LO:
			return &ast.PCRelLo{Expr: expr, Location: loc}, nil
		case token.PCREL_HI:
			return &ast.PCRelHi{Expr: expr, Location: loc}, nil
		}

	case token.LPAREN:
		// Parenthesized expression
		expr, err := p.parseExpression(1)
		if err != nil {
			return nil, err
		}
		if p.check(token.RPAREN) {
			p.advance()
		}
		return expr, nil

	case token.PLUS, token.MINUS:
		// Unary operator
		op := "+"
		if tok.Type == token.MINUS {
			op = "-"
		}
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryOp{Op: op, Operand: operand, Location: loc}, nil
	}

	return nil, fmt.Errorf("unexpected token in expression: %s (%q) at line %d", tok.Type, tok.Value, tok.Row)
}
