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
		{"ecall", []byte{0x73, 0x00, 0x00, 0x00}}, // 0x00000073
		{"wfi", []byte{0x73, 0x00, 0x50, 0x10}},   // 0x10500073
		{"mret", []byte{0x73, 0x00, 0x20, 0x30}},  // 0x30200073
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
