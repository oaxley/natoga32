package parser

import (
	"github.com/natoga32/n32asm/internal/ast"
	"github.com/natoga32/n32asm/internal/token"
)

// parseStatement parses a single statement
func (p *Parser) parseStatement() (ast.Statement, error) {
	tok := p.peek()

	switch tok.Type {
	case token.EOF:
		return nil, nil

	case token.EOL:
		p.advance()
		return nil, nil

	case token.LABEL:
		return p.parseLabel()

	case token.IDENT:
		return p.parseInstruction()

	case token.DIRECTIVE:
		return p.parseDirective()
	}

	return nil, tokenError(tok, "unexpected %s at the start of a line", describeToken(tok))
}
