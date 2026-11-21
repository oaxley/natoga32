package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// assembleSource assembles src with the given output format and returns
// the output bytes, the combined goasm output and the run error.
func assembleSource(t *testing.T, goasm, src, format string) ([]byte, string, error) {
	t.Helper()
	dir := t.TempDir()
	input := filepath.Join(dir, "in.s")
	output := filepath.Join(dir, "out.bin")
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(goasm, input, "-f", format, "-o", output).CombinedOutput()
	if err != nil {
		return nil, string(out), err
	}
	data, readErr := os.ReadFile(output)
	if readErr != nil {
		t.Fatal(readErr)
	}
	return data, string(out), nil
}

func TestCsrImmediateEncoding(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		name string
		src  string
		want []byte
	}{
		{"csrwi", "csrwi 0x305, 5", []byte{0x73, 0xd0, 0x52, 0x30}},
		{"csrsi", "csrsi 0x300, 8", []byte{0x73, 0x60, 0x04, 0x30}},
		{"csrci", "csrci 0x344, 31", []byte{0x73, 0xf0, 0x4f, 0x34}},
		{"csrrwi numeric uimm", "csrrwi a0, 3, 0x340", []byte{0x73, 0xd5, 0x01, 0x34}},
		{"csrrsi register-style uimm", "csrrsi a1, x5, 0x304", []byte{0xf3, 0xe5, 0x42, 0x30}},
		{"csrrci symbolic uimm", ".equ BIT, 2\ncsrrci a2, BIT, 0x305", []byte{0x73, 0x76, 0x51, 0x30}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, out, err := assembleSource(t, goasm, ".text\n"+tt.src+"\n", "bin")
			if err != nil {
				t.Fatalf("assembling failed: %v\n%s", err, out)
			}
			if !bytes.Equal(data, tt.want) {
				t.Errorf("%s = % x, want % x", tt.src, data, tt.want)
			}
		})
	}
}

func TestCsrImmediateOutOfRange(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	for _, src := range []string{"csrwi 0x305, 32", "csrrsi a0, -1, 0x300"} {
		_, out, err := assembleSource(t, goasm, ".text\n"+src+"\n", "bin")
		if err == nil {
			t.Errorf("%q: expected an error, assembled fine", src)
			continue
		}
		if !strings.Contains(out, "out of range") {
			t.Errorf("%q: expected 'out of range' in output, got:\n%s", src, out)
		}
	}
}

// System instructions with a fixed funct12 field: ecall (0x000), wfi (0x105), mret (0x302)
func TestSystemInstructionEncoding(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want []byte
	}{
		{"ecall", []byte{0x73, 0x00, 0x00, 0x00}},  // 0x00000073
		{"ebreak", []byte{0x73, 0x00, 0x10, 0x00}}, // 0x00100073
		{"wfi", []byte{0x73, 0x00, 0x50, 0x10}},    // 0x10500073
		{"mret", []byte{0x73, 0x00, 0x20, 0x30}},   // 0x30200073
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			data, out, err := assembleSource(t, goasm, ".rom\n"+tt.src+"\n", "rom")
			if err != nil {
				t.Fatalf("assembling failed: %v\n%s", err, out)
			}
			if !bytes.Equal(data, tt.want) {
				t.Errorf("%s = % x, want % x", tt.src, data, tt.want)
			}
		})
	}
}

// fence pred, succ: pred in bits 27:24, succ in bits 23:20 (i=8, o=4, r=2, w=1).
// fence alone is fence iorw, iorw. fence.i has funct3 001 and no operand.
func TestFenceEncoding(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want []byte
	}{
		{"fence", []byte{0x0f, 0x00, 0xf0, 0x0f}},            // 0x0FF0000F
		{"fence iorw, iorw", []byte{0x0f, 0x00, 0xf0, 0x0f}}, // same word
		{"fence rw, rw", []byte{0x0f, 0x00, 0x30, 0x03}},     // 0x0330000F
		{"fence w, r", []byte{0x0f, 0x00, 0x20, 0x01}},       // 0x0120000F
		{"fence i, o", []byte{0x0f, 0x00, 0x40, 0x08}},       // 0x0840000F
		{"fence wr, ow", []byte{0x0f, 0x00, 0x50, 0x03}},     // any letter order: 0x0350000F
		{"fence 0, w", []byte{0x0f, 0x00, 0x10, 0x00}},       // empty set, as printed by the disassembler: 0x0010000F
		{"fence.i", []byte{0x0f, 0x10, 0x00, 0x00}},          // 0x0000100F
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			data, out, err := assembleSource(t, goasm, ".rom\n"+tt.src+"\n", "rom")
			if err != nil {
				t.Fatalf("assembling failed: %v\n%s", err, out)
			}
			if !bytes.Equal(data, tt.want) {
				t.Errorf("%s = % x, want % x", tt.src, data, tt.want)
			}
		})
	}
}

// Bad fence operands, and operands given to fence.i or ebreak, are errors
func TestFenceInvalid(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want string
	}{
		{"fence rw", "fence expects 0 or 2 operands (pred, succ)"},
		{"fence rw, rw, rw", "fence expects 0 or 2 operands (pred, succ)"},
		{"fence rx, w", "invalid fence set 'rx'"},
		{"fence rr, w", "invalid fence set 'rr'"},
		{"fence 3, w", "invalid fence set"},
		{"fence.i a0", "fence.i takes no operands"},
		{"ebreak 1", "ebreak takes no operands"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			_, out, err := assembleSource(t, goasm, ".rom\n"+tt.src+"\n", "rom")
			if err == nil {
				t.Fatalf("expected an error, assembled fine")
			}
			if !strings.Contains(out, tt.want) {
				t.Errorf("expected %q in output, got:\n%s", tt.want, out)
			}
		})
	}
}

// -O keeps the instruction after a replacement, and only drops a move with no effect
func TestOptimizerEndToEnd(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "in.s")
	output := filepath.Join(dir, "out.rom")
	src := ".rom\nadd t0, t1, zero\nli a0, 1\nli a1, 2\nmv t2, t0\nmv t0, t2\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(goasm, input, "-O", "-f", "rom", "-o", output).CombinedOutput(); err != nil {
		t.Fatalf("assembling failed: %v\n%s", err, out)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{
		0x93, 0x02, 0x03, 0x00, // addi t0, t1, 0 (was add t0, t1, zero)
		0x13, 0x05, 0x10, 0x00, // addi a0, zero, 1
		0x93, 0x05, 0x20, 0x00, // addi a1, zero, 2
		0x93, 0x83, 0x02, 0x00, // addi t2, t0, 0 (the mv t0, t2 after it is dropped)
	}
	if !bytes.Equal(data, want) {
		t.Errorf("-O output = % x, want % x", data, want)
	}
}

// A recursive macro with no end stops with an error; one that ends through .if works
func TestRecursiveMacro(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	errors := []struct {
		name, src, want string
	}{
		{"self", ".macro boom\nboom\n.endm\n.rom\nboom", "macro 'boom' expanded too deeply"},
		{"mutual", ".macro a\nb\n.endm\n.macro b\na\n.endm\n.rom\na", "expanded too deeply"},
	}
	for _, tt := range errors {
		t.Run(tt.name, func(t *testing.T) {
			_, out, err := assembleSource(t, goasm, tt.src+"\n", "rom")
			if err == nil {
				t.Fatalf("expected an error, assembled fine")
			}
			if !strings.Contains(out, tt.want) || !strings.Contains(out, "in.s:") {
				t.Errorf("expected %q with the file name in output, got:\n%s", tt.want, out)
			}
		})
	}

	t.Run("ends through .if", func(t *testing.T) {
		src := ".macro cnt n\n.if \\n > 0\naddi a0, a0, \\n\ncnt \\n - 1\n.endif\n.endm\n.rom\ncnt 3\n"
		data, out, err := assembleSource(t, goasm, src, "rom")
		if err != nil {
			t.Fatalf("assembling failed: %v\n%s", err, out)
		}
		want := []byte{
			0x13, 0x05, 0x35, 0x00, // addi a0, a0, 3
			0x13, 0x05, 0x25, 0x00, // addi a0, a0, 2
			0x13, 0x05, 0x15, 0x00, // addi a0, a0, 1
		}
		if !bytes.Equal(data, want) {
			t.Errorf("% x, want % x", data, want)
		}
	})
}

// Every operator works in branch/jal targets and in immediates, and so does $.
// Each source is padded so that the instruction under test is at 0x10.
func TestExpressionOperands(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	pad := "nop\nnop\nnop\nnop\n"
	tests := []struct {
		src  string
		want uint32
	}{
		// branch targets: beq a0, a1, 0x18 is +8 = 0x00B50463
		{"beq a0, a1, 3 << 3", 0x00B50463},
		{"beq a0, a1, 0x10 | 8", 0x00B50463},
		{"beq a0, a1, 0x1F & 0x18", 0x00B50463},
		{"beq a0, a1, 0x10 ^ 8", 0x00B50463},
		{"beq a0, a1, 0xC0 >> 3", 0x00B50463},
		{"beq a0, a1, 0x118 % 0x100", 0x00B50463},
		{"beq a0, a1, $ + 8", 0x00B50463},
		// jal targets: jal ra, 0x18 is +8 = 0x008000EF
		{"jal ra, 3 << 3", 0x008000EF},
		{"jal 0x10 | 8", 0x008000EF},
		// $ in immediates is the address of the instruction (0x10)
		{"addi a0, a0, $", 0x01050513},
		{"addi a0, a0, $ + 4", 0x01450513},
		{"lui a0, $", 0x00010537},
		{"lw a0, $(sp)", 0x01012503},
		{"sw a0, $(sp)", 0x00A12823},
		// % in immediates
		{"addi a0, a0, 23 % 8", 0x00750513},
		// precedence and grouping
		{"addi a0, a0, 1 + 2 * 3", 0x00750513},
		{"addi a0, a0, (1 + 2) << 2", 0x00C50513},
		{"addi a0, a0, -(2 | 1)", 0xFFD50513},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			data, out, err := assembleSource(t, goasm, ".rom\n"+pad+tt.src+"\n", "rom")
			if err != nil {
				t.Fatalf("assembling failed: %v\n%s", err, out)
			}
			if len(data) != 20 {
				t.Fatalf("expected 20 bytes, got %d", len(data))
			}
			got := uint32(data[16]) | uint32(data[17])<<8 | uint32(data[18])<<16 | uint32(data[19])<<24
			if got != tt.want {
				t.Errorf("0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

// Data directives share the evaluator: every operator works there too
func TestDataDirectiveOperators(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	src := ".rom\n.word 1 << 3, 0xF0 | 0xF, 0xFF & 0x3C, 6 ^ 3, -8 >> 1, 7 % 4\n.byte 0x80 >> 4, $\n"
	data, out, err := assembleSource(t, goasm, src, "rom")
	if err != nil {
		t.Fatalf("assembling failed: %v\n%s", err, out)
	}
	want := []byte{
		0x08, 0, 0, 0, 0xFF, 0, 0, 0, 0x3C, 0, 0, 0,
		0x05, 0, 0, 0, 0xFC, 0xFF, 0xFF, 0xFF, 0x03, 0, 0, 0,
		0x08, 0x19, // $ is the address of the second .byte value (0x19)
	}
	if !bytes.Equal(data, want) {
		t.Errorf("% x, want % x", data, want)
	}
}

// An expression that cannot be computed is an error, never a silent 0
func TestExpressionErrors(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want string
	}{
		{"addi a0, a0, 5 % 0", "modulo by zero"},
		{"beq a0, a1, 8 % 0", "modulo by zero"},
		{"addi a0, a0, 1 << -1", "shift amount"},
		{"beq a0, a1, 1 >> 64", "shift amount"},
		{"addi a0, a0, \"x\"", "cannot be used in an expression"},
		{"beq a0, a1, \"x\"", "cannot be used in an expression"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			_, out, err := assembleSource(t, goasm, ".rom\n"+tt.src+"\n", "rom")
			if err == nil {
				t.Fatalf("expected an error, assembled fine")
			}
			if !strings.Contains(out, tt.want) {
				t.Errorf("expected %q in output, got:\n%s", tt.want, out)
			}
		})
	}
}

// jal label is jal ra, label (RISC-V standard one-operand form)
func TestJalEncoding(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		name string
		src  string
		want []byte
	}{
		// 0x00: nop, 0x04: jal, 0x08: nop, 0x0C: target
		{"one operand forward", "nop\njal target\nnop\ntarget: nop",
			[]byte{0x13, 0, 0, 0, 0xef, 0x00, 0x80, 0x00, 0x13, 0, 0, 0, 0x13, 0, 0, 0}}, // jal ra, +8 = 0x008000EF
		{"one operand same as jal ra", "nop\njal ra, target\nnop\ntarget: nop",
			[]byte{0x13, 0, 0, 0, 0xef, 0x00, 0x80, 0x00, 0x13, 0, 0, 0, 0x13, 0, 0, 0}},
		{"one operand backward", "target: nop\njal target",
			[]byte{0x13, 0, 0, 0, 0xef, 0xf0, 0xdf, 0xff}}, // jal ra, -4 = 0xFFDFF0EF
		{"two operands rd x0", "jal x0, target\ntarget: nop",
			[]byte{0x6f, 0x00, 0x40, 0x00, 0x13, 0, 0, 0}}, // jal zero, +4 = 0x0040006F
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, out, err := assembleSource(t, goasm, ".rom\n"+tt.src+"\n", "rom")
			if err != nil {
				t.Fatalf("assembling failed: %v\n%s", err, out)
			}
			if !bytes.Equal(data, tt.want) {
				t.Errorf("% x, want % x", data, tt.want)
			}
		})
	}
}

// jal takes 1 or 2 operands
func TestJalOperandCount(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	for _, src := range []string{"jal", "jal ra, target, target"} {
		t.Run(src, func(t *testing.T) {
			_, out, err := assembleSource(t, goasm, ".rom\n"+src+"\ntarget: nop\n", "rom")
			if err == nil {
				t.Fatalf("expected an error, assembled fine")
			}
			if !strings.Contains(out, "jal expects 1 or 2 operands") {
				t.Errorf("expected 'jal expects 1 or 2 operands' in output, got:\n%s", out)
			}
		})
	}
}

// Zbb/Zbs immediate shifts: the third operand is a shift amount (0..31), not a register
func TestShiftImmediateEncoding(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want []byte
	}{
		{"rori a0, a1, 7", []byte{0x13, 0xd5, 0x75, 0x60}},   // 0x6075D513
		{"rori a0, a1, 31", []byte{0x13, 0xd5, 0xf5, 0x61}},  // 0x61F5D513
		{"bseti t3, t4, 5", []byte{0x13, 0x9e, 0x5e, 0x28}},  // 0x285E9E13
		{"bclri a2, a3, 10", []byte{0x13, 0x96, 0xa6, 0x48}}, // 0x48A69613
		{"binvi a7, s0, 15", []byte{0x93, 0x18, 0xf4, 0x68}}, // 0x68F41893
		{"bexti s4, s5, 20", []byte{0x13, 0xda, 0x4a, 0x49}}, // 0x494ADA13
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			data, out, err := assembleSource(t, goasm, ".rom\n"+tt.src+"\n", "rom")
			if err != nil {
				t.Fatalf("assembling failed: %v\n%s", err, out)
			}
			if !bytes.Equal(data, tt.want) {
				t.Errorf("%s = % x, want % x", tt.src, data, tt.want)
			}
		})
	}
}

// A shift amount outside 0..31, or a register in its place, is an error
func TestShiftImmediateInvalid(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want string
	}{
		{"rori a0, a1, 32", "out of range"},
		{"bseti a0, a1, -1", "out of range"},
		{"slli a0, a1, 32", "out of range"},
		{"binvi a0, a1, a2", "shift amount"},
		{"srai a0, a1, t0", "shift amount"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			_, out, err := assembleSource(t, goasm, ".rom\n"+tt.src+"\n", "rom")
			if err == nil {
				t.Fatalf("expected an error, assembled fine")
			}
			if !strings.Contains(out, tt.want) {
				t.Errorf("expected %q in output, got:\n%s", tt.want, out)
			}
		})
	}
}

// zext.h rd, rs1 is pack rd, rs1, x0; pack keeps its three operands
func TestZextHAndPackEncoding(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want []byte
	}{
		{"zext.h a0, a1", []byte{0x33, 0xc5, 0x05, 0x08}},    // 0x0805C533
		{"pack a0, a1, x0", []byte{0x33, 0xc5, 0x05, 0x08}},  // same word
		{"pack a0, a1, a2", []byte{0x33, 0xc5, 0xc5, 0x08}},  // 0x08C5C533
		{"packh a0, a1, a2", []byte{0x33, 0xf5, 0xc5, 0x08}}, // 0x08C5F533
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			data, out, err := assembleSource(t, goasm, ".rom\n"+tt.src+"\n", "rom")
			if err != nil {
				t.Fatalf("assembling failed: %v\n%s", err, out)
			}
			if !bytes.Equal(data, tt.want) {
				t.Errorf("%s = % x, want % x", tt.src, data, tt.want)
			}
		})
	}
}

// Unary instructions (zext.h, clz, ...) take exactly rd and rs1
func TestUnaryOperandCount(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	for _, src := range []string{"zext.h a0, a1, a2", "clz a0, a1, a2", "rev8 a0", "sext.b a0, a1, a2"} {
		t.Run(src, func(t *testing.T) {
			_, out, err := assembleSource(t, goasm, ".rom\n"+src+"\n", "rom")
			if err == nil {
				t.Fatalf("expected an error, assembled fine")
			}
			if !strings.Contains(out, "expects 2 operands (rd, rs1)") {
				t.Errorf("expected 'expects 2 operands (rd, rs1)' in output, got:\n%s", out)
			}
		})
	}
}

func TestCharLiteralImmediates(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	src := ".text\n" +
		"addi t5, t4, '0'\n" +
		"addi a0, a0, 'A' - 10\n" +
		"li a0, 'A'\n" +
		".byte 'Z', 0, 0, 0\n"
	data, out, err := assembleSource(t, goasm, src, "bin")
	if err != nil {
		t.Fatalf("assembling failed: %v\n%s", err, out)
	}
	want := []byte{
		0x13, 0x8f, 0x0e, 0x03, // addi t5, t4, 48
		0x13, 0x05, 0x75, 0x03, // addi a0, a0, 55
		0x13, 0x05, 0x10, 0x04, // addi a0, x0, 65 (li fits in one instruction)
		0x5a, 0x00, 0x00, 0x00, // .byte 'Z'
	}
	if !bytes.Equal(data, want) {
		t.Errorf("char literals = % x, want % x", data, want)
	}
}

// GNU-style split pair: %pcrel_lo references the label of the auipc,
// which does not have to be the immediately preceding instruction.
func TestPCRelLoSplitPair(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	src := ".rom\n" +
		"here: auipc t0, %pcrel_hi(target)\n" + // 0x00
		"nop\n" + // 0x04
		"addi t0, t0, %pcrel_lo(here)\n" + // 0x08
		"lw t1, %pcrel_lo(here)(t0)\n" + // 0x0c
		"nop\n" + // 0x10
		"target: nop\n" // 0x14
	data, out, err := assembleSource(t, goasm, src, "rom")
	if err != nil {
		t.Fatalf("assembling failed: %v\n%s", err, out)
	}
	want := []byte{
		0x97, 0x02, 0x00, 0x00, // auipc t0, 0
		0x13, 0x00, 0x00, 0x00, // nop
		0x93, 0x82, 0x42, 0x01, // addi t0, t0, 20
		0x03, 0xa3, 0x42, 0x01, // lw   t1, 20(t0)
		0x13, 0x00, 0x00, 0x00, // nop
		0x13, 0x00, 0x00, 0x00, // target: nop
	}
	if !bytes.Equal(data, want) {
		t.Errorf("split pair = % x, want % x", data, want)
	}
}

// la to a function whose first instruction is itself an auipc (call):
// the adjacent pair must win over the "label of an auipc" interpretation.
func TestPCRelLoTargetStartsWithAuipc(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	// func2 - func (12) differs from func - la (8), so a wrong pairing shows
	src := ".rom\nla t0, func\nfunc: call func2\nnop\nfunc2: nop\n"
	data, out, err := assembleSource(t, goasm, src, "rom")
	if err != nil {
		t.Fatalf("assembling failed: %v\n%s", err, out)
	}
	want := []byte{
		0x97, 0x02, 0x00, 0x00, // auipc t0, 0
		0x93, 0x82, 0x82, 0x00, // addi  t0, t0, 8  (0x00 + 8 = func)
		0x97, 0x00, 0x00, 0x00, // func: auipc ra, 0
		0xe7, 0x80, 0xc0, 0x00, // jalr  ra, ra, 12 (0x08 + 12 = func2)
		0x13, 0x00, 0x00, 0x00, // nop
		0x13, 0x00, 0x00, 0x00, // func2: nop
	}
	if !bytes.Equal(data, want) {
		t.Errorf("la/call = % x, want % x", data, want)
	}
}

// %pcrel_lo(sym) without an adjacent matching auipc, where sym is not an
// auipc label, cannot be resolved and must be an error.
func TestPCRelLoUnpairedIsAnError(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	src := ".rom\nauipc t0, %pcrel_hi(target)\nnop\naddi t0, t0, %pcrel_lo(target)\ntarget: nop\n"
	_, out, err := assembleSource(t, goasm, src, "rom")
	if err == nil {
		t.Fatalf("expected an error for an unpaired %%pcrel_lo, assembled fine:\n%s", out)
	}
	if !strings.Contains(out, "%pcrel_lo") {
		t.Errorf("expected a %%pcrel_lo error, got:\n%s", out)
	}
}

func TestUnknownFormatCreatesNoFile(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "in.s")
	output := filepath.Join(dir, "out.xyz")
	if err := os.WriteFile(input, []byte(".text\nnop\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(goasm, input, "-f", "xyz", "-o", output).CombinedOutput()
	if err == nil {
		t.Fatalf("expected goasm to fail on an unknown format, it succeeded:\n%s", out)
	}
	if !strings.Contains(string(out), "unknown output format") {
		t.Errorf("expected 'unknown output format' in output, got:\n%s", out)
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Errorf("output file %s should not have been created", output)
	}
}
