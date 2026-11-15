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
		// Could be either:
		// 1. Parenthesized expression: (a + b)
		// 2. Zero-offset shorthand: (sp) equivalent to 0(sp)

		// Peek ahead to check if this is just (identifier)
		if p.check(token.IDENT) {
			baseTok := p.peek()
			// Look ahead one more to see if it's followed by )
			savedPos := p.pos
			p.advance() // consume identifier

			if p.check(token.RPAREN) {
				// This is (register) shorthand for 0(register)
				p.advance() // consume )
				offset := &ast.Number{Value: 0, Location: loc}
				base := &ast.Identifier{Name: baseTok.Value, Location: locationFromToken(baseTok)}
				return &ast.OffsetBase{
					Offset:   offset,
					Base:     base,
					Location: loc,
				}, nil
			}

			// Not a shorthand, restore position and parse as normal expression
			p.pos = savedPos
		}

		// Regular parenthesized expression
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

// parseOffsetBase parses the (base) part of an offset(base) expression
// Example: 4(sp), label(a0), (sp)
func (p *Parser) parseOffsetBase(offset ast.Expression, loc ast.SourceLocation) (ast.Expression, error) {
	// Consume the '('
	p.advance()

	// Parse the base register (must be an identifier)
	if !p.check(token.IDENT) {
		return nil, fmt.Errorf("expected register identifier in offset(base) expression")
	}

	baseTok := p.advance()
	base := &ast.Identifier{Name: baseTok.Value, Location: locationFromToken(baseTok)}

	// Expect closing ')'
	if !p.check(token.RPAREN) {
		return nil, fmt.Errorf("expected ')' after base register in offset(base) expression")
	}
	p.advance()

	return &ast.OffsetBase{
		Offset:   offset,
		Base:     base,
		Location: loc,
	}, nil
}
