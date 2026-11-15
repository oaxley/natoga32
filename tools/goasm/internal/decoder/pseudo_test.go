// Unit tests for pseudo-instruction recognition
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package decoder

import (
	"testing"
)

func TestPseudoNop(t *testing.T) {
	pm := NewPseudoMatcher()

	// addi x0, x0, 0 => nop
	word := encodeForTest("addi", 0, 0, 0, 0)
	instr := Decode(word)
	mnemonic, operands, matched := pm.Match(instr, 0, true)

	if !matched {
		t.Error("expected nop to be matched")
	}
	if mnemonic != "nop" {
		t.Errorf("got mnemonic %q, want %q", mnemonic, "nop")
	}
	if operands != "" {
		t.Errorf("got operands %q, want empty string", operands)
	}
}

// TestPseudoLiNotMatched verifies that "addi rd, x0, imm" is NOT converted to "li"
// because the assembler's "li" pseudo may expand to lui+addi for boundary values.
func TestPseudoLiNotMatched(t *testing.T) {
	pm := NewPseudoMatcher()

	tests := []struct {
		name string
		rd   uint32
		imm  int32
	}{
		{"li positive", 1, 10},
		{"li negative", 10, -5},
		{"li boundary", 5, -2048},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// addi rd, x0, imm should NOT match as "li"
			word := encodeForTest("addi", tt.rd, 0, 0, tt.imm)
			instr := Decode(word)
			_, _, matched := pm.Match(instr, 0, true)

			if matched {
				t.Error("addi rd, x0, imm should NOT be matched as li")
			}
		})
	}
}

func TestPseudoMv(t *testing.T) {
	pm := NewPseudoMatcher()

	// addi rd, rs1, 0 => mv rd, rs1
	tests := []struct {
		name    string
		rd      uint32
		rs1     uint32
		wantOps string
	}{
		{"mv t0, ra", 5, 1, "t0, ra"},
		{"mv a0, sp", 10, 2, "a0, sp"},
		{"mv t0, zero", 5, 0, "t0, zero"}, // addi rd, x0, 0 is also mv rd, zero
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest("addi", tt.rd, tt.rs1, 0, 0)
			instr := Decode(word)
			mnemonic, operands, matched := pm.Match(instr, 0, true)

			if !matched {
				t.Error("expected mv to be matched")
			}
			if mnemonic != "mv" {
				t.Errorf("got mnemonic %q, want %q", mnemonic, "mv")
			}
			if operands != tt.wantOps {
				t.Errorf("got operands %q, want %q", operands, tt.wantOps)
			}
		})
	}
}

func TestPseudoRet(t *testing.T) {
	pm := NewPseudoMatcher()

	// jalr x0, ra, 0 => ret
	word := encodeForTest("jalr", 0, 1, 0, 0)
	instr := Decode(word)
	mnemonic, operands, matched := pm.Match(instr, 0, true)

	if !matched {
		t.Error("expected ret to be matched")
	}
	if mnemonic != "ret" {
		t.Errorf("got mnemonic %q, want %q", mnemonic, "ret")
	}
	if operands != "" {
		t.Errorf("got operands %q, want empty string", operands)
	}
}

func TestPseudoJump(t *testing.T) {
	pm := NewPseudoMatcher()

	tests := []struct {
		name       string
		rd         uint32
		imm        int32
		addr       uint32
		wantMnem   string
		wantTarget string
	}{
		{"j forward", 0, 8, 0x1000, "j", "0x1008"},
		{"j backward", 0, -8, 0x1000, "j", "0xFF8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// jal x0, offset => j target
			word := encodeForTest("jal", tt.rd, 0, 0, tt.imm)
			instr := Decode(word)
			mnemonic, operands, matched := pm.Match(instr, tt.addr, true)

			if !matched {
				t.Error("expected j to be matched")
			}
			if mnemonic != tt.wantMnem {
				t.Errorf("got mnemonic %q, want %q", mnemonic, tt.wantMnem)
			}
			if operands != tt.wantTarget {
				t.Errorf("got operands %q, want %q", operands, tt.wantTarget)
			}
		})
	}
}

// TestPseudoCallNotMatched verifies that "jal ra, offset" is NOT converted to "call"
// because the assembler's "call" pseudo expands to auipc+jalr (2 instructions).
func TestPseudoCallNotMatched(t *testing.T) {
	pm := NewPseudoMatcher()

	// jal ra, offset should NOT match as "call"
	word := encodeForTest("jal", 1, 0, 0, 100)
	instr := Decode(word)
	_, _, matched := pm.Match(instr, 0x1000, true)

	if matched {
		t.Error("jal ra, offset should NOT be matched as call")
	}
}

func TestPseudoJr(t *testing.T) {
	pm := NewPseudoMatcher()

	// jalr x0, rs1, 0 => jr rs1 (where rs1 != ra, otherwise it's ret)
	word := encodeForTest("jalr", 0, 10, 0, 0) // jalr x0, a0, 0
	instr := Decode(word)
	mnemonic, operands, matched := pm.Match(instr, 0, true)

	if !matched {
		t.Error("expected jr to be matched")
	}
	if mnemonic != "jr" {
		t.Errorf("got mnemonic %q, want %q", mnemonic, "jr")
	}
	if operands != "a0" {
		t.Errorf("got operands %q, want %q", operands, "a0")
	}
}

func TestPseudoBranchZero(t *testing.T) {
	pm := NewPseudoMatcher()

	tests := []struct {
		name     string
		mnemonic string
		rs1      uint32
		rs2      uint32
		imm      int32
		addr     uint32
		wantMnem string
		wantOps  string
	}{
		// beq rs1, x0, offset => beqz rs1, target
		{"beqz", "beq", 5, 0, 8, 0x1000, "beqz", "t0, 0x1008"},
		// bne rs1, x0, offset => bnez rs1, target
		{"bnez", "bne", 10, 0, 16, 0x2000, "bnez", "a0, 0x2010"},
		// blt x0, rs2, offset => bgtz rs2, target
		{"bgtz", "blt", 0, 5, 32, 0x3000, "bgtz", "t0, 0x3020"},
		// bge rs1, x0, offset => bgez rs1, target
		{"bgez", "bge", 6, 0, -16, 0x4000, "bgez", "t1, 0x3FF0"},
		// blt rs1, x0, offset => bltz rs1, target
		{"bltz", "blt", 7, 0, 64, 0x5000, "bltz", "t2, 0x5040"},
		// bge x0, rs2, offset => blez rs2, target
		{"blez", "bge", 0, 8, -32, 0x6000, "blez", "s0, 0x5FE0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, 0, tt.rs1, tt.rs2, tt.imm)
			instr := Decode(word)
			mnemonic, operands, matched := pm.Match(instr, tt.addr, true)

			if !matched {
				t.Errorf("expected %s to be matched", tt.wantMnem)
			}
			if mnemonic != tt.wantMnem {
				t.Errorf("got mnemonic %q, want %q", mnemonic, tt.wantMnem)
			}
			if operands != tt.wantOps {
				t.Errorf("got operands %q, want %q", operands, tt.wantOps)
			}
		})
	}
}

func TestPseudoBranchSwapped(t *testing.T) {
	pm := NewPseudoMatcher()

	tests := []struct {
		name     string
		mnemonic string
		rs1      uint32
		rs2      uint32
		imm      int32
		addr     uint32
		wantMnem string
		wantOps  string
	}{
		// blt rs2, rs1, offset => bgt rs1, rs2, target
		{"bgt", "blt", 6, 5, 8, 0x1000, "bgt", "t0, t1, 0x1008"},
		// bge rs2, rs1, offset => ble rs1, rs2, target
		{"ble", "bge", 11, 10, 16, 0x2000, "ble", "a0, a1, 0x2010"},
		// bltu rs2, rs1, offset => bgtu rs1, rs2, target
		{"bgtu", "bltu", 13, 12, 32, 0x3000, "bgtu", "a2, a3, 0x3020"},
		// bgeu rs2, rs1, offset => bleu rs1, rs2, target
		{"bleu", "bgeu", 15, 14, -16, 0x4000, "bleu", "a4, a5, 0x3FF0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, 0, tt.rs1, tt.rs2, tt.imm)
			instr := Decode(word)
			mnemonic, operands, matched := pm.Match(instr, tt.addr, true)

			if !matched {
				t.Errorf("expected %s to be matched", tt.wantMnem)
			}
			if mnemonic != tt.wantMnem {
				t.Errorf("got mnemonic %q, want %q", mnemonic, tt.wantMnem)
			}
			if operands != tt.wantOps {
				t.Errorf("got operands %q, want %q", operands, tt.wantOps)
			}
		})
	}
}

func TestPseudoSetConditionally(t *testing.T) {
	pm := NewPseudoMatcher()

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		rs2      uint32
		imm      int32
		wantMnem string
		wantOps  string
	}{
		// sltu rd, x0, rs1 => snez rd, rs1
		{"snez", "sltu", 5, 0, 10, 0, "snez", "t0, a0"},
		// sltiu rd, rs1, 1 => seqz rd, rs1
		{"seqz", "sltiu", 6, 11, 0, 1, "seqz", "t1, a1"},
		// slt rd, rs1, x0 => sltz rd, rs1
		{"sltz", "slt", 7, 12, 0, 0, "sltz", "t2, a2"},
		// slt rd, x0, rs1 => sgtz rd, rs1
		{"sgtz", "slt", 8, 0, 13, 0, "sgtz", "s0, a3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, tt.rd, tt.rs1, tt.rs2, tt.imm)
			instr := Decode(word)
			mnemonic, operands, matched := pm.Match(instr, 0, true)

			if !matched {
				t.Errorf("expected %s to be matched", tt.wantMnem)
			}
			if mnemonic != tt.wantMnem {
				t.Errorf("got mnemonic %q, want %q", mnemonic, tt.wantMnem)
			}
			if operands != tt.wantOps {
				t.Errorf("got operands %q, want %q", operands, tt.wantOps)
			}
		})
	}
}

func TestPseudoNegNot(t *testing.T) {
	pm := NewPseudoMatcher()

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		rs2      uint32
		imm      int32
		wantMnem string
		wantOps  string
	}{
		// sub rd, x0, rs1 => neg rd, rs1
		{"neg", "sub", 5, 0, 10, 0, "neg", "t0, a0"},
		// xori rd, rs1, -1 => not rd, rs1
		{"not", "xori", 6, 11, 0, -1, "not", "t1, a1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, tt.rd, tt.rs1, tt.rs2, tt.imm)
			instr := Decode(word)
			mnemonic, operands, matched := pm.Match(instr, 0, true)

			if !matched {
				t.Errorf("expected %s to be matched", tt.wantMnem)
			}
			if mnemonic != tt.wantMnem {
				t.Errorf("got mnemonic %q, want %q", mnemonic, tt.wantMnem)
			}
			if operands != tt.wantOps {
				t.Errorf("got operands %q, want %q", operands, tt.wantOps)
			}
		})
	}
}

func TestPseudoNoMatch(t *testing.T) {
	pm := NewPseudoMatcher()

	// Regular instructions that should not be matched as pseudo
	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		rs2      uint32
		imm      int32
	}{
		{"add", "add", 1, 2, 3, 0},
		{"addi non-zero imm", "addi", 1, 2, 0, 10},
		{"jalr with offset", "jalr", 0, 1, 0, 4},
		{"regular beq", "beq", 0, 5, 6, 8},
		{"regular bne", "bne", 0, 10, 11, 16},
		{"regular slt", "slt", 5, 6, 7, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := encodeForTest(tt.mnemonic, tt.rd, tt.rs1, tt.rs2, tt.imm)
			instr := Decode(word)
			_, _, matched := pm.Match(instr, 0x1000, true)

			if matched {
				t.Errorf("did not expect %s to be matched as pseudo", tt.name)
			}
		})
	}
}

func TestPseudoMatcherNumericNames(t *testing.T) {
	pm := NewPseudoMatcher()

	// Test that numeric register names are used when useABI is false
	// nop
	word := encodeForTest("addi", 0, 0, 0, 0)
	instr := Decode(word)
	mnemonic, _, matched := pm.Match(instr, 0, false)
	if !matched || mnemonic != "nop" {
		t.Errorf("nop should match with numeric names")
	}

	// mv
	word = encodeForTest("addi", 5, 10, 0, 0)
	instr = Decode(word)
	mnemonic, operands, matched := pm.Match(instr, 0, false)
	if !matched || mnemonic != "mv" {
		t.Errorf("mv should match with numeric names")
	}
	if operands != "x5, x10" {
		t.Errorf("got operands %q, want %q", operands, "x5, x10")
	}

	// jr
	word = encodeForTest("jalr", 0, 5, 0, 0)
	instr = Decode(word)
	mnemonic, operands, matched = pm.Match(instr, 0, false)
	if !matched || mnemonic != "jr" {
		t.Errorf("jr should match with numeric names")
	}
	if operands != "x5" {
		t.Errorf("got operands %q, want %q", operands, "x5")
	}
}
