// RISC-V instruction encoders for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package arch

import (
	"encoding/binary"
	"fmt"
)

// EncodeResult holds the encoded instruction and any relocations
type EncodeResult struct {
	Bytes  []byte
	Relocs []Relocation
}

// RelocationType defines the type of relocation needed
type RelocationType int

const (
	RelocNone RelocationType = iota
	RelocHi20              // %hi(symbol) - upper 20 bits
	RelocLo12I             // %lo(symbol) - lower 12 bits for I-type
	RelocLo12S             // %lo(symbol) - lower 12 bits for S-type
	RelocPCRelHi20         // %pcrel_hi(symbol) - PC-relative upper 20 bits
	RelocPCRelLo12I        // %pcrel_lo(symbol) - PC-relative lower 12 bits for I-type
	RelocPCRelLo12S        // %pcrel_lo(symbol) - PC-relative lower 12 bits for S-type
	RelocBranch            // Branch offset (B-type)
	RelocJump              // Jump offset (J-type)
)

// Relocation represents a relocation entry
type Relocation struct {
	Type   RelocationType
	Symbol string
	Offset int // Offset within the instruction
}

// EncodeTypeR encodes an R-type instruction
// Format: funct7[31:25] | rs2[24:20] | rs1[19:15] | funct3[14:12] | rd[11:7] | opcode[6:0]
func EncodeTypeR(op InstrOpcode, rd, rs1, rs2 uint32) []byte {
	value := (op.Funct7 << 25) |
		((rs2 & 0x1F) << 20) |
		((rs1 & 0x1F) << 15) |
		(op.Funct3 << 12) |
		((rd & 0x1F) << 7) |
		op.Opcode

	return uint32ToBytes(value)
}

// EncodeTypeR2 encodes an R-type instruction with two source registers (rd=0)
// Used for instructions like sleep.t that have no destination register
// Format: funct7[31:25] | rs2[24:20] | rs1[19:15] | funct3[14:12] | rd[11:7] | opcode[6:0]
func EncodeTypeR2(op InstrOpcode, rs1, rs2 uint32) []byte {
	value := (op.Funct7 << 25) |
		((rs2 & 0x1F) << 20) |
		((rs1 & 0x1F) << 15) |
		(op.Funct3 << 12) |
		(0 << 7) | // rd = 0
		op.Opcode

	return uint32ToBytes(value)
}

// EncodeTypeI encodes an I-type instruction
// Format: imm[31:20] | rs1[19:15] | funct3[14:12] | rd[11:7] | opcode[6:0]
func EncodeTypeI(op InstrOpcode, rd, rs1 uint32, imm int32) []byte {
	// Sign-extend immediate to 12 bits
	immBits := uint32(imm) & 0xFFF

	value := (immBits << 20) |
		((rs1 & 0x1F) << 15) |
		(op.Funct3 << 12) |
		((rd & 0x1F) << 7) |
		op.Opcode

	return uint32ToBytes(value)
}

// EncodeTypeI2 encodes an I-type shift instruction (slli, srli, srai)
// Format: funct7[31:25] | shamt[24:20] | rs1[19:15] | funct3[14:12] | rd[11:7] | opcode[6:0]
func EncodeTypeI2(op InstrOpcode, rd, rs1 uint32, shamt uint32) []byte {
	value := (op.Funct7 << 25) |
		((shamt & 0x1F) << 20) |
		((rs1 & 0x1F) << 15) |
		(op.Funct3 << 12) |
		((rd & 0x1F) << 7) |
		op.Opcode

	return uint32ToBytes(value)
}

// EncodeTypeS encodes an S-type instruction
// Format: imm[11:5][31:25] | rs2[24:20] | rs1[19:15] | funct3[14:12] | imm[4:0][11:7] | opcode[6:0]
func EncodeTypeS(op InstrOpcode, rs1, rs2 uint32, imm int32) []byte {
	immBits := uint32(imm) & 0xFFF
	immHi := (immBits >> 5) & 0x7F
	immLo := immBits & 0x1F

	value := (immHi << 25) |
		((rs2 & 0x1F) << 20) |
		((rs1 & 0x1F) << 15) |
		(op.Funct3 << 12) |
		(immLo << 7) |
		op.Opcode

	return uint32ToBytes(value)
}

// EncodeTypeB encodes a B-type instruction
// Format: imm[12|10:5][31:25] | rs2[24:20] | rs1[19:15] | funct3[14:12] | imm[4:1|11][11:7] | opcode[6:0]
func EncodeTypeB(op InstrOpcode, rs1, rs2 uint32, imm int32) []byte {
	// B-type immediate is shifted left by 1 (halfword aligned)
	immBits := uint32(imm)

	// Extract bits: imm[12], imm[10:5], imm[4:1], imm[11]
	bit12 := (immBits >> 12) & 0x1
	bits10_5 := (immBits >> 5) & 0x3F
	bits4_1 := (immBits >> 1) & 0xF
	bit11 := (immBits >> 11) & 0x1

	value := (bit12 << 31) |
		(bits10_5 << 25) |
		((rs2 & 0x1F) << 20) |
		((rs1 & 0x1F) << 15) |
		(op.Funct3 << 12) |
		(bits4_1 << 8) |
		(bit11 << 7) |
		op.Opcode

	return uint32ToBytes(value)
}

// EncodeTypeU encodes a U-type instruction
// Format: imm[31:12] | rd[11:7] | opcode[6:0]
func EncodeTypeU(op InstrOpcode, rd uint32, imm int32) []byte {
	// U-type immediate is the upper 20 bits
	immBits := uint32(imm) & 0xFFFFF

	value := (immBits << 12) |
		((rd & 0x1F) << 7) |
		op.Opcode

	return uint32ToBytes(value)
}

// EncodeTypeJ encodes a J-type instruction
// Format: imm[20|10:1|11|19:12][31:12] | rd[11:7] | opcode[6:0]
func EncodeTypeJ(op InstrOpcode, rd uint32, imm int32) []byte {
	// J-type immediate encoding is complex
	immBits := uint32(imm)

	// Extract bits: imm[20], imm[10:1], imm[11], imm[19:12]
	bit20 := (immBits >> 20) & 0x1
	bits10_1 := (immBits >> 1) & 0x3FF
	bit11 := (immBits >> 11) & 0x1
	bits19_12 := (immBits >> 12) & 0xFF

	value := (bit20 << 31) |
		(bits10_1 << 21) |
		(bit11 << 20) |
		(bits19_12 << 12) |
		((rd & 0x1F) << 7) |
		op.Opcode

	return uint32ToBytes(value)
}

// Encode encodes an instruction given its opcode info and operands.
// This is a high-level function that dispatches to the appropriate encoder.
func Encode(mnemonic string, rd, rs1, rs2 uint32, imm int32) ([]byte, error) {
	op, ok := LookupOpcode(mnemonic)
	if !ok {
		return nil, fmt.Errorf("unknown instruction: %s", mnemonic)
	}

	switch op.Type {
	case TypeR:
		return EncodeTypeR(op, rd, rs1, rs2), nil
	case TypeR2:
		return EncodeTypeR2(op, rs1, rs2), nil
	case TypeI:
		return EncodeTypeI(op, rd, rs1, imm), nil
	case TypeI2:
		return EncodeTypeI2(op, rd, rs1, uint32(imm)), nil
	case TypeS:
		return EncodeTypeS(op, rs1, rs2, imm), nil
	case TypeB:
		return EncodeTypeB(op, rs1, rs2, imm), nil
	case TypeU:
		return EncodeTypeU(op, rd, imm), nil
	case TypeJ:
		return EncodeTypeJ(op, rd, imm), nil
	default:
		return nil, fmt.Errorf("unsupported instruction type: %v", op.Type)
	}
}

// uint32ToBytes converts a uint32 to little-endian bytes
func uint32ToBytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

// SignExtend12 sign-extends a 12-bit value to int32
func SignExtend12(v uint32) int32 {
	if v&0x800 != 0 {
		return int32(v | 0xFFFFF000)
	}
	return int32(v)
}

// SignExtend20 sign-extends a 20-bit value to int32
func SignExtend20(v uint32) int32 {
	if v&0x80000 != 0 {
		return int32(v | 0xFFF00000)
	}
	return int32(v)
}
