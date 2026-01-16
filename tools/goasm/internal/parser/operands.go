package parser

import (
	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/token"
)

// parseOperands parses a list of operands separated by commas
func (p *Parser) parseOperands() ([]ast.Expression, error) {
	var operands []ast.Expression

	for {
		tok := p.peek()
		if tok.Type == token.EOL || tok.Type == token.EOF {
			break
		}

		// Skip commas
		if tok.Type == token.COMMA {
			p.advance()
			continue
		}

		expr, err := p.parseExpression(1)
		if err != nil {
			return nil, err
		}
		operands = append(operands, expr)
	}

	return operands, nil
}
