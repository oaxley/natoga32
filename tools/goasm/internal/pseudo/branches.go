package pseudo

import (
	"github.com/natoga32/goasm/internal/ast"
)

// Branch pseudo-instructions (1 register + offset)

// beqz rs, offset - Branch if equal to zero
func (e *Expander) expandBeqz(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("beq", i.Operands[0], x(0), i.Operands[1]),
	}
}

// bnez rs, offset - Branch if not equal to zero
func (e *Expander) expandBnez(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("bne", i.Operands[0], x(0), i.Operands[1]),
	}
}

// blez rs, offset - Branch if less than or equal to zero
func (e *Expander) expandBlez(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("bge", x(0), i.Operands[0], i.Operands[1]),
	}
}

// bgez rs, offset - Branch if greater than or equal to zero
func (e *Expander) expandBgez(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("bge", i.Operands[0], x(0), i.Operands[1]),
	}
}

// bltz rs, offset - Branch if less than zero
func (e *Expander) expandBltz(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("blt", i.Operands[0], x(0), i.Operands[1]),
	}
}

// bgtz rs, offset - Branch if greater than zero
func (e *Expander) expandBgtz(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("blt", x(0), i.Operands[0], i.Operands[1]),
	}
}

// Branch pseudo-instructions (2 registers + offset)

// bgt rs, rt, offset - Branch if greater than
func (e *Expander) expandBgt(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 3 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("blt", i.Operands[1], i.Operands[0], i.Operands[2]),
	}
}

// ble rs, rt, offset - Branch if less than or equal
func (e *Expander) expandBle(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 3 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("bge", i.Operands[1], i.Operands[0], i.Operands[2]),
	}
}

// bgtu rs, rt, offset - Branch if greater than (unsigned)
func (e *Expander) expandBgtu(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 3 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("bltu", i.Operands[1], i.Operands[0], i.Operands[2]),
	}
}

// bleu rs, rt, offset - Branch if less than or equal (unsigned)
func (e *Expander) expandBleu(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 3 {
		return []*ast.Instruction{i}
	}
	return []*ast.Instruction{
		instr("bgeu", i.Operands[1], i.Operands[0], i.Operands[2]),
	}
}
