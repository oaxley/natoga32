// Unit tests for instruction formatter
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package formatter

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/natoga32/goasm/internal/arch"
	"github.com/natoga32/goasm/internal/decoder"
)

// Helper to encode and decode instructions for testing
func decodeForTest(mnemonic string, rd, rs1, rs2 uint32, imm int32) *decoder.DecodedInstruction {
	bytes, err := arch.Encode(mnemonic, rd, rs1, rs2, imm)
	if err != nil {
		panic("failed to encode test instruction: " + err.Error())
	}
	word := binary.LittleEndian.Uint32(bytes)
	return decoder.Decode(word)
}

func TestFormatRType(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		rs2      uint32
		wantABI  string
		wantNum  string
	}{
		{"add", "add", 1, 2, 3, "add ra, sp, gp", "add x1, x2, x3"},
		{"sub", "sub", 10, 11, 12, "sub a0, a1, a2", "sub x10, x11, x12"},
		{"and", "and", 5, 6, 7, "and t0, t1, t2", "and x5, x6, x7"},
		{"or", "or", 8, 9, 10, "or s0, s1, a0", "or x8, x9, x10"},
		{"xor", "xor", 0, 1, 2, "xor zero, ra, sp", "xor x0, x1, x2"},
		{"mul", "mul", 15, 16, 17, "mul a5, a6, a7", "mul x15, x16, x17"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, tt.rd, tt.rs1, tt.rs2, 0)

			// Test with ABI names
			f.opts.UseABINames = true
			got := f.FormatInstruction(0, instr)
			if got != tt.wantABI {
				t.Errorf("ABI: got %q, want %q", got, tt.wantABI)
			}

			// Test with numeric names
			f.opts.UseABINames = false
			got = f.FormatInstruction(0, instr)
			if got != tt.wantNum {
				t.Errorf("Numeric: got %q, want %q", got, tt.wantNum)
			}
		})
	}
}

func TestFormatR2Type(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name     string
		mnemonic string
		rs1      uint32
		rs2      uint32
		want     string
	}{
		{"sleep.t", "sleep.t", 1, 2, "sleep.t ra, sp"},
		{"wake.t", "wake.t", 10, 11, "wake.t a0, a1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, 0, tt.rs1, tt.rs2, 0)
			got := f.FormatInstruction(0, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatIType(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		imm      int32
		want     string
	}{
		{"addi positive", "addi", 1, 2, 10, "addi ra, sp, 10"},
		{"addi negative", "addi", 1, 2, -10, "addi ra, sp, -10"},
		{"addi zero", "addi", 1, 0, 0, "addi ra, zero, 0"},
		{"slti", "slti", 5, 6, 100, "slti t0, t1, 100"},
		{"xori", "xori", 10, 11, 255, "xori a0, a1, 255"},
		{"ori", "ori", 12, 13, 0x7FF, "ori a2, a3, 2047"},
		{"andi", "andi", 14, 15, 0x123, "andi a4, a5, 291"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, tt.rd, tt.rs1, 0, tt.imm)
			got := f.FormatInstruction(0, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatLoadInstructions(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		imm      int32
		want     string
	}{
		{"lw zero offset", "lw", 1, 2, 0, "lw ra, 0(sp)"},
		{"lw positive offset", "lw", 10, 11, 4, "lw a0, 4(a1)"},
		{"lw negative offset", "lw", 12, 13, -4, "lw a2, -4(a3)"},
		{"lb", "lb", 5, 6, 1, "lb t0, 1(t1)"},
		{"lh", "lh", 7, 8, 2, "lh t2, 2(s0)"},
		{"lbu", "lbu", 9, 10, 0, "lbu s1, 0(a0)"},
		{"lhu", "lhu", 11, 12, 10, "lhu a1, 10(a2)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, tt.rd, tt.rs1, 0, tt.imm)
			got := f.FormatInstruction(0, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatI2Type(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		shamt    int32
		want     string
	}{
		{"slli", "slli", 1, 2, 5, "slli ra, sp, 5"},
		{"srli", "srli", 10, 11, 10, "srli a0, a1, 10"},
		{"srai", "srai", 15, 16, 15, "srai a5, a6, 15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, tt.rd, tt.rs1, 0, tt.shamt)
			got := f.FormatInstruction(0, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatSType(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name     string
		mnemonic string
		rs1      uint32
		rs2      uint32
		imm      int32
		want     string
	}{
		{"sw zero offset", "sw", 2, 1, 0, "sw ra, 0(sp)"},
		{"sw positive offset", "sw", 10, 11, 4, "sw a1, 4(a0)"},
		{"sw negative offset", "sw", 12, 13, -4, "sw a3, -4(a2)"},
		{"sh", "sh", 5, 6, 2, "sh t1, 2(t0)"},
		{"sb", "sb", 7, 8, 1, "sb s0, 1(t2)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, 0, tt.rs1, tt.rs2, tt.imm)
			got := f.FormatInstruction(0, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatBType(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name     string
		mnemonic string
		rs1      uint32
		rs2      uint32
		imm      int32
		addr     uint32
		want     string
	}{
		{"beq forward", "beq", 1, 2, 8, 0x1000, "beq ra, sp, 0x1008"},
		{"bne backward", "bne", 10, 11, -8, 0x1000, "bne a0, a1, 0xFF8"},
		{"blt", "blt", 5, 6, 16, 0x2000, "blt t0, t1, 0x2010"},
		{"bge", "bge", 7, 8, 32, 0x3000, "bge t2, s0, 0x3020"},
		{"bltu", "bltu", 9, 10, 64, 0x4000, "bltu s1, a0, 0x4040"},
		{"bgeu", "bgeu", 11, 12, -16, 0x5000, "bgeu a1, a2, 0x4FF0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, 0, tt.rs1, tt.rs2, tt.imm)
			got := f.FormatInstruction(tt.addr, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatUType(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		imm      int32
		want     string
	}{
		{"lui", "lui", 1, 0x12345, "lui ra, 0x12345"},
		{"lui zero", "lui", 2, 0, "lui sp, 0x0"},
		{"auipc", "auipc", 10, 0x1000, "auipc a0, 0x1000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, tt.rd, 0, 0, tt.imm)
			got := f.FormatInstruction(0, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatJType(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		imm      int32
		addr     uint32
		want     string
	}{
		{"jal forward", "jal", 1, 1024, 0x1000, "jal ra, 0x1400"},
		{"jal backward", "jal", 0, -1024, 0x2000, "jal zero, 0x1C00"},
		{"jal ra", "jal", 1, 8, 0x3000, "jal ra, 0x3008"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, tt.rd, 0, 0, tt.imm)
			got := f.FormatInstruction(tt.addr, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatJALR(t *testing.T) {
	f := New(DefaultOptions())

	tests := []struct {
		name string
		rd   uint32
		rs1  uint32
		imm  int32
		want string
	}{
		{"jalr ret", 0, 1, 0, "jalr zero, ra, 0"},
		{"jalr call", 1, 10, 0, "jalr ra, a0, 0"},
		{"jalr offset", 1, 2, 4, "jalr ra, sp, 4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest("jalr", tt.rd, tt.rs1, 0, tt.imm)
			got := f.FormatInstruction(0, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatSystemInstructions(t *testing.T) {
	f := New(DefaultOptions())

	// ecall
	t.Run("ecall", func(t *testing.T) {
		instr := decodeForTest("ecall", 0, 0, 0, 0)
		got := f.FormatInstruction(0, instr)
		if got != "ecall" {
			t.Errorf("got %q, want %q", got, "ecall")
		}
	})
}

func TestFormatUnknownInstruction(t *testing.T) {
	f := New(DefaultOptions())

	instr := &decoder.DecodedInstruction{
		Mnemonic: "",
		RawWord:  0xDEADBEEF,
		Rd:       decoder.NoRegister,
		Rs1:      decoder.NoRegister,
		Rs2:      decoder.NoRegister,
	}

	got := f.FormatInstruction(0, instr)
	want := ".word 0xDEADBEEF"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatWithFullOutput(t *testing.T) {
	opts := Options{
		UseABINames: true,
		ShowAddress: true,
		ShowHex:     true,
	}
	f := New(opts)

	instr := decodeForTest("add", 1, 2, 3, 0)
	got := f.Format(0x1000, instr)

	// Should contain address, hex, and instruction
	if !strings.Contains(got, "0x00001000:") {
		t.Errorf("missing address in %q", got)
	}
	if !strings.Contains(got, "add ra, sp, gp") {
		t.Errorf("missing instruction in %q", got)
	}
}

func TestFormatWithMinimalOutput(t *testing.T) {
	opts := Options{
		UseABINames: true,
		ShowAddress: false,
		ShowHex:     false,
	}
	f := New(opts)

	instr := decodeForTest("add", 1, 2, 3, 0)
	got := f.Format(0x1000, instr)

	// Should only contain instruction
	want := "add ra, sp, gp"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRegisterName(t *testing.T) {
	tests := []struct {
		reg     int
		useABI  bool
		want    string
	}{
		{0, true, "zero"},
		{0, false, "x0"},
		{1, true, "ra"},
		{1, false, "x1"},
		{2, true, "sp"},
		{2, false, "x2"},
		{10, true, "a0"},
		{10, false, "x10"},
		{31, true, "t6"},
		{31, false, "x31"},
		{-1, true, ""},
		{32, true, ""},
	}

	for _, tt := range tests {
		name := arch.RegisterName(tt.reg, tt.useABI)
		if name != tt.want {
			t.Errorf("RegisterName(%d, %v) = %q, want %q", tt.reg, tt.useABI, name, tt.want)
		}
	}
}

func TestFormatBranchTarget(t *testing.T) {
	tests := []struct {
		addr   uint32
		offset int32
		want   string
	}{
		{0x1000, 8, "0x00001008"},
		{0x1000, -8, "0x00000FF8"},
		{0x2000, 0, "0x00002000"},
		{0xFFFFFFFC, 4, "0x00000000"},
	}

	for _, tt := range tests {
		got := FormatBranchTarget(tt.addr, tt.offset)
		if got != tt.want {
			t.Errorf("FormatBranchTarget(0x%X, %d) = %q, want %q", tt.addr, tt.offset, got, tt.want)
		}
	}
}

func TestCalculateTarget(t *testing.T) {
	tests := []struct {
		addr   uint32
		offset int32
		want   uint32
	}{
		{0x1000, 8, 0x1008},
		{0x1000, -8, 0x0FF8},
		{0x0, 100, 100},
		{0x100, -0x100, 0},
	}

	for _, tt := range tests {
		got := calculateTarget(tt.addr, tt.offset)
		if got != tt.want {
			t.Errorf("calculateTarget(0x%X, %d) = 0x%X, want 0x%X", tt.addr, tt.offset, got, tt.want)
		}
	}
}

func TestFormatWithPseudo(t *testing.T) {
	// Test that pseudo-instructions are formatted correctly when ShowPseudo is enabled
	opts := Options{
		UseABINames: true,
		ShowAddress: false,
		ShowHex:     false,
		ShowPseudo:  true,
	}
	f := New(opts)

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		rs2      uint32
		imm      int32
		addr     uint32
		want     string
	}{
		{"nop", "addi", 0, 0, 0, 0, 0, "nop"},
		{"li", "addi", 1, 0, 0, 10, 0, "li ra, 10"},
		{"mv", "addi", 5, 10, 0, 0, 0, "mv t0, a0"},
		{"ret", "jalr", 0, 1, 0, 0, 0, "ret"},
		{"jr", "jalr", 0, 10, 0, 0, 0, "jr a0"},
		{"j", "jal", 0, 0, 0, 8, 0x1000, "j 0x1008"},
		{"call", "jal", 1, 0, 0, 100, 0x1000, "call 0x1064"},
		{"beqz", "beq", 0, 5, 0, 8, 0x1000, "beqz t0, 0x1008"},
		{"bnez", "bne", 0, 10, 0, 16, 0x2000, "bnez a0, 0x2010"},
		{"neg", "sub", 5, 0, 10, 0, 0, "neg t0, a0"},
		{"not", "xori", 6, 11, 0, -1, 0, "not t1, a1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, tt.rd, tt.rs1, tt.rs2, tt.imm)
			got := f.FormatInstruction(tt.addr, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatWithoutPseudo(t *testing.T) {
	// Test that pseudo-instructions are NOT formatted when ShowPseudo is disabled
	opts := Options{
		UseABINames: true,
		ShowAddress: false,
		ShowHex:     false,
		ShowPseudo:  false, // Disabled
	}
	f := New(opts)

	tests := []struct {
		name     string
		mnemonic string
		rd       uint32
		rs1      uint32
		rs2      uint32
		imm      int32
		addr     uint32
		want     string
	}{
		// Without pseudo, nop should show as addi
		{"nop -> addi", "addi", 0, 0, 0, 0, 0, "addi zero, zero, 0"},
		{"li -> addi", "addi", 1, 0, 0, 10, 0, "addi ra, zero, 10"},
		{"mv -> addi", "addi", 5, 10, 0, 0, 0, "addi t0, a0, 0"},
		{"ret -> jalr", "jalr", 0, 1, 0, 0, 0, "jalr zero, ra, 0"},
		{"j -> jal", "jal", 0, 0, 0, 8, 0x1000, "jal zero, 0x1008"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := decodeForTest(tt.mnemonic, tt.rd, tt.rs1, tt.rs2, tt.imm)
			got := f.FormatInstruction(tt.addr, instr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
