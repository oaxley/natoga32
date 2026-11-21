package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// runDebug assembles a small cartridge with -d level and returns the output and
// whether the output file was written
func runDebug(t *testing.T, goasm string, level int, extra ...string) (string, bool, error) {
	t.Helper()
	dir := t.TempDir()
	input := filepath.Join(dir, "in.s")
	output := filepath.Join(dir, "out.rscv")
	if err := os.WriteFile(input, []byte(".text\nmain: nop\nloop: j loop\n.data\nv: .word 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	args := append([]string{input, "-f", "rscv", "-o", output, "-d", strconv.Itoa(level)}, extra...)
	out, err := exec.Command(goasm, args...).CombinedOutput()
	_, statErr := os.Stat(output)
	return string(out), statErr == nil, err
}

// -d 1 to 6 print one phase and stop without a file; -d 7 writes the file and
// prints what was written; 0 prints nothing more
func TestDebugLevels(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	headers := map[int]string{
		1: "=== TOKENS ===",
		2: "=== PREPROCESSED TOKENS ===",
		3: "=== AST ===",
		4: "=== EXPANDED AST ===",
		5: "=== SYMBOLS ===",
		6: "=== ENCODED SECTIONS ===",
		7: "=== OUTPUT ===",
	}
	for level := 0; level <= 7; level++ {
		t.Run("-d "+strconv.Itoa(level), func(t *testing.T) {
			out, written, err := runDebug(t, goasm, level)
			if err != nil {
				t.Fatalf("goasm failed: %v\n%s", err, out)
			}
			for l, h := range headers {
				if got := strings.Contains(out, h); got != (l == level) {
					t.Errorf("%q printed = %v, want %v:\n%s", h, got, l == level, out)
				}
			}
			if wantFile := level == 0 || level == 7; written != wantFile {
				t.Errorf("output file written = %v, want %v:\n%s", written, wantFile, out)
			}
			if strings.Contains(out, "Phase 8") || strings.Contains(out, "pass 3") {
				t.Errorf("phases are 1 to 7, with no pass 3:\n%s", out)
			}
		})
	}
}

// the phase lines are numbered 1 to 7, the debug level of each one
func TestDebugPhaseNumbers(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	out, _, err := runDebug(t, goasm, 0)
	if err != nil {
		t.Fatalf("goasm failed: %v\n%s", err, out)
	}
	for _, want := range []string{"Phase 5: Semantic Analyzer - pass 1", "Phase 6: Semantic Analyzer - pass 2", "Phase 7: Output"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}

// -d 7 lists the format, the sections with data, the file size and the dSYM labels
func TestDebugOutputSummary(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	out, written, err := runDebug(t, goasm, 7, "--sym", filepath.Join(t.TempDir(), "out.sym"))
	if err != nil || !written {
		t.Fatalf("goasm failed (written=%v): %v\n%s", written, err, out)
	}
	for _, want := range []string{
		"format: rscv",
		".text: base=0x01A00000 size=8",
		".data: base=0x01400000 size=4",
		"out.rscv:",
		"bytes",
		"out.sym: 3 labels",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}

// a level outside 0..7 stops before anything is written
func TestDebugLevelRange(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	for _, level := range []int{8, 9, -1} {
		t.Run("-d "+strconv.Itoa(level), func(t *testing.T) {
			out, written, err := runDebug(t, goasm, level)
			if err == nil {
				t.Fatalf("expected an error, assembled fine:\n%s", out)
			}
			if want := "--debug must be 0 to 7, got " + strconv.Itoa(level); !strings.Contains(out, want) {
				t.Errorf("expected %q in output, got:\n%s", want, out)
			}
			if written {
				t.Errorf("no file should be written")
			}
		})
	}
}
