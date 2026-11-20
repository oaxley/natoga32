package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natoga32/goasm/internal/section"
)

func newMappedSections() *section.Manager {
	sections := section.NewManager()
	sections.SetBaseAddresses(
		MMAP_TEXT_SEGMENT,
		MMAP_DATA_SEGMENT,
		MMAP_BSS_SEGMENT,
		MMAP_ROM_SEGMENT,
		MMAP_RODATA_SEGMENT,
	)
	return sections
}

func TestSectionSizeLimits(t *testing.T) {
	cases := []struct {
		name    string
		limit   int64
		message string
	}{
		{".bss", 2 * 1024 * 1024, "section .bss is 2097153 bytes, exceeds its 2MiB window (0x01000000-0x011FFFFF)"},
		{".rodata", 2 * 1024 * 1024, "section .rodata is 2097153 bytes, exceeds its 2MiB window (0x01200000-0x013FFFFF)"},
		{".data", 6 * 1024 * 1024, "section .data is 6291457 bytes, exceeds its 6MiB window (0x01400000-0x019FFFFF)"},
		{".text", 6 * 1024 * 1024, "section .text is 6291457 bytes, exceeds its 6MiB window (0x01A00000-0x01FFFFFF)"},
		{".rom", 2 * 1024 * 1024, "section .rom is 2097153 bytes, exceeds its 2MiB window (0x00000000-0x001FFFFF)"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sections := newMappedSections()

			sections.Get(c.name).Size = c.limit
			if err := checkSectionSizes(sections); err != nil {
				t.Fatalf("section exactly at its limit must be accepted, got: %v", err)
			}

			sections.Get(c.name).Size = c.limit + 1
			err := checkSectionSizes(sections)
			if err == nil {
				t.Fatal("expected an error for a section one byte over its limit, got nil")
			}
			if err.Error() != c.message {
				t.Errorf("unexpected error message:\n got: %s\nwant: %s", err, c.message)
			}
		})
	}
}

func TestEmptySectionsPassSizeCheck(t *testing.T) {
	if err := checkSectionSizes(newMappedSections()); err != nil {
		t.Errorf("empty sections must pass, got: %v", err)
	}
}

// TestRodataOverflowEndToEnd checks that goasm refuses a .rodata
// section that would spill into .data.
func TestRodataOverflowEndToEnd(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "big.asm")
	output := filepath.Join(dir, "big.rscv")
	src := ".rodata\n.space 2097152\n.byte 1\n.text\nnop\n"
	if err := os.WriteFile(input, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goasm, input, "-o", output)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected goasm to fail on .rodata overflow, it succeeded:\n%s", out)
	}
	want := "section .rodata is 2097153 bytes, exceeds its 2MiB window (0x01200000-0x013FFFFF)"
	if !strings.Contains(string(out), want) {
		t.Errorf("expected %q in output, got:\n%s", want, out)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Errorf("no output file must be written on overflow")
	}
}
