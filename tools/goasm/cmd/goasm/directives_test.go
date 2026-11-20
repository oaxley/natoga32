package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

func TestOrgDirectivePass1(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000, 0x5000)

	// Start in .text at 0x1000
	// Emit 4 bytes first
	sections.Current().Size += 4 // Simulate 4 bytes emitted

	// .org 0x1010 should add 12 bytes of padding (0x1010 - 0x1004 = 12)
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1010}},
	}

	err := handleDirective(dir, symbols, sections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSize := int64(4 + 12) // original 4 + 12 padding
	if sections.Current().Size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, sections.Current().Size)
	}
}

func TestOrgDirectivePass1AtCurrentAddress(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000, 0x5000)

	// .org at current address should be a no-op
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1000}},
	}

	err := handleDirective(dir, symbols, sections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sections.Current().Size != 0 {
		t.Errorf("expected size 0, got %d", sections.Current().Size)
	}
}

func TestOrgDirectivePass1BackwardError(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000, 0x5000)

	// Emit 16 bytes first
	sections.Current().Size += 16 // Current address is now 0x1010

	// .org 0x1008 should fail (trying to go backward)
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1008}},
	}

	err := handleDirective(dir, symbols, sections)
	if err == nil {
		t.Error("expected error for backward .org, got nil")
	}
}

func TestOrgDirectivePass3(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000, 0x5000)

	// Emit 4 bytes first
	sections.EmitWord(0xDEADBEEF)

	// .org 0x1010 should emit 12 zero bytes
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1010}},
	}

	err := emitDirective(dir, symbols, sections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSize := int64(4 + 12)
	if sections.Current().Size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, sections.Current().Size)
	}

	// Verify the padding bytes are all zeros
	data := sections.Current().Data
	for i := 4; i < 16; i++ {
		if data[i] != 0 {
			t.Errorf("expected zero at offset %d, got 0x%02X", i, data[i])
		}
	}

	// Verify current address
	if sections.CurrentAddress() != 0x1010 {
		t.Errorf("expected current address 0x1010, got 0x%X", sections.CurrentAddress())
	}
}

func TestOrgDirectivePass3BackwardError(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000, 0x5000)

	// Emit 16 bytes first
	for i := 0; i < 4; i++ {
		sections.EmitWord(0x11111111)
	}

	// .org 0x1008 should fail (trying to go backward)
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1008}},
	}

	err := emitDirective(dir, symbols, sections)
	if err == nil {
		t.Error("expected error for backward .org, got nil")
	}
}

func TestUnknownDirectiveIsAnError(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000, 0x5000)

	// a typo of .word must not be silently ignored
	dir := &ast.Directive{
		Name: ".wrod",
		Args: []ast.Expression{&ast.Number{Value: 1}},
	}

	err := handleDirective(dir, symbols, sections)
	if err == nil {
		t.Fatal("expected error for unknown directive .wrod, got nil")
	}
	if !strings.Contains(err.Error(), "unknown directive '.wrod'") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRodataDirectiveSwitchesSection(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000, 0x5000)

	dir := &ast.Directive{Name: ".rodata"}

	if err := handleDirective(dir, symbols, sections); err != nil {
		t.Fatalf("handleDirective(.rodata) failed: %v", err)
	}
	if sections.CurrentName() != ".rodata" {
		t.Errorf("handleDirective: expected current section .rodata, got %s", sections.CurrentName())
	}

	sections.SetSection(".text")
	if err := emitDirective(dir, symbols, sections); err != nil {
		t.Fatalf("emitDirective(.rodata) failed: %v", err)
	}
	if sections.CurrentName() != ".rodata" {
		t.Errorf("emitDirective: expected current section .rodata, got %s", sections.CurrentName())
	}
}

// TestAllDocumentedDirectivesAssemble makes sure that rejecting unknown
// directives does not break any directive documented in README.md,
// including the preprocessor ones.
func TestAllDocumentedDirectivesAssemble(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "inc.asm"), []byte("    nop\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "blob.bin"), []byte{1, 2, 3, 4}, 0o644); err != nil {
		t.Fatal(err)
	}

	src := `
.define SIZE 4
.equ ANSWER, 42
.set OTHER, 7
.macro push reg
    addi sp, sp, -4
    sw   \reg, 0(sp)
.endmacro
.global main
.globl helper

.rom
    nop

.text
main:
    push a0
.include "inc.asm"
.ifdef SIZE
    nop
.endif
.ifndef UNDEFINED_SYMBOL
    nop
.endif
.if SIZE == 4
    nop
.else
    addi a0, a0, 1
.endif
.for i = 0, 2
    nop
.endf
helper:
    ret

.data
    .byte 1, 2
    .half 3
    .short 4
    .word 5
    .word .dup 2 0
    .string "abc"
    .asciz "def"
    .ascii "gh"
    .align 2
    .space 4
    .skip 4
    .incbin "blob.bin"
    .org 0x01400040

.rodata
    .string "ro"

.bss
    .space 16
`
	input := filepath.Join(dir, "all.asm")
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-o", filepath.Join(dir, "all.on32"))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("assembling all documented directives failed: %v\n%s", err, out)
	}
}

// TestRodataEndToEnd checks that .rodata labels resolve to the
// .rodata segment of the memory map (0x01200000).
func TestRodataEndToEnd(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "rodata.asm")
	src := ".rodata\nmsg: .word 0x11223344\n.text\nmain: nop\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-d", "5", "-o", filepath.Join(dir, "rodata.on32"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("assembling .rodata failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "msg: 0x01200000") {
		t.Errorf("expected 'msg: 0x01200000' in output, got:\n%s", out)
	}
	if !strings.Contains(string(out), ".rodata: base=0x01200000 size=4") {
		t.Errorf("expected '.rodata: base=0x01200000 size=4' in output, got:\n%s", out)
	}
}

func TestRomFormatEndToEnd(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "boot.s")
	output := filepath.Join(dir, "rombios.sys")
	src := ".rom\n.ascii \"N32R\"\n.org 0x100\nstart: nop\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-f", "rom", "-o", output)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("assembling rom failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0x104 {
		t.Fatalf("rom length = %#x, want 0x104", len(data))
	}
	if string(data[:4]) != "N32R" {
		t.Errorf("magic = %q, want N32R", data[:4])
	}
	for i := 4; i < 0x100; i++ {
		if data[i] != 0 {
			t.Fatalf("byte %#x = %#x, want 0", i, data[i])
		}
	}
	if want := []byte{0x13, 0, 0, 0}; !bytes.Equal(data[0x100:], want) {
		t.Errorf("nop = % x, want % x", data[0x100:], want)
	}
}

func TestRomFormatTooLarge(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "big.s")
	if err := os.WriteFile(input, []byte(".rom\n.space 2097153\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-f", "rom", "-o", filepath.Join(dir, "big.sys"))
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected goasm to fail on oversized ROM, it succeeded:\n%s", out)
	}
	want := "section .rom is 2097153 bytes, exceeds its 2MiB window (0x00000000-0x001FFFFF)"
	if !strings.Contains(string(out), want) {
		t.Errorf("expected %q in output, got:\n%s", want, out)
	}
}

func TestRomFormatWithoutRomSection(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "norom.s")
	output := filepath.Join(dir, "norom.sys")
	if err := os.WriteFile(input, []byte(".text\nstart: nop\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-f", "rom", "-o", output)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected goasm to fail without a .rom section, it succeeded:\n%s", out)
	}
	if !strings.Contains(string(out), "no .rom section") {
		t.Errorf("expected 'no .rom section' in output, got:\n%s", out)
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Errorf("output file %s should not have been created", output)
	}
}

func TestCsrPseudoEndToEnd(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "csr.s")
	output := filepath.Join(dir, "csr.bin")
	src := ".text\nmain:\n csrw 0x305, t0\n csrr t1, 0x342\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-f", "bin", "-o", output)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("assembling csr pseudos failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x73, 0x90, 0x52, 0x30, 0x73, 0x23, 0x20, 0x34}
	if !bytes.Equal(data, want) {
		t.Errorf("csr words = % x, want % x", data, want)
	}
}

// la and call expand to auipc + addi/jalr: %pcrel_lo is relative to the auipc
func TestPCRelPairEndToEnd(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "pcrel.s")
	output := filepath.Join(dir, "pcrel.bin")
	src := ".rom\nla t0, target\ncall target\nnop\ntarget: nop\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-f", "rom", "-o", output)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("assembling la/call failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{
		0x97, 0x02, 0x00, 0x00, // auipc t0, 0
		0x93, 0x82, 0x42, 0x01, // addi  t0, t0, 20  (0x00 + 20 = target)
		0x97, 0x00, 0x00, 0x00, // auipc ra, 0
		0xE7, 0x80, 0xC0, 0x00, // jalr  ra, ra, 12  (0x08 + 12 = target)
		0x13, 0x00, 0x00, 0x00, // nop
		0x13, 0x00, 0x00, 0x00, // target: nop
	}
	if !bytes.Equal(data, want) {
		t.Errorf("la/call words = % x, want % x", data, want)
	}
}

// TestRodataNotWrittenWarns checks that goasm warns when .rodata bytes
// are dropped because the output format does not support them yet.
func TestRodataNotWrittenWarns(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "rodata.asm")
	output := filepath.Join(dir, "rodata.on32")
	src := ".rodata\n.word 1\n.text\nnop\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-o", output)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("assembling failed: %v\n%s", err, out)
	}
	want := "warning: .rodata section (4 bytes) is not written to on32 output yet"
	if !strings.Contains(string(out), want) {
		t.Errorf("expected %q in output, got:\n%s", want, out)
	}

	info, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 16+4 {
		t.Errorf("expected on32 file of 20 bytes (header + nop), got %d", info.Size())
	}
}

// TestRscvFormatEndToEnd checks that -f rscv writes one RSCV chunk with
// every non-empty section, .rodata included, and no warning.
func TestRscvFormatEndToEnd(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "cart.s")
	output := filepath.Join(dir, "out.rscv")
	src := ".text\nmain: nop\n.rodata\n.word 0x11223344\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-f", "rscv", "-o", output)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("assembling rscv failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "warning:") {
		t.Errorf("unexpected warning in output:\n%s", out)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 12+(12+4)+(12+4) {
		t.Fatalf("rscv file size = %d, want 44", len(data))
	}
	if string(data[4:8]) != "RSCV" {
		t.Errorf("chunk type = %q, want RSCV", data[4:8])
	}
	secs := parseRscv(t, data)
	if len(secs) != 2 {
		t.Fatalf("got %d sections, want 2: %+v", len(secs), secs)
	}
	ro := secs[1]
	if ro.kind != rscvKindRodata || ro.addr != 0x01200000 || ro.size != 4 ||
		!bytes.Equal(ro.data, []byte{0x44, 0x33, 0x22, 0x11}) {
		t.Errorf(".rodata section = %+v, want kind 3 at 0x01200000 with 44 33 22 11", ro)
	}
}
