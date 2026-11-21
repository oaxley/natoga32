package parser

import (
	"github.com/natoga32/n32asm/internal/ast"
	"github.com/natoga32/n32asm/internal/token"
)

// parseInstruction parses an instruction
func (p *Parser) parseInstruction() (*ast.Instruction, error) {
	opTok := p.advance()
	if opTok.Type != token.IDENT {
		return nil, tokenError(opTok, "expected an instruction, got %s", describeToken(opTok))
	}

	operands, err := p.parseOperands()
	if err != nil {
		return nil, err
	}

	p.expect(token.EOL)

	return &ast.Instruction{
		Opcode:   opTok.Value,
		Operands: operands,
		Location: locationFromToken(opTok),
	}, nil
}
