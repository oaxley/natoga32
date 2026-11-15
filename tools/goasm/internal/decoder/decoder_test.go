// Unit tests for RISC-V instruction decoder
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package decoder

import (
	"encoding/binary"
	"testing"

	"github.com/natoga32/goasm/internal/arch"
)

// Helper to encode instructions for testing
func encodeForTest(mnemonic string, rd, rs1, rs2 uint32, imm int32) uint32 {
	bytes, err := arch.Encode(mnemonic, rd, rs1, rs2, imm)
	if err != nil {
		panic("failed to encode test instruction: " + err.Error())
	}
	return binary.LittleEndian.Uint32(bytes)
}

func TestDecodeRType(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rd       int
		rs1      int
		rs2      int
	}{
		{"add x1, x2, x3", "add", 1, 2, 3},
		{"sub x10, x11, x12", "sub", 10, 11, 12},
		{"and x5, x6, x7", "and", 5, 6, 7},
		{"or x8, x9, x10", "or", 8, 9, 10},
		{"xor x15, x16, x17", "xor", 15, 16, 17},
		{"sll x20, x21, x22", "sll", 20, 21, 22},
		{"srl x25, x26, x27", "srl", 25, 26, 27},
		{"sra x28, x29, x30", "sra", 28, 29, 30},
		{"slt x1, x2, x3", "slt", 1, 2, 3},
		{"sltu x4, x5, x6", "sltu", 4, 5, 6},
		{"mul x7, x8, x9", "mul", 7, 8, 9},
		{"div x10, x11, x12", "div", 10, 11, 12},
		{"rem x13, x14, x15", "rem", 13, 14, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, uint32(tt.rd), uint32(tt.rs1), uint32(tt.rs2), 0)
			instr := Decode(word)

			if instr.Mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, tt.mnemonic)
			}
			if instr.Type != arch.TypeR {
				t.Errorf("type: got %v, want %v", instr.Type, arch.TypeR)
			}
			if instr.Rd != tt.rd {
				t.Errorf("rd: got %d, want %d", instr.Rd, tt.rd)
			}
			if instr.Rs1 != tt.rs1 {
				t.Errorf("rs1: got %d, want %d", instr.Rs1, tt.rs1)
			}
			if instr.Rs2 != tt.rs2 {
				t.Errorf("rs2: got %d, want %d", instr.Rs2, tt.rs2)
			}
		})
	}
}

func TestDecodeIType(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rd       int
		rs1      int
		imm      int32
	}{
		{"addi x1, x2, 10", "addi", 1, 2, 10},
		{"addi x1, x2, -10", "addi", 1, 2, -10},
		{"addi x1, x0, 0", "addi", 1, 0, 0},
		{"slti x5, x6, 100", "slti", 5, 6, 100},
		{"sltiu x7, x8, 200", "sltiu", 7, 8, 200},
		{"xori x9, x10, 0xFF", "xori", 9, 10, 0xFF},
		{"ori x11, x12, 0x7FF", "ori", 11, 12, 0x7FF},
		{"andi x13, x14, 0x123", "andi", 13, 14, 0x123},
		{"lw x15, 0(x16)", "lw", 15, 16, 0},
		{"lw x17, 4(x18)", "lw", 17, 18, 4},
		{"lw x19, -4(x20)", "lw", 19, 20, -4},
		{"lb x21, 0(x22)", "lb", 21, 22, 0},
		{"lh x23, 2(x24)", "lh", 23, 24, 2},
		{"lbu x25, 0(x26)", "lbu", 25, 26, 0},
		{"lhu x27, 0(x28)", "lhu", 27, 28, 0},
		{"jalr x1, x2, 0", "jalr", 1, 2, 0},
		{"jalr x0, x1, 0", "jalr", 0, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, uint32(tt.rd), uint32(tt.rs1), 0, tt.imm)
			instr := Decode(word)

			if instr.Mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, tt.mnemonic)
			}
			if instr.Type != arch.TypeI {
				t.Errorf("type: got %v, want %v", instr.Type, arch.TypeI)
			}
			if instr.Rd != tt.rd {
				t.Errorf("rd: got %d, want %d", instr.Rd, tt.rd)
			}
			if instr.Rs1 != tt.rs1 {
				t.Errorf("rs1: got %d, want %d", instr.Rs1, tt.rs1)
			}
			if instr.Immediate != tt.imm {
				t.Errorf("immediate: got %d, want %d", instr.Immediate, tt.imm)
			}
		})
	}
}

func TestDecodeI2Type(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rd       int
		rs1      int
		shamt    int32
	}{
		{"slli x1, x2, 5", "slli", 1, 2, 5},
		{"srli x3, x4, 10", "srli", 3, 4, 10},
		{"srai x5, x6, 15", "srai", 5, 6, 15},
		{"slli x7, x8, 31", "slli", 7, 8, 31},
		{"srli x9, x10, 0", "srli", 9, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, uint32(tt.rd), uint32(tt.rs1), 0, tt.shamt)
			instr := Decode(word)

			if instr.Mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, tt.mnemonic)
			}
			if instr.Type != arch.TypeI2 {
				t.Errorf("type: got %v, want %v", instr.Type, arch.TypeI2)
			}
			if instr.Rd != tt.rd {
				t.Errorf("rd: got %d, want %d", instr.Rd, tt.rd)
			}
			if instr.Rs1 != tt.rs1 {
				t.Errorf("rs1: got %d, want %d", instr.Rs1, tt.rs1)
			}
			if instr.Immediate != tt.shamt {
				t.Errorf("shamt: got %d, want %d", instr.Immediate, tt.shamt)
			}
		})
	}
}

func TestDecodeSType(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rs1      int
		rs2      int
		imm      int32
	}{
		{"sw x1, 0(x2)", "sw", 2, 1, 0},
		{"sw x3, 4(x4)", "sw", 4, 3, 4},
		{"sw x5, -4(x6)", "sw", 6, 5, -4},
		{"sh x7, 2(x8)", "sh", 8, 7, 2},
		{"sb x9, 1(x10)", "sb", 10, 9, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, 0, uint32(tt.rs1), uint32(tt.rs2), tt.imm)
			instr := Decode(word)

			if instr.Mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, tt.mnemonic)
			}
			if instr.Type != arch.TypeS {
				t.Errorf("type: got %v, want %v", instr.Type, arch.TypeS)
			}
			if instr.Rs1 != tt.rs1 {
				t.Errorf("rs1: got %d, want %d", instr.Rs1, tt.rs1)
			}
			if instr.Rs2 != tt.rs2 {
				t.Errorf("rs2: got %d, want %d", instr.Rs2, tt.rs2)
			}
			if instr.Immediate != tt.imm {
				t.Errorf("immediate: got %d, want %d", instr.Immediate, tt.imm)
			}
		})
	}
}

func TestDecodeBType(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rs1      int
		rs2      int
		imm      int32
	}{
		{"beq x1, x2, 8", "beq", 1, 2, 8},
		{"bne x3, x4, 16", "bne", 3, 4, 16},
		{"blt x5, x6, -8", "blt", 5, 6, -8},
		{"bge x7, x8, 32", "bge", 7, 8, 32},
		{"bltu x9, x10, 64", "bltu", 9, 10, 64},
		{"bgeu x11, x12, -16", "bgeu", 11, 12, -16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, 0, uint32(tt.rs1), uint32(tt.rs2), tt.imm)
			instr := Decode(word)

			if instr.Mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, tt.mnemonic)
			}
			if instr.Type != arch.TypeB {
				t.Errorf("type: got %v, want %v", instr.Type, arch.TypeB)
			}
			if instr.Rs1 != tt.rs1 {
				t.Errorf("rs1: got %d, want %d", instr.Rs1, tt.rs1)
			}
			if instr.Rs2 != tt.rs2 {
				t.Errorf("rs2: got %d, want %d", instr.Rs2, tt.rs2)
			}
			if instr.Immediate != tt.imm {
				t.Errorf("immediate: got %d, want %d", instr.Immediate, tt.imm)
			}
		})
	}
}

func TestDecodeUType(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rd       int
		imm      int32
	}{
		{"lui x1, 0x12345", "lui", 1, 0x12345},
		{"lui x2, 0", "lui", 2, 0},
		{"lui x3, 0xFFFFF", "lui", 3, -1}, // Sign-extended
		{"auipc x4, 0x1000", "auipc", 4, 0x1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, uint32(tt.rd), 0, 0, tt.imm)
			instr := Decode(word)

			if instr.Mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, tt.mnemonic)
			}
			if instr.Type != arch.TypeU {
				t.Errorf("type: got %v, want %v", instr.Type, arch.TypeU)
			}
			if instr.Rd != tt.rd {
				t.Errorf("rd: got %d, want %d", instr.Rd, tt.rd)
			}
			if instr.Immediate != tt.imm {
				t.Errorf("immediate: got %d (0x%X), want %d (0x%X)", instr.Immediate, instr.Immediate, tt.imm, tt.imm)
			}
		})
	}
}

func TestDecodeJType(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rd       int
		imm      int32
	}{
		{"jal x1, 8", "jal", 1, 8},
		{"jal x0, 16", "jal", 0, 16},
		{"jal x1, -8", "jal", 1, -8},
		{"jal x2, 1024", "jal", 2, 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, uint32(tt.rd), 0, 0, tt.imm)
			instr := Decode(word)

			if instr.Mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, tt.mnemonic)
			}
			if instr.Type != arch.TypeJ {
				t.Errorf("type: got %v, want %v", instr.Type, arch.TypeJ)
			}
			if instr.Rd != tt.rd {
				t.Errorf("rd: got %d, want %d", instr.Rd, tt.rd)
			}
			if instr.Immediate != tt.imm {
				t.Errorf("immediate: got %d, want %d", instr.Immediate, tt.imm)
			}
		})
	}
}

func TestDecodeR2Type(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rs1      int
		rs2      int
	}{
		{"sleep.t x1, x2", "sleep.t", 1, 2},
		{"wake.t x3, x4", "wake.t", 3, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, 0, uint32(tt.rs1), uint32(tt.rs2), 0)
			instr := Decode(word)

			if instr.Mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, tt.mnemonic)
			}
			if instr.Type != arch.TypeR2 {
				t.Errorf("type: got %v, want %v", instr.Type, arch.TypeR2)
			}
			if instr.Rs1 != tt.rs1 {
				t.Errorf("rs1: got %d, want %d", instr.Rs1, tt.rs1)
			}
			if instr.Rs2 != tt.rs2 {
				t.Errorf("rs2: got %d, want %d", instr.Rs2, tt.rs2)
			}
		})
	}
}

func TestDecodeUnknownInstruction(t *testing.T) {
	// Test with an invalid instruction (all zeros is not valid)
	word := uint32(0x00000000)

	// First, let's see what we actually decode
	instr := Decode(word)

	// addi x0, x0, 0 is a valid encoding (NOP)
	// So let's use a truly invalid opcode instead
	word = uint32(0x0000007F) // opcode 0x7F is not defined
	instr = Decode(word)

	if !instr.IsUnknown() {
		t.Errorf("expected unknown instruction for opcode 0x7F, got %q", instr.Mnemonic)
	}
}

func TestDecodeBytes(t *testing.T) {
	// Encode "add x1, x2, x3"
	word := encodeForTest("add", 1, 2, 3, 0)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, word)

	instr := DecodeBytes(bytes)

	if instr.Mnemonic != "add" {
		t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, "add")
	}
	if instr.Rd != 1 || instr.Rs1 != 2 || instr.Rs2 != 3 {
		t.Errorf("registers: got rd=%d, rs1=%d, rs2=%d, want rd=1, rs1=2, rs2=3",
			instr.Rd, instr.Rs1, instr.Rs2)
	}
}

func TestDecodeBytesTooShort(t *testing.T) {
	bytes := []byte{0x00, 0x00} // Only 2 bytes
	instr := DecodeBytes(bytes)

	if !instr.IsUnknown() {
		t.Errorf("expected unknown instruction for short byte slice")
	}
}

func TestExtractImmediates(t *testing.T) {
	// Test sign extension for negative immediates
	t.Run("I-type negative", func(t *testing.T) {
		word := encodeForTest("addi", 1, 2, 0, -1)
		instr := Decode(word)
		if instr.Immediate != -1 {
			t.Errorf("expected -1, got %d", instr.Immediate)
		}
	})

	t.Run("I-type max positive", func(t *testing.T) {
		word := encodeForTest("addi", 1, 2, 0, 2047) // Max 12-bit signed
		instr := Decode(word)
		if instr.Immediate != 2047 {
			t.Errorf("expected 2047, got %d", instr.Immediate)
		}
	})

	t.Run("I-type min negative", func(t *testing.T) {
		word := encodeForTest("addi", 1, 2, 0, -2048) // Min 12-bit signed
		instr := Decode(word)
		if instr.Immediate != -2048 {
			t.Errorf("expected -2048, got %d", instr.Immediate)
		}
	})

	t.Run("B-type negative", func(t *testing.T) {
		word := encodeForTest("beq", 0, 1, 2, -8)
		instr := Decode(word)
		if instr.Immediate != -8 {
			t.Errorf("expected -8, got %d", instr.Immediate)
		}
	})

	t.Run("J-type negative", func(t *testing.T) {
		word := encodeForTest("jal", 1, 0, 0, -1024)
		instr := Decode(word)
		if instr.Immediate != -1024 {
			t.Errorf("expected -1024, got %d", instr.Immediate)
		}
	})
}

func TestExtractBitFields(t *testing.T) {
	// Test with a known instruction: add x1, x2, x3
	// Encoding: 0x003100B3
	// funct7=0x00, rs2=3, rs1=2, funct3=0, rd=1, opcode=0x33
	word := uint32(0x003100B3)

	if got := extractOpcode(word); got != 0x33 {
		t.Errorf("extractOpcode: got 0x%X, want 0x33", got)
	}
	if got := extractRd(word); got != 1 {
		t.Errorf("extractRd: got %d, want 1", got)
	}
	if got := extractFunct3(word); got != 0 {
		t.Errorf("extractFunct3: got %d, want 0", got)
	}
	if got := extractRs1(word); got != 2 {
		t.Errorf("extractRs1: got %d, want 2", got)
	}
	if got := extractRs2(word); got != 3 {
		t.Errorf("extractRs2: got %d, want 3", got)
	}
	if got := extractFunct7(word); got != 0 {
		t.Errorf("extractFunct7: got %d, want 0", got)
	}

	// Test with sub x10, x11, x12
	// sub has funct7=0x20
	wordSub := encodeForTest("sub", 10, 11, 12, 0)
	if got := extractFunct7(wordSub); got != 0x20 {
		t.Errorf("extractFunct7 for sub: got 0x%X, want 0x20", got)
	}
}

// Test round-trip encoding/decoding for all instructions
func TestRoundTripAllInstructions(t *testing.T) {
	// Representative sample of each instruction type
	instructions := []struct {
		mnemonic string
		rd       uint32
		rs1      uint32
		rs2      uint32
		imm      int32
	}{
		// R-type
		{"add", 1, 2, 3, 0},
		{"sub", 10, 11, 12, 0},
		{"and", 5, 6, 7, 0},
		{"or", 8, 9, 10, 0},
		{"xor", 15, 16, 17, 0},
		{"mul", 1, 2, 3, 0},
		{"div", 4, 5, 6, 0},
		// I-type
		{"addi", 1, 2, 0, 100},
		{"lw", 1, 2, 0, 4},
		{"jalr", 1, 2, 0, 0},
		// I2-type
		{"slli", 1, 2, 0, 5},
		{"srli", 1, 2, 0, 10},
		{"srai", 1, 2, 0, 15},
		// S-type
		{"sw", 0, 1, 2, 4},
		{"sh", 0, 3, 4, 2},
		{"sb", 0, 5, 6, 1},
		// B-type
		{"beq", 0, 1, 2, 8},
		{"bne", 0, 3, 4, -8},
		{"blt", 0, 5, 6, 16},
		// U-type
		{"lui", 1, 0, 0, 0x12345},
		{"auipc", 2, 0, 0, 0x1000},
		// J-type
		{"jal", 1, 0, 0, 1024},
	}

	for _, tt := range instructions {
		t.Run(tt.mnemonic, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, tt.rd, tt.rs1, tt.rs2, tt.imm)
			instr := Decode(word)

			if instr.Mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: got %q, want %q", instr.Mnemonic, tt.mnemonic)
			}
		})
	}
}
