package parser

import (
	"fmt"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/token"
)

// parseInstruction parses an instruction
func (p *Parser) parseInstruction() (*ast.Instruction, error) {
	opTok := p.advance()
	if opTok.Type != token.IDENT {
		return nil, fmt.Errorf("expected instruction opcode, got %s", opTok.Type)
	}

	operands, err := p.parseOperands()
	if err != nil {
		return nil, err
	}

	p.expect(token.EOL)

	return &ast.Instruction{
		Opcode:   opTok.Value,
		Operands: operands,
	}, nil
}
