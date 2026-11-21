package arch

import (
	"testing"
)

func TestEncodeTypeR(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		rs2      uint32
		expected uint32
	}{
		// add x1, x2, x3 => 0x003100B3
		{"add x1,x2,x3", "add", 1, 2, 3, 0x003100B3},
		// sub x5, x6, x7 => 0x407302B3
		{"sub x5,x6,x7", "sub", 5, 6, 7, 0x407302B3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, ok := LookupOpcode(tt.mnemonic)
			if !ok {
				t.Fatalf("unknown opcode: %s", tt.mnemonic)
			}
			result := EncodeTypeR(op, tt.rd, tt.rs1, tt.rs2)
			got := uint32(result[0]) | uint32(result[1])<<8 | uint32(result[2])<<16 | uint32(result[3])<<24
			if got != tt.expected {
				t.Errorf("got 0x%08X, want 0x%08X", got, tt.expected)
			}
		})
	}
}

func TestEncodeTypeI(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		imm      int32
		expected uint32
	}{
		// addi x1, x0, 10 => 0x00A00093
		{"addi x1,x0,10", "addi", 1, 0, 10, 0x00A00093},
		// addi x2, x3, -1 => 0xFFF18113
		{"addi x2,x3,-1", "addi", 2, 3, -1, 0xFFF18113},
		// lw x5, 4(x10) => 0x00452283
		{"lw x5,4(x10)", "lw", 5, 10, 4, 0x00452283},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, ok := LookupOpcode(tt.mnemonic)
			if !ok {
				t.Fatalf("unknown opcode: %s", tt.mnemonic)
			}
			result := EncodeTypeI(op, tt.rd, tt.rs1, tt.imm)
			got := uint32(result[0]) | uint32(result[1])<<8 | uint32(result[2])<<16 | uint32(result[3])<<24
			if got != tt.expected {
				t.Errorf("got 0x%08X, want 0x%08X", got, tt.expected)
			}
		})
	}
}

func TestEncodeTypeS(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rs1      uint32
		rs2      uint32
		imm      int32
		expected uint32
	}{
		// sw x5, 8(x10) => 0x00552423
		{"sw x5,8(x10)", "sw", 10, 5, 8, 0x00552423},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, ok := LookupOpcode(tt.mnemonic)
			if !ok {
				t.Fatalf("unknown opcode: %s", tt.mnemonic)
			}
			result := EncodeTypeS(op, tt.rs1, tt.rs2, tt.imm)
			got := uint32(result[0]) | uint32(result[1])<<8 | uint32(result[2])<<16 | uint32(result[3])<<24
			if got != tt.expected {
				t.Errorf("got 0x%08X, want 0x%08X", got, tt.expected)
			}
		})
	}
}

func TestEncodeTypeB(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rs1      uint32
		rs2      uint32
		imm      int32
		expected uint32
	}{
		// beq x1, x2, 8 => 0x00208463
		{"beq x1,x2,8", "beq", 1, 2, 8, 0x00208463},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, ok := LookupOpcode(tt.mnemonic)
			if !ok {
				t.Fatalf("unknown opcode: %s", tt.mnemonic)
			}
			result := EncodeTypeB(op, tt.rs1, tt.rs2, tt.imm)
			got := uint32(result[0]) | uint32(result[1])<<8 | uint32(result[2])<<16 | uint32(result[3])<<24
			if got != tt.expected {
				t.Errorf("got 0x%08X, want 0x%08X", got, tt.expected)
			}
		})
	}
}

func TestEncodeTypeU(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		imm      int32
		expected uint32
	}{
		// lui x1, 0x12345 => 0x123450B7
		{"lui x1,0x12345", "lui", 1, 0x12345, 0x123450B7},
		// auipc x2, 0x1000 => 0x01000117
		{"auipc x2,0x1000", "auipc", 2, 0x1000, 0x01000117},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, ok := LookupOpcode(tt.mnemonic)
			if !ok {
				t.Fatalf("unknown opcode: %s", tt.mnemonic)
			}
			result := EncodeTypeU(op, tt.rd, tt.imm)
			got := uint32(result[0]) | uint32(result[1])<<8 | uint32(result[2])<<16 | uint32(result[3])<<24
			if got != tt.expected {
				t.Errorf("got 0x%08X, want 0x%08X", got, tt.expected)
			}
		})
	}
}

func TestEncodeTypeJ(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		imm      int32
		expected uint32
	}{
		// jal x1, 0 => 0x000000EF
		{"jal x1,0", "jal", 1, 0, 0x000000EF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, ok := LookupOpcode(tt.mnemonic)
			if !ok {
				t.Fatalf("unknown opcode: %s", tt.mnemonic)
			}
			result := EncodeTypeJ(op, tt.rd, tt.imm)
			got := uint32(result[0]) | uint32(result[1])<<8 | uint32(result[2])<<16 | uint32(result[3])<<24
			if got != tt.expected {
				t.Errorf("got 0x%08X, want 0x%08X", got, tt.expected)
			}
		})
	}
}

func TestLookupRegister(t *testing.T) {
	tests := []struct {
		name     string
		expected uint32
	}{
		{"x0", 0},
		{"x1", 1},
		{"x31", 31},
		{"zero", 0},
		{"ra", 1},
		{"sp", 2},
		{"fp", 8},
		{"s0", 8},
		{"a0", 10},
		{"t0", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := LookupRegister(tt.name)
			if !ok {
				t.Fatalf("register not found: %s", tt.name)
			}
			if got != tt.expected {
				t.Errorf("got %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestIsInstruction(t *testing.T) {
	validInstructions := []string{
		"add", "sub", "addi", "lui", "auipc", "jal", "jalr",
		"beq", "bne", "lw", "sw", "and", "or", "xor",
	}

	for _, instr := range validInstructions {
		if !IsInstruction(instr) {
			t.Errorf("expected %s to be a valid instruction", instr)
		}
	}

	invalidInstructions := []string{"foo", "bar", "invalid"}
	for _, instr := range invalidInstructions {
		if IsInstruction(instr) {
			t.Errorf("expected %s to be an invalid instruction", instr)
		}
	}
}

func TestIsPseudoInstruction(t *testing.T) {
	pseudos := []string{"li", "la", "mv", "nop", "ret", "call", "j"}
	for _, p := range pseudos {
		if !IsPseudoInstruction(p) {
			t.Errorf("expected %s to be a pseudo-instruction", p)
		}
	}

	notPseudos := []string{"add", "sub", "lui", "jal"}
	for _, p := range notPseudos {
		if IsPseudoInstruction(p) {
			t.Errorf("expected %s to NOT be a pseudo-instruction", p)
		}
	}
}
