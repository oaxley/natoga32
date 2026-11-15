// Round-trip tests for the NATOGA32 assembler/disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// getProjectRoot returns the path to the goasm project root
func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// filename is .../cmd/goasm/roundtrip_test.go
	// project root is .../
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// buildGoasm builds the goasm binary and returns the path
func buildGoasm(t *testing.T) string {
	t.Helper()
	projectRoot := getProjectRoot()
	goasmPath := filepath.Join(projectRoot, "goasm_test")

	// Build goasm
	cmd := exec.Command("go", "build", "-o", goasmPath, "./cmd/goasm/")
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build goasm: %v\n%s", err, out)
	}

	return goasmPath
}

// stripDisasmFormatting converts disassembly output to reassemblable assembly.
// It removes address prefixes and hex encodings, keeping only the instructions.
func stripDisasmFormatting(disasm string) string {
	var result strings.Builder
	lines := strings.Split(disasm, "\n")

	for _, line := range lines {
		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Check for ".text" or ".data" section markers
		if trimmed == ".text" || trimmed == ".data" || trimmed == ".bss" {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Match lines with address format: "0xXXXXXXXX:  XXXXXXXX  instruction"
		if strings.HasPrefix(trimmed, "0x") && strings.Contains(trimmed, ":") {
			// Find the instruction part after "0x01A00000:  XXXXXXXX  "
			parts := strings.SplitN(trimmed, "  ", 3)
			if len(parts) >= 3 {
				// parts[2] is the instruction
				result.WriteString("    ")
				result.WriteString(parts[2])
				result.WriteString("\n")
				continue
			}
		}

		// Keep other lines as-is (like labels or directives)
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// TestRoundTripAllInstructions tests that assembling, disassembling, and
// reassembling produces identical binary output.
func TestRoundTripAllInstructions(t *testing.T) {
	projectRoot := getProjectRoot()
	testFile := filepath.Join(projectRoot, "testdata", "disasm", "all_instructions.asm")

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	// Create temp directory for test artifacts
	tempDir, err := os.MkdirTemp("", "goasm-roundtrip-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Paths for test artifacts
	bin1 := filepath.Join(tempDir, "pass1.on32")
	disasm := filepath.Join(tempDir, "disasm.asm")
	bin2 := filepath.Join(tempDir, "pass2.on32")

	// Step 1: Assemble original file
	t.Log("Step 1: Assembling original file...")
	cmd := exec.Command(goasm, testFile, "-o", bin1)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Step 1 failed (assemble): %v\n%s", err, out)
	}

	// Step 2: Disassemble to assembly
	t.Log("Step 2: Disassembling to assembly...")
	cmd = exec.Command(goasm, "--disasm", bin1)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Step 2 failed (disassemble): %v", err)
	}

	// Strip formatting to make it reassemblable
	cleanAsm := stripDisasmFormatting(string(output))
	if err := os.WriteFile(disasm, []byte(cleanAsm), 0644); err != nil {
		t.Fatalf("failed to write disassembly: %v", err)
	}

	// Step 3: Reassemble the disassembly
	t.Log("Step 3: Reassembling disassembly...")
	cmd = exec.Command(goasm, disasm, "-o", bin2)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Step 3 failed (reassemble): %v\n%s", err, out)
	}

	// Step 4: Compare binaries
	t.Log("Step 4: Comparing binaries...")
	data1, err := os.ReadFile(bin1)
	if err != nil {
		t.Fatalf("failed to read first binary: %v", err)
	}
	data2, err := os.ReadFile(bin2)
	if err != nil {
		t.Fatalf("failed to read second binary: %v", err)
	}

	if !bytes.Equal(data1, data2) {
		t.Errorf("Round-trip failed: binaries differ\n  Original size: %d bytes\n  Reassembled size: %d bytes",
			len(data1), len(data2))

		// Find first difference
		minLen := len(data1)
		if len(data2) < minLen {
			minLen = len(data2)
		}
		for i := 0; i < minLen; i++ {
			if data1[i] != data2[i] {
				t.Errorf("First difference at offset %d (0x%X): 0x%02X vs 0x%02X",
					i, i, data1[i], data2[i])
				break
			}
		}
	} else {
		t.Logf("Round-trip successful: %d bytes identical", len(data1))
	}
}

// TestRoundTripWithPseudo tests round-trip with --pseudo flag
// Note: With --pseudo, the output may be semantically equivalent but not byte-identical
// because pseudo-instructions may produce different encodings than the original
func TestRoundTripWithPseudo(t *testing.T) {
	projectRoot := getProjectRoot()
	testFile := filepath.Join(projectRoot, "testdata", "disasm", "all_instructions.asm")

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("test file not found: %s", testFile)
	}

	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	// Create temp directory for test artifacts
	tempDir, err := os.MkdirTemp("", "goasm-roundtrip-pseudo-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Paths for test artifacts
	bin1 := filepath.Join(tempDir, "pass1.on32")
	disasm := filepath.Join(tempDir, "disasm.asm")
	bin2 := filepath.Join(tempDir, "pass2.on32")

	// Step 1: Assemble original file
	t.Log("Step 1: Assembling original file...")
	cmd := exec.Command(goasm, testFile, "-o", bin1)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Step 1 failed (assemble): %v\n%s", err, out)
	}

	// Step 2: Disassemble with --pseudo flag
	t.Log("Step 2: Disassembling with --pseudo...")
	cmd = exec.Command(goasm, "--disasm", "--pseudo", bin1)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Step 2 failed (disassemble): %v", err)
	}

	// Strip formatting to make it reassemblable
	cleanAsm := stripDisasmFormatting(string(output))
	if err := os.WriteFile(disasm, []byte(cleanAsm), 0644); err != nil {
		t.Fatalf("failed to write disassembly: %v", err)
	}

	// Step 3: Reassemble the disassembly
	t.Log("Step 3: Reassembling disassembly...")
	cmd = exec.Command(goasm, disasm, "-o", bin2)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Step 3 failed (reassemble): %v\n%s", err, out)
	}

	// Step 4: Compare binaries (should be identical since pseudo-instructions
	// expand to the same machine code)
	t.Log("Step 4: Comparing binaries...")
	data1, err := os.ReadFile(bin1)
	if err != nil {
		t.Fatalf("failed to read first binary: %v", err)
	}
	data2, err := os.ReadFile(bin2)
	if err != nil {
		t.Fatalf("failed to read second binary: %v", err)
	}

	if !bytes.Equal(data1, data2) {
		t.Errorf("Round-trip with --pseudo failed: binaries differ\n  Original size: %d bytes\n  Reassembled size: %d bytes",
			len(data1), len(data2))
	} else {
		t.Logf("Round-trip with --pseudo successful: %d bytes identical", len(data1))
	}
}

// TestDisasmOutputFormat verifies the disassembly output format is correct
func TestDisasmOutputFormat(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	// Create temp directory for test artifacts
	tempDir, err := os.MkdirTemp("", "goasm-format-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple asm file
	asmFile := filepath.Join(tempDir, "test.asm")
	asmContent := `.text
    nop
    addi ra, zero, 10
    add sp, ra, sp
    ret
.data
    .word 0x12345678
`
	if err := os.WriteFile(asmFile, []byte(asmContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	binFile := filepath.Join(tempDir, "test.on32")

	// Assemble
	cmd := exec.Command(goasm, asmFile, "-o", binFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to assemble: %v\n%s", err, out)
	}

	// Disassemble
	cmd = exec.Command(goasm, "--disasm", binFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to disassemble: %v", err)
	}

	outStr := string(output)

	// Verify output contains expected sections
	checks := []string{
		"; Disassembly of",
		"; .text:",
		"; .data:",
		".text",
		".data",
		"0x01A00000:", // First text address
		"0x01200000:", // First data address
	}

	for _, check := range checks {
		if !strings.Contains(outStr, check) {
			t.Errorf("Output missing expected string: %q\nFull output:\n%s", check, outStr)
		}
	}
}

// TestDisasmRawBinary tests disassembly of raw binary files
func TestDisasmRawBinary(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tempDir, err := os.MkdirTemp("", "goasm-rawbin-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple raw binary (nop instruction: 0x00000013)
	binFile := filepath.Join(tempDir, "test.bin")
	binData := []byte{0x13, 0x00, 0x00, 0x00} // Little-endian nop
	if err := os.WriteFile(binFile, binData, 0644); err != nil {
		t.Fatalf("failed to write binary: %v", err)
	}

	// Disassemble with custom base address
	cmd := exec.Command(goasm, "--disasm", "--base", "0x1000", binFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to disassemble: %v", err)
	}

	outStr := string(output)

	// Verify output
	if !strings.Contains(outStr, "(bin format)") {
		t.Errorf("Output should indicate bin format\nFull output:\n%s", outStr)
	}
	if !strings.Contains(outStr, "0x00001000:") {
		t.Errorf("Output should use custom base address 0x1000\nFull output:\n%s", outStr)
	}
	if !strings.Contains(outStr, "addi zero, zero, 0") {
		t.Errorf("Output should contain nop instruction (addi zero, zero, 0)\nFull output:\n%s", outStr)
	}
}

// TestDisasmPseudoOutput tests that --pseudo flag produces readable output
func TestDisasmPseudoOutput(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tempDir, err := os.MkdirTemp("", "goasm-pseudo-out-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create asm file with instructions that map to pseudo-instructions
	asmFile := filepath.Join(tempDir, "test.asm")
	asmContent := `.text
    nop
    li ra, 10
    mv t0, a0
    ret
    j loop
    beqz t0, loop
loop:
    neg t1, a1
`
	if err := os.WriteFile(asmFile, []byte(asmContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	binFile := filepath.Join(tempDir, "test.on32")

	// Assemble
	cmd := exec.Command(goasm, asmFile, "-o", binFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to assemble: %v\n%s", err, out)
	}

	// Disassemble with --pseudo
	cmd = exec.Command(goasm, "--disasm", "--pseudo", binFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to disassemble: %v", err)
	}

	outStr := string(output)

	// Verify pseudo-instructions appear in output
	// Note: "li" and "call" are intentionally NOT converted to pseudos
	// to ensure round-trip compatibility
	pseudoChecks := []string{
		"nop",
		"addi ra, zero, 10", // NOT "li ra, 10" - see pseudo.go comments
		"mv t0, a0",
		"ret",
		"j 0x",
		"beqz t0, 0x",
		"neg t1, a1",
	}

	for _, check := range pseudoChecks {
		if !strings.Contains(outStr, check) {
			t.Errorf("Pseudo output missing: %q\nFull output:\n%s", check, outStr)
		}
	}
}
