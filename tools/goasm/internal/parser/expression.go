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
		p.advance()

		// Parse right side with higher precedence
		right, err := p.parseExpression(prec + 1)
		if err != nil {
			return nil, err
		}

		left = &ast.BinaryOp{
			Op:    op,
			Left:  left,
			Right: right,
		}
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
