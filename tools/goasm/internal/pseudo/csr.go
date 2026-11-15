package pseudo

import (
	"github.com/natoga32/goasm/internal/ast"
)

// CSR pseudo-instructions

// csrr rd, csr - Read CSR
func (e *Expander) expandCsrr(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("csrrs", i.Operands[0], i.Operands[1], x(0)),
	}
}

// csrw csr, rs - Write CSR
func (e *Expander) expandCsrw(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("csrrw", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrs csr, rs - Set bits in CSR
func (e *Expander) expandCsrs(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("csrrs", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrc csr, rs - Clear bits in CSR
func (e *Expander) expandCsrc(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("csrrc", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrwi csr, imm - Write CSR immediate
func (e *Expander) expandCsrwi(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("csrrwi", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrsi csr, imm - Set bits in CSR immediate
func (e *Expander) expandCsrsi(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("csrrsi", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrci csr, imm - Clear bits in CSR immediate
func (e *Expander) expandCsrci(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("csrrci", x(0), i.Operands[0], i.Operands[1]),
	}
}
