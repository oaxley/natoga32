package parser

import (
	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/token"
)

// parseOperands parses a list of operands separated by commas. Exactly one
// comma separates two operands: a missing, extra, leading or trailing comma
// is an error.
func (p *Parser) parseOperands() ([]ast.Expression, error) {
	var operands []ast.Expression

	for {
		tok := p.peek()
		if tok.Type == token.EOL || tok.Type == token.EOF {
			return operands, nil
		}
		if tok.Type == token.COMMA {
			return nil, tokenError(tok, "expected a value before ','")
		}

		expr, err := p.parseExpression(1)
		if err != nil {
			return nil, err
		}
		operands = append(operands, expr)

		// after an operand: the end of the line, or ',' and another operand
		tok = p.peek()
		switch tok.Type {
		case token.EOL, token.EOF:
			return operands, nil
		case token.COMMA:
			p.advance()
			if next := p.peek(); next.Type == token.EOL || next.Type == token.EOF || next.Type == token.COMMA {
				return nil, tokenError(tok, "expected a value after ','")
			}
		case token.LAND, token.LOR:
			return nil, tokenError(tok, "'%s' can be used only in .if and .define", tok.Value)
		default:
			return nil, tokenError(tok, "expected ',' before %s", describeToken(tok))
		}
	}
}
