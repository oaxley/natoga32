package main

import (
	"fmt"

	"github.com/natoga32/goasm/internal/arch"
	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

// encodeInstruction encodes a single instruction
func encodeInstruction(instr *ast.Instruction, symbols *symbol.Table, sections *section.Manager) error {
	op, ok := arch.LookupOpcode(instr.Opcode)
	if !ok {
		return fmt.Errorf("unknown instruction: %s", instr.Opcode)
	}

	pc := sections.CurrentAddress()
	var encoded []byte

	switch op.Type {
	case arch.TypeR:
		rd, rs1, rs2 := uint32(0), uint32(0), uint32(0)
		if len(instr.Operands) >= 1 {
			rd = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			rs1 = evalRegister(instr.Operands[1])
		}
		if len(instr.Operands) >= 3 {
			rs2 = evalRegister(instr.Operands[2])
		}
		encoded = arch.EncodeTypeR(op, rd, rs1, rs2)

	case arch.TypeI:
		rd, rs1 := uint32(0), uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			rd = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			rs1 = evalRegister(instr.Operands[1])
		}
		if len(instr.Operands) >= 3 {
			imm = evalImmediate(instr.Operands[2], symbols, pc)
		}
		encoded = arch.EncodeTypeI(op, rd, rs1, imm)

	case arch.TypeI2:
		rd, rs1 := uint32(0), uint32(0)
		shamt := uint32(0)
		if len(instr.Operands) >= 1 {
			rd = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			rs1 = evalRegister(instr.Operands[1])
		}
		if len(instr.Operands) >= 3 {
			shamt = uint32(evalImmediate(instr.Operands[2], symbols, pc))
		}
		encoded = arch.EncodeTypeI2(op, rd, rs1, shamt)

	case arch.TypeS:
		rs1, rs2 := uint32(0), uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			rs2 = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			imm = evalImmediate(instr.Operands[1], symbols, pc)
		}
		if len(instr.Operands) >= 3 {
			rs1 = evalRegister(instr.Operands[2])
		}
		encoded = arch.EncodeTypeS(op, rs1, rs2, imm)

	case arch.TypeB:
		rs1, rs2 := uint32(0), uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			rs1 = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			rs2 = evalRegister(instr.Operands[1])
		}
		if len(instr.Operands) >= 3 {
			target := evalExpression(instr.Operands[2], symbols, pc)
			imm = int32(target - pc)
		}
		encoded = arch.EncodeTypeB(op, rs1, rs2, imm)

	case arch.TypeU:
		rd := uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			rd = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			imm = evalImmediate(instr.Operands[1], symbols, pc)
		}
		encoded = arch.EncodeTypeU(op, rd, imm)

	case arch.TypeJ:
		rd := uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			rd = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			target := evalExpression(instr.Operands[1], symbols, pc)
			imm = int32(target - pc)
		}
		encoded = arch.EncodeTypeJ(op, rd, imm)

	default:
		return fmt.Errorf("unsupported instruction type: %v", op.Type)
	}

	sections.Emit(encoded)
	return nil
}
