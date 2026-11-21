// Bit extraction helpers for RISC-V instruction decoding
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package decoder

// extractOpcode extracts the opcode field (bits [6:0])
func extractOpcode(word uint32) uint32 {
	return word & 0x7F
}

// extractRd extracts the destination register field (bits [11:7])
func extractRd(word uint32) uint32 {
	return (word >> 7) & 0x1F
}

// extractFunct3 extracts the funct3 field (bits [14:12])
func extractFunct3(word uint32) uint32 {
	return (word >> 12) & 0x7
}

// extractRs1 extracts the source register 1 field (bits [19:15])
func extractRs1(word uint32) uint32 {
	return (word >> 15) & 0x1F
}

// extractRs2 extracts the source register 2 field (bits [24:20])
func extractRs2(word uint32) uint32 {
	return (word >> 20) & 0x1F
}

// extractFunct7 extracts the funct7 field (bits [31:25])
func extractFunct7(word uint32) uint32 {
	return (word >> 25) & 0x7F
}

// extractImmI extracts and sign-extends an I-type immediate (bits [31:20])
// Format: imm[11:0] at bits [31:20]
func extractImmI(word uint32) int32 {
	imm := (word >> 20) & 0xFFF
	// Sign-extend from 12 bits
	if imm&0x800 != 0 {
		imm |= 0xFFFFF000
	}
	return int32(imm)
}

// extractImmS extracts and sign-extends an S-type immediate
// Format: imm[11:5] at bits [31:25], imm[4:0] at bits [11:7]
func extractImmS(word uint32) int32 {
	immHi := (word >> 25) & 0x7F // bits [11:5]
	immLo := (word >> 7) & 0x1F  // bits [4:0]
	imm := (immHi << 5) | immLo
	// Sign-extend from 12 bits
	if imm&0x800 != 0 {
		imm |= 0xFFFFF000
	}
	return int32(imm)
}

// extractImmB extracts and sign-extends a B-type immediate
// Format: imm[12|10:5] at bits [31:25], imm[4:1|11] at bits [11:7]
// Note: bit 0 is always 0 (halfword aligned)
func extractImmB(word uint32) int32 {
	bit12 := (word >> 31) & 0x1
	bits10_5 := (word >> 25) & 0x3F
	bits4_1 := (word >> 8) & 0xF
	bit11 := (word >> 7) & 0x1

	imm := (bit12 << 12) | (bit11 << 11) | (bits10_5 << 5) | (bits4_1 << 1)
	// Sign-extend from 13 bits
	if imm&0x1000 != 0 {
		imm |= 0xFFFFE000
	}
	return int32(imm)
}

// extractImmU extracts a U-type immediate (bits [31:12])
// Format: imm[31:12] at bits [31:12]
// Returns the value already shifted (upper 20 bits in position)
func extractImmU(word uint32) int32 {
	return int32(word & 0xFFFFF000)
}

// extractImmUValue extracts a U-type immediate as the 20-bit value
// (not shifted, just the raw upper 20 bits value)
func extractImmUValue(word uint32) int32 {
	imm := (word >> 12) & 0xFFFFF
	// Sign-extend from 20 bits
	if imm&0x80000 != 0 {
		imm |= 0xFFF00000
	}
	return int32(imm)
}

// extractImmJ extracts and sign-extends a J-type immediate
// Format: imm[20|10:1|11|19:12] at bits [31:12]
// Note: bit 0 is always 0 (halfword aligned)
func extractImmJ(word uint32) int32 {
	bit20 := (word >> 31) & 0x1
	bits10_1 := (word >> 21) & 0x3FF
	bit11 := (word >> 20) & 0x1
	bits19_12 := (word >> 12) & 0xFF

	imm := (bit20 << 20) | (bits19_12 << 12) | (bit11 << 11) | (bits10_1 << 1)
	// Sign-extend from 21 bits
	if imm&0x100000 != 0 {
		imm |= 0xFFE00000
	}
	return int32(imm)
}

// extractShamt extracts the shift amount for I2-type instructions (bits [24:20])
func extractShamt(word uint32) uint32 {
	return (word >> 20) & 0x1F
}
