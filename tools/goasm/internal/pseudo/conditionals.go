package pseudo

import (
	"github.com/natoga32/goasm/internal/ast"
)

// seqz rd, rs - Set if equal to zero
func (e *Expander) expandSeqz(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("sltiu", i.Operands[0], i.Operands[1], num(1)),
	}
}

// snez rd, rs - Set if not equal to zero
func (e *Expander) expandSnez(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("sltu", i.Operands[0], x(0), i.Operands[1]),
	}
}

// sltz rd, rs - Set if less than zero
func (e *Expander) expandSltz(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("slt", i.Operands[0], i.Operands[1], x(0)),
	}
}

// sgtz rd, rs - Set if greater than zero
func (e *Expander) expandSgtz(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("slt", i.Operands[0], x(0), i.Operands[1]),
	}
}
