package pseudo

import (
	"github.com/natoga32/goasm/internal/ast"
)

// mv rd, rs - Move register
func (e *Expander) expandMv(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("addi", i.Operands[0], i.Operands[1], zero()),
	}
}

// nop - No operation
func (e *Expander) expandNop() []*ast.Instruction {
	return []*ast.Instruction{
		instr("addi", x(0), x(0), zero()),
	}
}

// not rd, rs - Bitwise NOT
func (e *Expander) expandNot(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("xori", i.Operands[0], i.Operands[1], num(-1)),
	}
}

// neg rd, rs - Negate
func (e *Expander) expandNeg(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("sub", i.Operands[0], x(0), i.Operands[1]),
	}
}

// syscall n - System call
func (e *Expander) expandSyscall(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 1 {
		return []*ast.Instruction{i}
	}

	param := i.Operands[0]

	// If it's a number, load it into a17 (x17)
	if _, ok := param.(*ast.Number); ok {
		return []*ast.Instruction{
			instr("addi", x(17), x(0), param),
			instr("ecall"),
		}
	}

	// If it's a register, move it to a17
	return []*ast.Instruction{
		instr("add", x(17), x(0), param),
		instr("ecall"),
	}
}
