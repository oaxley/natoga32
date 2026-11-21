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
		return nil, tokenError(tok, "expected a value at the end of the file")

	case token.NUMBER:
		val, err := helpers.ParseNumber(tok.Value)
		if err != nil {
			return nil, fmt.Errorf("%s: error: invalid number %q: %v", loc, tok.Value, err)
		}
		return &ast.Number{Value: val, Location: loc}, nil

	case token.IDENT:
		return &ast.Identifier{Name: tok.Value, Location: loc}, nil

	case token.DOLLAR:
		return &ast.CurrentPC{Location: loc}, nil

	case token.STRING:
		// Strip quotes, then translate the escapes
		s := tok.Value
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
		}
		text, err := helpers.Unescape(s)
		if err != nil {
			return nil, fmt.Errorf("%s: error: %v", loc, err)
		}
		return &ast.StringLiteral{Text: text, Location: loc}, nil

	case token.CHAR:
		ch, err := helpers.CharValue(tok.Value)
		if err != nil {
			return nil, fmt.Errorf("%s: error: %v", loc, err)
		}
		return &ast.CharLiteral{Char: rune(ch), Location: loc}, nil

	case token.NOT, token.LAND, token.LOR:
		return nil, fmt.Errorf("%s: error: '%s' can be used only in .if and .define", loc, tok.Value)

	case token.ABS_LO, token.ABS_HI, token.PCREL_LO, token.PCREL_HI:
		// Relocation modifier: %lo(expr), %hi(expr), etc.
		if !p.check(token.LPAREN) {
			return nil, tokenError(tok, "expected '(' after %s", tok.Value)
		}
		p.advance() // consume '('

		expr, err := p.parseExpression(1)
		if err != nil {
			return nil, err
		}

		if !p.expect(token.RPAREN) {
			return nil, fmt.Errorf("%s: error: missing ')' after %s(", loc, tok.Value)
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
		if !p.expect(token.RPAREN) {
			return nil, fmt.Errorf("%s: error: missing ')'", loc)
		}
		return expr, nil

	case token.PLUS, token.MINUS, token.TILDE:
		// Unary operator
		op := tok.Value
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryOp{Op: op, Operand: operand, Location: loc}, nil
	}

	if tok.Type == token.EOL {
		return nil, tokenError(tok, "expected a value at the end of the line")
	}
	return nil, tokenError(tok, "unexpected %s in an expression", describeToken(tok))
}
