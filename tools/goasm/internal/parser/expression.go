package parser

import (
	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/token"
)

// parseExpression parses an expression with operator precedence
func (p *Parser) parseExpression(minPrec int) (ast.Expression, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()
		if tok.Type == token.EOF {
			break
		}

		// Check if this is an operator
		op := tokenToOperator(tok.Type)
		if op == "" {
			break
		}

		prec, ok := opPrecedence[op]
		if !ok || prec < minPrec {
			break
		}

		// Consume the operator
		opTok := p.advance()

		// Parse right side with higher precedence
		right, err := p.parseExpression(prec + 1)
		if err != nil {
			return nil, err
		}

		left = &ast.BinaryOp{
			Op:       op,
			Left:     left,
			Right:    right,
			Location: locationFromToken(opTok),
		}
	}

	// After parsing the complete expression, check for offset(base) syntax
	// This must happen AFTER all operators are parsed, so -4(sp) becomes
	// OffsetBase{UnaryOp{-, 4}, sp} not UnaryOp{-, OffsetBase{4, sp}}
	if p.check(token.LPAREN) {
		// Peek ahead to see if this is (identifier)
		savedPos := p.pos
		p.advance() // consume (

		if p.check(token.IDENT) {
			baseTok := p.peek()
			p.advance() // consume identifier

			if p.check(token.RPAREN) {
				// This is offset(base) syntax
				p.advance() // consume )
				base := &ast.Identifier{Name: baseTok.Value, Location: locationFromToken(baseTok)}
				return &ast.OffsetBase{
					Offset:   left,
					Base:     base,
					Location: locationFromToken(baseTok),
				}, nil
			}
		}

		// Not offset(base) syntax, restore position
		p.pos = savedPos
	}

	return left, nil
}

// tokenToOperator converts a token type to its operator string
func tokenToOperator(typ token.Type) string {
	switch typ {
	case token.PLUS:
		return "+"
	case token.MINUS:
		return "-"
	case token.STAR:
		return "*"
	case token.SLASH:
		return "/"
	case token.MODULO:
		return "%"
	case token.LSHIFT:
		return "<<"
	case token.RSHIFT:
		return ">>"
	case token.AND:
		return "&"
	case token.OR:
		return "|"
	case token.XOR:
		return "^"
	}
	return ""
}
