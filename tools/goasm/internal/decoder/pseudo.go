// Pseudo-instruction recognition for the NATOGA32 disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package decoder

import (
	"fmt"

	"github.com/natoga32/goasm/internal/arch"
)

// PseudoMatcher recognizes instruction patterns that map to pseudo-instructions
type PseudoMatcher struct{}

// NewPseudoMatcher creates a new PseudoMatcher
func NewPseudoMatcher() *PseudoMatcher {
	return &PseudoMatcher{}
}

// Match attempts to match a decoded instruction to a pseudo-instruction.
// Returns the pseudo-instruction mnemonic and formatted operands, or empty strings if no match.
func (pm *PseudoMatcher) Match(instr *DecodedInstruction, addr uint32, useABI bool) (mnemonic string, operands string, matched bool) {
	if instr.IsUnknown() {
		return "", "", false
	}

	switch instr.Mnemonic {
	case "addi":
		return pm.matchAddi(instr, useABI)
	case "jalr":
		return pm.matchJalr(instr, useABI)
	case "jal":
		return pm.matchJal(instr, addr, useABI)
	case "beq":
		return pm.matchBeq(instr, addr, useABI)
	case "bne":
		return pm.matchBne(instr, addr, useABI)
	case "blt":
		return pm.matchBlt(instr, addr, useABI)
	case "bge":
		return pm.matchBge(instr, addr, useABI)
	case "bltu":
		return pm.matchBltu(instr, addr, useABI)
	case "bgeu":
		return pm.matchBgeu(instr, addr, useABI)
	case "slt":
		return pm.matchSlt(instr, useABI)
	case "sltu":
		return pm.matchSltu(instr, useABI)
	case "sltiu":
		return pm.matchSltiu(instr, useABI)
	case "sub":
		return pm.matchSub(instr, useABI)
	case "xori":
		return pm.matchXori(instr, useABI)
	}

	return "", "", false
}

// matchAddi handles addi-based pseudo-instructions:
// - nop: addi x0, x0, 0
// - li: addi rd, x0, imm
// - mv: addi rd, rs1, 0
func (pm *PseudoMatcher) matchAddi(instr *DecodedInstruction, useABI bool) (string, string, bool) {
	// nop: addi x0, x0, 0
	if instr.Rd == 0 && instr.Rs1 == 0 && instr.Immediate == 0 {
		return "nop", "", true
	}

	// li: addi rd, x0, imm
	if instr.Rs1 == 0 {
		rd := regName(instr.Rd, useABI)
		return "li", fmt.Sprintf("%s, %d", rd, instr.Immediate), true
	}

	// mv: addi rd, rs1, 0
	if instr.Immediate == 0 {
		rd := regName(instr.Rd, useABI)
		rs1 := regName(instr.Rs1, useABI)
		return "mv", fmt.Sprintf("%s, %s", rd, rs1), true
	}

	return "", "", false
}

// matchJalr handles jalr-based pseudo-instructions:
// - ret: jalr x0, ra, 0
// - jr: jalr x0, rs1, 0
func (pm *PseudoMatcher) matchJalr(instr *DecodedInstruction, useABI bool) (string, string, bool) {
	// Only match when rd=x0 and imm=0
	if instr.Rd != 0 || instr.Immediate != 0 {
		return "", "", false
	}

	// ret: jalr x0, ra, 0
	if instr.Rs1 == 1 { // ra = x1
		return "ret", "", true
	}

	// jr: jalr x0, rs1, 0
	rs1 := regName(instr.Rs1, useABI)
	return "jr", rs1, true
}

// matchJal handles jal-based pseudo-instructions:
// - j: jal x0, offset
// - call: jal ra, offset
func (pm *PseudoMatcher) matchJal(instr *DecodedInstruction, addr uint32, useABI bool) (string, string, bool) {
	target := calculateTarget(addr, instr.Immediate)

	// j: jal x0, offset
	if instr.Rd == 0 {
		return "j", fmt.Sprintf("0x%X", target), true
	}

	// call: jal ra, offset
	if instr.Rd == 1 { // ra = x1
		return "call", fmt.Sprintf("0x%X", target), true
	}

	return "", "", false
}

// matchBeq handles beq-based pseudo-instructions:
// - beqz: beq rs1, x0, offset
func (pm *PseudoMatcher) matchBeq(instr *DecodedInstruction, addr uint32, useABI bool) (string, string, bool) {
	// beqz: beq rs1, x0, offset
	if instr.Rs2 == 0 {
		target := calculateTarget(addr, instr.Immediate)
		rs1 := regName(instr.Rs1, useABI)
		return "beqz", fmt.Sprintf("%s, 0x%X", rs1, target), true
	}

	return "", "", false
}

// matchBne handles bne-based pseudo-instructions:
// - bnez: bne rs1, x0, offset
func (pm *PseudoMatcher) matchBne(instr *DecodedInstruction, addr uint32, useABI bool) (string, string, bool) {
	// bnez: bne rs1, x0, offset
	if instr.Rs2 == 0 {
		target := calculateTarget(addr, instr.Immediate)
		rs1 := regName(instr.Rs1, useABI)
		return "bnez", fmt.Sprintf("%s, 0x%X", rs1, target), true
	}

	return "", "", false
}

// matchBlt handles blt-based pseudo-instructions:
// - bgtz: blt x0, rs2, offset
// - bltz: blt rs1, x0, offset
// - bgt: blt rs2, rs1, offset (swapped operands)
func (pm *PseudoMatcher) matchBlt(instr *DecodedInstruction, addr uint32, useABI bool) (string, string, bool) {
	target := calculateTarget(addr, instr.Immediate)

	// bgtz: blt x0, rs2, offset
	if instr.Rs1 == 0 && instr.Rs2 != 0 {
		rs2 := regName(instr.Rs2, useABI)
		return "bgtz", fmt.Sprintf("%s, 0x%X", rs2, target), true
	}

	// bltz: blt rs1, x0, offset
	if instr.Rs2 == 0 && instr.Rs1 != 0 {
		rs1 := regName(instr.Rs1, useABI)
		return "bltz", fmt.Sprintf("%s, 0x%X", rs1, target), true
	}

	// bgt: blt rs2, rs1, offset (swapped operands, both non-zero)
	if instr.Rs1 != 0 && instr.Rs2 != 0 {
		// Swap: blt rs1, rs2 => bgt rs2, rs1
		rs1 := regName(instr.Rs1, useABI)
		rs2 := regName(instr.Rs2, useABI)
		return "bgt", fmt.Sprintf("%s, %s, 0x%X", rs2, rs1, target), true
	}

	return "", "", false
}

// matchBge handles bge-based pseudo-instructions:
// - bgez: bge rs1, x0, offset
// - blez: bge x0, rs2, offset
// - ble: bge rs2, rs1, offset (swapped operands)
func (pm *PseudoMatcher) matchBge(instr *DecodedInstruction, addr uint32, useABI bool) (string, string, bool) {
	target := calculateTarget(addr, instr.Immediate)

	// bgez: bge rs1, x0, offset
	if instr.Rs2 == 0 && instr.Rs1 != 0 {
		rs1 := regName(instr.Rs1, useABI)
		return "bgez", fmt.Sprintf("%s, 0x%X", rs1, target), true
	}

	// blez: bge x0, rs2, offset
	if instr.Rs1 == 0 && instr.Rs2 != 0 {
		rs2 := regName(instr.Rs2, useABI)
		return "blez", fmt.Sprintf("%s, 0x%X", rs2, target), true
	}

	// ble: bge rs2, rs1, offset (swapped operands, both non-zero)
	if instr.Rs1 != 0 && instr.Rs2 != 0 {
		// Swap: bge rs1, rs2 => ble rs2, rs1
		rs1 := regName(instr.Rs1, useABI)
		rs2 := regName(instr.Rs2, useABI)
		return "ble", fmt.Sprintf("%s, %s, 0x%X", rs2, rs1, target), true
	}

	return "", "", false
}

// matchBltu handles bltu-based pseudo-instructions:
// - bgtu: bltu rs2, rs1, offset (swapped operands)
func (pm *PseudoMatcher) matchBltu(instr *DecodedInstruction, addr uint32, useABI bool) (string, string, bool) {
	// bgtu: bltu rs2, rs1, offset (swapped operands, both non-zero)
	if instr.Rs1 != 0 && instr.Rs2 != 0 {
		target := calculateTarget(addr, instr.Immediate)
		rs1 := regName(instr.Rs1, useABI)
		rs2 := regName(instr.Rs2, useABI)
		return "bgtu", fmt.Sprintf("%s, %s, 0x%X", rs2, rs1, target), true
	}

	return "", "", false
}

// matchBgeu handles bgeu-based pseudo-instructions:
// - bleu: bgeu rs2, rs1, offset (swapped operands)
func (pm *PseudoMatcher) matchBgeu(instr *DecodedInstruction, addr uint32, useABI bool) (string, string, bool) {
	// bleu: bgeu rs2, rs1, offset (swapped operands, both non-zero)
	if instr.Rs1 != 0 && instr.Rs2 != 0 {
		target := calculateTarget(addr, instr.Immediate)
		rs1 := regName(instr.Rs1, useABI)
		rs2 := regName(instr.Rs2, useABI)
		return "bleu", fmt.Sprintf("%s, %s, 0x%X", rs2, rs1, target), true
	}

	return "", "", false
}

// matchSlt handles slt-based pseudo-instructions:
// - sltz: slt rd, rs1, x0
// - sgtz: slt rd, x0, rs1
func (pm *PseudoMatcher) matchSlt(instr *DecodedInstruction, useABI bool) (string, string, bool) {
	rd := regName(instr.Rd, useABI)

	// sltz: slt rd, rs1, x0
	if instr.Rs2 == 0 && instr.Rs1 != 0 {
		rs1 := regName(instr.Rs1, useABI)
		return "sltz", fmt.Sprintf("%s, %s", rd, rs1), true
	}

	// sgtz: slt rd, x0, rs1
	if instr.Rs1 == 0 && instr.Rs2 != 0 {
		rs2 := regName(instr.Rs2, useABI)
		return "sgtz", fmt.Sprintf("%s, %s", rd, rs2), true
	}

	return "", "", false
}

// matchSltu handles sltu-based pseudo-instructions:
// - snez: sltu rd, x0, rs1
func (pm *PseudoMatcher) matchSltu(instr *DecodedInstruction, useABI bool) (string, string, bool) {
	// snez: sltu rd, x0, rs1
	if instr.Rs1 == 0 && instr.Rs2 != 0 {
		rd := regName(instr.Rd, useABI)
		rs2 := regName(instr.Rs2, useABI)
		return "snez", fmt.Sprintf("%s, %s", rd, rs2), true
	}

	return "", "", false
}

// matchSltiu handles sltiu-based pseudo-instructions:
// - seqz: sltiu rd, rs1, 1
func (pm *PseudoMatcher) matchSltiu(instr *DecodedInstruction, useABI bool) (string, string, bool) {
	// seqz: sltiu rd, rs1, 1
	if instr.Immediate == 1 {
		rd := regName(instr.Rd, useABI)
		rs1 := regName(instr.Rs1, useABI)
		return "seqz", fmt.Sprintf("%s, %s", rd, rs1), true
	}

	return "", "", false
}

// matchSub handles sub-based pseudo-instructions:
// - neg: sub rd, x0, rs1
func (pm *PseudoMatcher) matchSub(instr *DecodedInstruction, useABI bool) (string, string, bool) {
	// neg: sub rd, x0, rs1
	if instr.Rs1 == 0 && instr.Rs2 != 0 {
		rd := regName(instr.Rd, useABI)
		rs2 := regName(instr.Rs2, useABI)
		return "neg", fmt.Sprintf("%s, %s", rd, rs2), true
	}

	return "", "", false
}

// matchXori handles xori-based pseudo-instructions:
// - not: xori rd, rs1, -1
func (pm *PseudoMatcher) matchXori(instr *DecodedInstruction, useABI bool) (string, string, bool) {
	// not: xori rd, rs1, -1
	if instr.Immediate == -1 {
		rd := regName(instr.Rd, useABI)
		rs1 := regName(instr.Rs1, useABI)
		return "not", fmt.Sprintf("%s, %s", rd, rs1), true
	}

	return "", "", false
}

// regName returns the register name based on the useABI flag
func regName(reg int, useABI bool) string {
	return arch.RegisterName(reg, useABI)
}

// calculateTarget calculates the absolute target address from PC and offset
func calculateTarget(addr uint32, offset int32) uint32 {
	return uint32(int32(addr) + offset)
}
