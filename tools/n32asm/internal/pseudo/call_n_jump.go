package pseudo

import (
	"github.com/natoga32/n32asm/internal/ast"
)

// ret - Return from function
func (e *Expander) expandRet() []*ast.Instruction {
	return []*ast.Instruction{
		instr("jalr", x(0), x(1), zero()),
	}
}

// call symbol - Call function
func (e *Expander) expandCall(i *ast.Instruction) []*ast.Instruction {

	symbol := i.Operands[0]

	return []*ast.Instruction{
		instr("auipc", x(1), &ast.PCRelHi{Expr: symbol}),
		instr("jalr", x(1), x(1), &ast.PCRelLo{Expr: symbol}),
	}
}

// j offset - Unconditional jump
func (e *Expander) expandJ(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("jal", x(0), i.Operands[0]),
	}
}

// jr rs - Jump register (using ra)
func (e *Expander) expandJr(i *ast.Instruction) []*ast.Instruction {
	return []*ast.Instruction{
		instr("jalr", x(0), i.Operands[0], zero()),
	}
}

// tail symbol - Jump to a function without saving ra (t1 holds the address)
func (e *Expander) expandTail(i *ast.Instruction) []*ast.Instruction {
	symbol := i.Operands[0]
	return []*ast.Instruction{
		instr("auipc", x(6), &ast.PCRelHi{Expr: symbol}),
		instr("jalr", x(0), x(6), &ast.PCRelLo{Expr: symbol}),
	}
}
