package pseudo

import (
	"github.com/natoga32/goasm/internal/ast"
)

// ret - Return from function
func (e *Expander) expandRet() []*ast.Instruction {
	return []*ast.Instruction{
		instr("jalr", x(0), x(1), zero()),
	}
}

// call symbol - Call function
func (e *Expander) expandCall(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 1 {
		return []*ast.Instruction{i}
	}

	symbol := i.Operands[0]

	return []*ast.Instruction{
		instr("auipc", x(1), &ast.PCRelHi{Expr: symbol}),
		instr("jalr", x(1), x(1), &ast.PCRelLo{Expr: symbol}),
	}
}

// j offset - Unconditional jump
func (e *Expander) expandJ(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 1 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("jal", x(0), i.Operands[0]),
	}
}

// jr rs - Jump register (using ra)
func (e *Expander) expandJr(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 1 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("jalr", x(0), i.Operands[0], zero()),
	}
}
