package main

import (
	"fmt"
	"strings"

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

	if isCSR(instr.Opcode) {
		encoded, err = encodeCSR(op, instr, symbols, pc)
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
			if rd, err = evalRegister(instr, instr.Operands[0]); err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 2 {
			if rs1, err = evalRegister(instr, instr.Operands[1]); err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 3 {
			if rs2, err = evalRegister(instr, instr.Operands[2]); err != nil {
				return err
			}
		}
		encoded = arch.EncodeTypeR(op, rd, rs1, rs2)

	case arch.TypeR2:
		rs1, rs2 := uint32(0), uint32(0)
		if len(instr.Operands) >= 1 {
			if rs1, err = evalRegister(instr, instr.Operands[0]); err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 2 {
			if rs2, err = evalRegister(instr, instr.Operands[1]); err != nil {
				return err
			}
		}
		encoded = arch.EncodeTypeR2(op, rs1, rs2)

	case arch.TypeRU:
		// Unary R-type: clz, ctz, cpop, sext.b, sext.h, zext.h, rev8
		// Takes rd, rs1 only - rs2 is fixed by the opcode
		if len(instr.Operands) != 2 {
			return fmt.Errorf("%s: error: %s expects 2 operands (rd, rs1), got %d",
				instr.Location, instr.Opcode, len(instr.Operands))
		}
		rd, err := evalRegister(instr, instr.Operands[0])
		if err != nil {
			return err
		}
		rs1, err := evalRegister(instr, instr.Operands[1])
		if err != nil {
			return err
		}
		encoded = arch.EncodeTypeRU(op, rd, rs1)

	case arch.TypeI:
		rd, rs1 := uint32(0), uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			if rd, err = evalRegister(instr, instr.Operands[0]); err != nil {
				return err
			}
		}

		// Check for standard RISC-V syntax: lw rd, offset(rs1)
		if len(instr.Operands) == 2 {
			if offsetBase, ok := instr.Operands[1].(*ast.OffsetBase); ok {
				if rs1, err = evalRegister(instr, offsetBase.Base); err != nil {
					return err
				}
				imm, err = evalIImmediate(offsetBase.Offset, symbols, pc, ctx)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("%s: error: invalid operand format for %s (expected offset(base))", instr.Location, instr.Opcode)
			}
		} else if len(instr.Operands) >= 3 {
			// Legacy syntax: lw rd, rs1, offset
			if rs1, err = evalRegister(instr, instr.Operands[1]); err != nil {
				return err
			}
			imm, err = evalIImmediate(instr.Operands[2], symbols, pc, ctx)
			if err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 2 {
			immExpr := instr.Operands[len(instr.Operands)-1]
			if ob, ok := immExpr.(*ast.OffsetBase); ok {
				immExpr = ob.Offset
			}
			if err := checkIImmediate(instr, immExpr, imm); err != nil {
				return err
			}
		}
		encoded = arch.EncodeTypeI(op, rd, rs1, imm)

	case arch.TypeI2:
		rd, rs1 := uint32(0), uint32(0)
		shamt := uint32(0)
		if len(instr.Operands) >= 1 {
			if rd, err = evalRegister(instr, instr.Operands[0]); err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 2 {
			if rs1, err = evalRegister(instr, instr.Operands[1]); err != nil {
				return err
			}
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
			if rs2, err = evalRegister(instr, instr.Operands[0]); err != nil {
				return err
			}
		}

		// Check for standard RISC-V syntax: sw rs2, offset(rs1)
		if len(instr.Operands) == 2 {
			if offsetBase, ok := instr.Operands[1].(*ast.OffsetBase); ok {
				if rs1, err = evalRegister(instr, offsetBase.Base); err != nil {
					return err
				}
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
			if rs1, err = evalRegister(instr, instr.Operands[2]); err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 2 {
			immExpr := instr.Operands[1]
			if ob, ok := immExpr.(*ast.OffsetBase); ok {
				immExpr = ob.Offset
			}
			if err := checkIImmediate(instr, immExpr, imm); err != nil {
				return err
			}
		}
		encoded = arch.EncodeTypeS(op, rs1, rs2, imm)

	case arch.TypeB:
		rs1, rs2 := uint32(0), uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			if rs1, err = evalRegister(instr, instr.Operands[0]); err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 2 {
			if rs2, err = evalRegister(instr, instr.Operands[1]); err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 3 {
			target, err := evalExpressionWithError(instr.Operands[2], symbols, pc)
			if err != nil {
				return err
			}
			if err := checkOffset(instr, target-pc, 4096); err != nil {
				return err
			}
			imm = int32(target - pc)
		}
		encoded = arch.EncodeTypeB(op, rs1, rs2, imm)

	case arch.TypeU:
		rd := uint32(0)
		imm := int32(0)
		if len(instr.Operands) >= 1 {
			if rd, err = evalRegister(instr, instr.Operands[0]); err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 2 {
			imm, err = evalImmediateWithError(instr.Operands[1], symbols, pc)
			if err != nil {
				return err
			}
		}
		if len(instr.Operands) >= 2 {
			if err := checkUImmediate(instr, instr.Operands[1], imm); err != nil {
				return err
			}
		}
		encoded = arch.EncodeTypeU(op, rd, imm)

	case arch.TypeJ:
		// jal rd, target; jal target is jal ra, target
		var rd uint32
		var targetExpr ast.Expression
		switch len(instr.Operands) {
		case 1:
			rd = 1 // ra
			targetExpr = instr.Operands[0]
		case 2:
			if rd, err = evalRegister(instr, instr.Operands[0]); err != nil {
				return err
			}
			targetExpr = instr.Operands[1]
		default:
			return fmt.Errorf("%s: error: %s expects 1 or 2 operands ([rd,] target), got %d",
				instr.Location, instr.Opcode, len(instr.Operands))
		}
		target, err := evalExpressionWithError(targetExpr, symbols, pc)
		if err != nil {
			return err
		}
		if err := checkOffset(instr, target-pc, 1<<20); err != nil {
			return err
		}
		encoded = arch.EncodeTypeJ(op, rd, int32(target-pc))

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

// isCSR reports whether mnemonic is one of the six CSR instructions
func isCSR(mnemonic string) bool {
	switch mnemonic {
	case "csrrw", "csrrs", "csrrc", "csrrwi", "csrrsi", "csrrci":
		return true
	}
	return false
}

// encodeCSR encodes the CSR instructions in the standard RISC-V order:
// "csrrX rd, csr, rs1" and "csrrXi rd, csr, uimm". The CSR number goes in
// the I-type immediate field, rs1 or the 5-bit uimm in the rs1 field.
// The old goasm order (rd, rs1/uimm, csr) is refused with an error.
func encodeCSR(op arch.InstrOpcode, instr *ast.Instruction, symbols *symbol.Table, pc int64) ([]byte, error) {
	immediate := strings.HasSuffix(instr.Opcode, "i")
	srcName := "rs1"
	if immediate {
		srcName = "uimm"
	}
	order := fmt.Sprintf("the order is %s rd, csr, %s", instr.Opcode, srcName)

	if len(instr.Operands) != 3 {
		return nil, fmt.Errorf("%s: error: %s expects 3 operands (rd, csr, %s), got %d",
			instr.Location, instr.Opcode, srcName, len(instr.Operands))
	}
	rd, err := evalRegister(instr, instr.Operands[0])
	if err != nil {
		return nil, err
	}

	// a register in the CSR place is the old order
	if ident, ok := instr.Operands[1].(*ast.Identifier); ok && arch.IsRegister(ident.Name) {
		return nil, fmt.Errorf("%s: error: %s: the CSR operand '%s' is a register; %s",
			instr.Location, instr.Opcode, ident.Name, order)
	}
	csr, err := evalImmediateWithError(instr.Operands[1], symbols, pc)
	if err != nil {
		return nil, err
	}
	if csr < 0 || csr > 0xFFF {
		return nil, fmt.Errorf("%s: error: CSR number %#x out of range (0..0xfff)", instr.Location, csr)
	}

	var src uint32
	if immediate {
		if ident, ok := instr.Operands[2].(*ast.Identifier); ok && arch.IsRegister(ident.Name) {
			return nil, fmt.Errorf("%s: error: %s: uimm must be a number (0..31), not register %s",
				instr.Location, instr.Opcode, ident.Name)
		}
		uimm, err := evalImmediateWithError(instr.Operands[2], symbols, pc)
		if err != nil {
			return nil, err
		}
		if uimm < 0 || uimm > 31 {
			hint := ""
			if csr <= 31 {
				// a small "CSR" and a big "uimm": likely rd, uimm, csr
				hint = fmt.Sprintf(" (old order rd, uimm, csr? %s)", order)
			}
			return nil, fmt.Errorf("%s: error: %s immediate %d out of range (0..31)%s",
				instr.Location, instr.Opcode, uimm, hint)
		}
		src = uint32(uimm)
	} else {
		if src, err = evalRegister(instr, instr.Operands[2]); err != nil {
			return nil, err
		}
	}

	return arch.EncodeTypeI(op, rd, src, csr), nil
}

// isFieldBits reports whether expr is a relocation modifier: its value is
// already the bits of the field (%lo is 0..4095, %hi up to 0xFFFFF), so it is
// not range-checked
func isFieldBits(expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.LoRel, *ast.HiRel, *ast.PCRelLo, *ast.PCRelHi:
		return true
	}
	return false
}

// checkIImmediate checks the 12-bit signed immediate of an I-type or S-type instruction
func checkIImmediate(instr *ast.Instruction, expr ast.Expression, imm int32) error {
	if isFieldBits(expr) {
		return nil
	}
	if imm < -2048 || imm > 2047 {
		return fmt.Errorf("%s: error: %s immediate %d out of range (-2048..2047)", instr.Location, instr.Opcode, imm)
	}
	return nil
}

// checkUImmediate checks the 20-bit immediate of lui/auipc: signed or unsigned
func checkUImmediate(instr *ast.Instruction, expr ast.Expression, imm int32) error {
	if isFieldBits(expr) {
		return nil
	}
	if imm < -(1<<19) || imm > (1<<20)-1 {
		return fmt.Errorf("%s: error: %s immediate %d out of range (%d..%d)", instr.Location, instr.Opcode, imm, -(1 << 19), (1<<20)-1)
	}
	return nil
}

// checkOffset checks a branch or jal offset: even, and in -limit..limit-2
func checkOffset(instr *ast.Instruction, offset int64, limit int64) error {
	if offset < -limit || offset > limit-2 {
		return fmt.Errorf("%s: error: %s offset %d out of range (%d..%d)", instr.Location, instr.Opcode, offset, -limit, limit-2)
	}
	if offset%2 != 0 {
		return fmt.Errorf("%s: error: %s offset %d is odd", instr.Location, instr.Opcode, offset)
	}
	return nil
}
