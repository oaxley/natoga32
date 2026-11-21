package main

import (
	"fmt"

	"github.com/natoga32/goasm/internal/arch"
	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

// encodeInstruction encodes a single instruction
func encodeInstruction(instr *ast.Instruction, symbols *symbol.Table, sections *section.Manager, ctx *encodeContext) error {
	op, ok := arch.LookupOpcode(instr.Opcode)
	if !ok {
		return fmt.Errorf("%s: error: unknown instruction '%s'", instr.Location, instr.Opcode)
	}

	pc := sections.CurrentAddress()
	var encoded []byte
	var err error

	if isCSRImmediate(instr.Opcode) {
		encoded, err = encodeCSRImmediate(op, instr, symbols, pc)
		if err != nil {
			return err
		}
		sections.Emit(encoded)
		return nil
	}

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
		// Unary R-type: clz, ctz, cpop, sext.b, sext.h, zext.h, rev8
		// Takes rd, rs1 only - rs2 is fixed by the opcode
		if len(instr.Operands) != 2 {
			return fmt.Errorf("%s: error: %s expects 2 operands (rd, rs1), got %d",
				instr.Location, instr.Opcode, len(instr.Operands))
		}
		rd := evalRegister(instr.Operands[0])
		rs1 := evalRegister(instr.Operands[1])
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
				imm, err = evalIImmediate(offsetBase.Offset, symbols, pc, ctx)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("%s: error: invalid operand format for %s (expected offset(base))", instr.Location, instr.Opcode)
			}
		} else if len(instr.Operands) >= 3 {
			// Legacy syntax: lw rd, rs1, offset
			rs1 = evalRegister(instr.Operands[1])
			imm, err = evalIImmediate(instr.Operands[2], symbols, pc, ctx)
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
			// the shift amount is a number, never a register
			if ident, ok := instr.Operands[2].(*ast.Identifier); ok && arch.IsRegister(ident.Name) {
				return fmt.Errorf("%s: error: %s shift amount must be a number (0..31), not register %s",
					instr.Location, instr.Opcode, ident.Name)
			}
			val, err := evalImmediateWithError(instr.Operands[2], symbols, pc)
			if err != nil {
				return err
			}
			if val < 0 || val > 31 {
				return fmt.Errorf("%s: error: %s shift amount %d out of range (0..31)", instr.Location, instr.Opcode, val)
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
				imm, err = evalIImmediate(offsetBase.Offset, symbols, pc, ctx)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("%s: error: invalid operand format for %s (expected offset(base))", instr.Location, instr.Opcode)
			}
		} else if len(instr.Operands) >= 3 {
			// Legacy syntax: sw rs2, offset, rs1
			imm, err = evalIImmediate(instr.Operands[1], symbols, pc, ctx)
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

	case arch.TypeSystem:
		// ecall, ebreak, wfi, mret, fence.i: the whole upper field is fixed by the opcode
		if len(instr.Operands) != 0 {
			return fmt.Errorf("%s: error: %s takes no operands", instr.Location, instr.Opcode)
		}
		encoded = arch.EncodeTypeSystem(op)

	case arch.TypeFence:
		// fence pred, succ; fence alone is fence iorw, iorw
		pred, succ := arch.FenceAll, arch.FenceAll
		switch len(instr.Operands) {
		case 0:
		case 2:
			if pred, err = evalFenceSet(instr, instr.Operands[0]); err != nil {
				return err
			}
			if succ, err = evalFenceSet(instr, instr.Operands[1]); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s: error: %s expects 0 or 2 operands (pred, succ), got %d",
				instr.Location, instr.Opcode, len(instr.Operands))
		}
		encoded = arch.EncodeTypeFence(op, pred, succ)

	default:
		return fmt.Errorf("%s: error: unsupported instruction type '%v'", instr.Location, op.Type)
	}

	sections.Emit(encoded)
	return nil
}

// evalFenceSet reads a fence set: letters among i, o, r, w (such as rw), or 0 for the empty set
func evalFenceSet(instr *ast.Instruction, operand ast.Expression) (uint32, error) {
	text := ""
	switch e := operand.(type) {
	case *ast.Identifier:
		text = e.Name
	case *ast.Number:
		text = fmt.Sprint(e.Value)
	}
	set, ok := arch.ParseFenceSet(text)
	if !ok {
		if text == "" {
			return 0, fmt.Errorf("%s: error: invalid fence set (expected letters among i, o, r, w, or 0)",
				instr.Location)
		}
		return 0, fmt.Errorf("%s: error: invalid fence set '%s' (expected letters among i, o, r, w, or 0)",
			instr.Location, text)
	}
	return set, nil
}

// isCSRImmediate reports whether mnemonic is a CSR immediate instruction
func isCSRImmediate(mnemonic string) bool {
	switch mnemonic {
	case "csrrwi", "csrrsi", "csrrci":
		return true
	}
	return false
}

// encodeCSRImmediate encodes "csrrXi rd, uimm, csr": the 5-bit uimm goes in
// the rs1 field and the CSR number in the I-type immediate field.
// The uimm can be an expression or, as emitted by the disassembler, a
// register name whose number is the uimm (e.g. x5 for 5).
func encodeCSRImmediate(op arch.InstrOpcode, instr *ast.Instruction, symbols *symbol.Table, pc int64) ([]byte, error) {
	if len(instr.Operands) != 3 {
		return nil, fmt.Errorf("%s: error: %s expects 3 operands (rd, uimm, csr)", instr.Location, instr.Opcode)
	}
	rd := evalRegister(instr.Operands[0])

	var uimm int32
	if ident, ok := instr.Operands[1].(*ast.Identifier); ok && arch.IsRegister(ident.Name) {
		uimm = int32(evalRegister(ident))
	} else {
		val, err := evalImmediateWithError(instr.Operands[1], symbols, pc)
		if err != nil {
			return nil, err
		}
		uimm = val
	}
	if uimm < 0 || uimm > 31 {
		return nil, fmt.Errorf("%s: error: %s immediate %d out of range (0..31)", instr.Location, instr.Opcode, uimm)
	}

	csr, err := evalImmediateWithError(instr.Operands[2], symbols, pc)
	if err != nil {
		return nil, err
	}
	if csr < 0 || csr > 0xFFF {
		return nil, fmt.Errorf("%s: error: CSR number %#x out of range (0..0xfff)", instr.Location, csr)
	}

	return arch.EncodeTypeI(op, rd, uint32(uimm), csr), nil
}
