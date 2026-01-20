// RISC-V instruction decoder for the NATOGA32 disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package decoder

import (
	"encoding/binary"

	"github.com/natoga32/goasm/internal/arch"
)

// Decode decodes a 32-bit instruction word into a DecodedInstruction
func Decode(word uint32) *DecodedInstruction {
	entry, found := LookupInstruction(word)
	if !found {
		return &DecodedInstruction{
			Mnemonic: "",
			RawWord:  word,
			Rd:       NoRegister,
			Rs1:      NoRegister,
			Rs2:      NoRegister,
		}
	}

	instr := &DecodedInstruction{
		Mnemonic: entry.Mnemonic,
		Type:     entry.Info.Type,
		RawWord:  word,
		Rd:       NoRegister,
		Rs1:      NoRegister,
		Rs2:      NoRegister,
	}

	// Decode operands based on instruction type
	switch entry.Info.Type {
	case arch.TypeR:
		decodeTypeR(word, instr)
	case arch.TypeR2:
		decodeTypeR2(word, instr)
	case arch.TypeI:
		decodeTypeI(word, instr)
	case arch.TypeI2:
		decodeTypeI2(word, instr)
	case arch.TypeS:
		decodeTypeS(word, instr)
	case arch.TypeB:
		decodeTypeB(word, instr)
	case arch.TypeU:
		decodeTypeU(word, instr)
	case arch.TypeJ:
		decodeTypeJ(word, instr)
	}

	return instr
}

// DecodeBytes decodes a 4-byte slice (little-endian) into a DecodedInstruction
func DecodeBytes(data []byte) *DecodedInstruction {
	if len(data) < 4 {
		return &DecodedInstruction{
			Mnemonic: "",
			Rd:       NoRegister,
			Rs1:      NoRegister,
			Rs2:      NoRegister,
		}
	}
	word := binary.LittleEndian.Uint32(data)
	return Decode(word)
}

// decodeTypeR decodes R-type instruction operands
// Format: funct7[31:25] | rs2[24:20] | rs1[19:15] | funct3[14:12] | rd[11:7] | opcode[6:0]
func decodeTypeR(word uint32, instr *DecodedInstruction) {
	instr.Rd = int(extractRd(word))
	instr.Rs1 = int(extractRs1(word))
	instr.Rs2 = int(extractRs2(word))
}

// decodeTypeR2 decodes R2-type instruction operands (rd=0, two source registers)
// Format: funct7[31:25] | rs2[24:20] | rs1[19:15] | funct3[14:12] | rd[11:7] | opcode[6:0]
func decodeTypeR2(word uint32, instr *DecodedInstruction) {
	instr.Rs1 = int(extractRs1(word))
	instr.Rs2 = int(extractRs2(word))
	// rd is not used in R2-type
}

// decodeTypeI decodes I-type instruction operands
// Format: imm[31:20] | rs1[19:15] | funct3[14:12] | rd[11:7] | opcode[6:0]
func decodeTypeI(word uint32, instr *DecodedInstruction) {
	instr.Rd = int(extractRd(word))
	instr.Rs1 = int(extractRs1(word))
	instr.Immediate = extractImmI(word)
}

// decodeTypeI2 decodes I2-type (shift immediate) instruction operands
// Format: funct7[31:25] | shamt[24:20] | rs1[19:15] | funct3[14:12] | rd[11:7] | opcode[6:0]
func decodeTypeI2(word uint32, instr *DecodedInstruction) {
	instr.Rd = int(extractRd(word))
	instr.Rs1 = int(extractRs1(word))
	instr.Immediate = int32(extractShamt(word))
}

// decodeTypeS decodes S-type instruction operands
// Format: imm[11:5][31:25] | rs2[24:20] | rs1[19:15] | funct3[14:12] | imm[4:0][11:7] | opcode[6:0]
func decodeTypeS(word uint32, instr *DecodedInstruction) {
	instr.Rs1 = int(extractRs1(word))
	instr.Rs2 = int(extractRs2(word))
	instr.Immediate = extractImmS(word)
}

// decodeTypeB decodes B-type instruction operands
// Format: imm[12|10:5][31:25] | rs2[24:20] | rs1[19:15] | funct3[14:12] | imm[4:1|11][11:7] | opcode[6:0]
func decodeTypeB(word uint32, instr *DecodedInstruction) {
	instr.Rs1 = int(extractRs1(word))
	instr.Rs2 = int(extractRs2(word))
	instr.Immediate = extractImmB(word)
}

// decodeTypeU decodes U-type instruction operands
// Format: imm[31:12] | rd[11:7] | opcode[6:0]
func decodeTypeU(word uint32, instr *DecodedInstruction) {
	instr.Rd = int(extractRd(word))
	instr.Immediate = extractImmUValue(word)
}

// decodeTypeJ decodes J-type instruction operands
// Format: imm[20|10:1|11|19:12][31:12] | rd[11:7] | opcode[6:0]
func decodeTypeJ(word uint32, instr *DecodedInstruction) {
	instr.Rd = int(extractRd(word))
	instr.Immediate = extractImmJ(word)
}
