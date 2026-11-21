package pseudo

import (
	"github.com/natoga32/goasm/internal/ast"
)

// li rd, imm - Load immediate
func (e *Expander) expandLi(i *ast.Instruction) []*ast.Instruction {

	rd := i.Operands[0]
	imm := i.Operands[1]

	// A char literal is a plain number
	if ch, ok := imm.(*ast.CharLiteral); ok {
		imm = &ast.Number{Value: int64(ch.Char), Location: ch.Location}
	}

	// A number checks if it fits in 12 bits. A number outside -0x80000000 to
	// 0xFFFFFFFF takes the %hi/%lo form below, which the assembler reports
	// as out of range once it has the values (relaxLi).
	if num, ok := imm.(*ast.Number); ok && num.Value >= -0x80000000 && num.Value <= 0xFFFFFFFF {
		// 0x80000000..0xFFFFFFFF is a 32-bit pattern: li a0, 0xFFFFFFFF is li a0, -1
		if num.Value > 0x7FFFFFFF && num.Value <= 0xFFFFFFFF {
			num = &ast.Number{Value: int64(int32(uint32(num.Value))), Location: num.Location}
			imm = num
		}
		if num.Value >= -2048 && num.Value <= 2047 {
			// Small immediate: addi rd, x0, imm
			return []*ast.Instruction{
				instr("addi", rd, x(0), imm),
			}
		}

		// Large immediate: lui + addi
		upper := (num.Value + 0x800) >> 12
		lower := num.Value - (upper << 12)

		if lower == 0 {
			// the low 12 bits are 0: lui alone is enough
			return []*ast.Instruction{instr("lui", rd, &ast.Number{Value: upper})}
		}
		return []*ast.Instruction{
			instr("lui", rd, &ast.Number{Value: upper}),
			instr("addi", rd, rd, &ast.Number{Value: lower}),
		}
	}

	// Symbol or expression: %hi/%lo. Once the names have their values, the
	// assembler makes it one instruction when the value allows it (relaxLi).
	return []*ast.Instruction{
		instr("lui", rd, &ast.HiRel{Expr: imm}),
		instr("addi", rd, rd, &ast.LoRel{Expr: imm}),
	}
}

// la rd, symbol - Load address
func (e *Expander) expandLa(i *ast.Instruction) []*ast.Instruction {

	rd := i.Operands[0]
	symbol := i.Operands[1]

	return []*ast.Instruction{
		instr("auipc", rd, &ast.PCRelHi{Expr: symbol}),
		instr("addi", rd, rd, &ast.PCRelLo{Expr: symbol}),
	}
}
