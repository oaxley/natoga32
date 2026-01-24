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
		return fmt.Errorf("%s: error: unknown instruction '%s'", instr.Location, instr.Opcode)
	}

	pc := sections.CurrentAddress()
	var encoded []byte
	var err error

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

	case arch.TypeR2:
		rs1, rs2 := uint32(0), uint32(0)
		if len(instr.Operands) >= 1 {
			rs1 = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			rs2 = evalRegister(instr.Operands[1])
		}
		encoded = arch.EncodeTypeR2(op, rs1, rs2)

	case arch.TypeRU:
		// Unary R-type: clz, ctz, cpop, sext.b, sext.h, rev8
		// Takes rd, rs1 only - rs2 is fixed by the opcode
		rd, rs1 := uint32(0), uint32(0)
		if len(instr.Operands) >= 1 {
			rd = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			rs1 = evalRegister(instr.Operands[1])
		}
		encoded = arch.EncodeTypeRU(op, rd, rs1)

	case arch.TypeI:
		rd, rs1 := uint32(0), uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			rd = evalRegister(instr.Operands[0])
		}

		// Check for standard RISC-V syntax: lw rd, offset(rs1)
		if len(instr.Operands) == 2 {
			if offsetBase, ok := instr.Operands[1].(*ast.OffsetBase); ok {
				rs1 = evalRegister(offsetBase.Base)
				imm, err = evalImmediateWithError(offsetBase.Offset, symbols, pc)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("%s: error: invalid operand format for %s (expected offset(base))", instr.Location, instr.Opcode)
			}
		} else if len(instr.Operands) >= 3 {
			// Legacy syntax: lw rd, rs1, offset
			rs1 = evalRegister(instr.Operands[1])
			imm, err = evalImmediateWithError(instr.Operands[2], symbols, pc)
			if err != nil {
				return err
			}
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
			val, err := evalImmediateWithError(instr.Operands[2], symbols, pc)
			if err != nil {
				return err
			}
			shamt = uint32(val)
		}
		encoded = arch.EncodeTypeI2(op, rd, rs1, shamt)

	case arch.TypeS:
		rs1, rs2 := uint32(0), uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			rs2 = evalRegister(instr.Operands[0])
		}

		// Check for standard RISC-V syntax: sw rs2, offset(rs1)
		if len(instr.Operands) == 2 {
			if offsetBase, ok := instr.Operands[1].(*ast.OffsetBase); ok {
				rs1 = evalRegister(offsetBase.Base)
				imm, err = evalImmediateWithError(offsetBase.Offset, symbols, pc)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("%s: error: invalid operand format for %s (expected offset(base))", instr.Location, instr.Opcode)
			}
		} else if len(instr.Operands) >= 3 {
			// Legacy syntax: sw rs2, offset, rs1
			imm, err = evalImmediateWithError(instr.Operands[1], symbols, pc)
			if err != nil {
				return err
			}
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
			target, err := evalExpressionWithError(instr.Operands[2], symbols, pc)
			if err != nil {
				return err
			}
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
			imm, err = evalImmediateWithError(instr.Operands[1], symbols, pc)
			if err != nil {
				return err
			}
		}
		encoded = arch.EncodeTypeU(op, rd, imm)

	case arch.TypeJ:
		rd := uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			rd = evalRegister(instr.Operands[0])
		}
		if len(instr.Operands) >= 2 {
			target, err := evalExpressionWithError(instr.Operands[1], symbols, pc)
			if err != nil {
				return err
			}
			imm = int32(target - pc)
		}
		encoded = arch.EncodeTypeJ(op, rd, imm)

	default:
		return fmt.Errorf("%s: error: unsupported instruction type '%v'", instr.Location, op.Type)
	}

	sections.Emit(encoded)
	return nil
}
