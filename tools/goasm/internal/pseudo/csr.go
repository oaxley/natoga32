package pseudo

import (
	"github.com/natoga32/goasm/internal/arch"
	"github.com/natoga32/goasm/internal/ast"
)

// CSR pseudo-instructions, expanded to the standard order "op rd, csr, rs1/uimm"

// csrr rd, csr - Read CSR
func (e *Expander) expandCsrr(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("csrrs", i.Operands[0], i.Operands[1], x(0)),
	}
}

// csrw csr, rs - Write CSR
func (e *Expander) expandCsrw(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("csrrw", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrs csr, rs - Set bits in CSR
func (e *Expander) expandCsrs(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("csrrs", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrc csr, rs - Clear bits in CSR
func (e *Expander) expandCsrc(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("csrrc", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrwi csr, imm - Write CSR immediate
func (e *Expander) expandCsrwi(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("csrrwi", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrsi csr, imm - Set bits in CSR immediate
func (e *Expander) expandCsrsi(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("csrrsi", x(0), i.Operands[0], i.Operands[1]),
	}
}

// csrci csr, imm - Clear bits in CSR immediate
func (e *Expander) expandCsrci(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("csrrci", x(0), i.Operands[0], i.Operands[1]),
	}
}

// rdcycle rd (and rdcycleh, rdtime, rdtimeh, rdinstret, rdinstreth) - Read a counter
func (e *Expander) expandCounter(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("csrrs", i.Operands[0], num(arch.CounterCSRs[i.Opcode]), x(0)),
	}
}

// unimp - csrrw x0, cycle, x0: writes a read-only CSR, so it traps
func (e *Expander) expandUnimp() []*ast.Instruction {
	return []*ast.Instruction{
		instr("csrrw", x(0), num(arch.UnimpWord>>20), x(0)),
	}
}
