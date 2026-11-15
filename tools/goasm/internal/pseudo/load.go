package pseudo

import (
	"github.com/natoga32/goasm/internal/ast"
)

// li rd, imm - Load immediate
func (e *Expander) expandLi(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}

	rd := i.Operands[0]
	imm := i.Operands[1]

	// If immediate is a number, check if it fits in 12 bits
	if num, ok := imm.(*ast.Number); ok {
		if num.Value >= -2048 && num.Value <= 2047 {
			// Small immediate: addi rd, x0, imm
			return []*ast.Instruction{
				instr("addi", rd, x(0), imm),
			}
		}

		// Large immediate: lui + addi
		upper := (num.Value + 0x800) >> 12
		lower := num.Value - (upper << 12)

		return []*ast.Instruction{
			instr("lui", rd, &ast.Number{Value: upper}),
			instr("addi", rd, rd, &ast.Number{Value: lower}),
		}
	}

	// Symbol: use %hi/%lo relocations
	return []*ast.Instruction{
		instr("lui", rd, &ast.HiRel{Expr: imm}),
		instr("addi", rd, rd, &ast.LoRel{Expr: imm}),
	}
}

// la rd, symbol - Load address
func (e *Expander) expandLa(i *ast.Instruction) []*ast.Instruction {
	if len(i.Operands) < 2 {
		return []*ast.Instruction{i}
	}

	rd := i.Operands[0]
	symbol := i.Operands[1]

	return []*ast.Instruction{
		instr("auipc", rd, &ast.PCRelHi{Expr: symbol}),
		instr("addi", rd, rd, &ast.PCRelLo{Expr: symbol}),
	}
}
