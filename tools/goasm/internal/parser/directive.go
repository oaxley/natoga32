package parser

import (
	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/token"
)

// parseDirective parses a directive
func (p *Parser) parseDirective() (*ast.Directive, error) {
	dirTok := p.advance()

	args, err := p.parseOperands()
	if err != nil {
		return nil, err
	}

	p.expect(token.EOL)

	return &ast.Directive{
		Name: dirTok.Value,
		Args: args,
	}, nil
}
