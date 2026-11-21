// Tests for the cartridge chunk output of the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package assembler

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"strings"
	"testing"

	"github.com/natoga32/n32asm/internal/symbol"
)

// rscvSection is one decoded section of an RSCV chunk
type rscvSection struct {
	kind byte
	addr uint32
	size uint32
	data []byte
}

// parseRscv checks the chunk framing and decodes the RSCV sections
func parseRscv(t *testing.T, c []byte) []rscvSection {
	t.Helper()
	if len(c) < 12 {
		t.Fatalf("chunk too short: %d bytes", len(c))
	}
	length := binary.LittleEndian.Uint32(c[0:4])
	if string(c[4:8]) != "RSCV" {
		t.Fatalf("chunk type = %q, want RSCV", c[4:8])
	}
	padded := (int(length) + 3) &^ 3
	if len(c) != 12+padded {
		t.Fatalf("chunk size = %d, want %d", len(c), 12+padded)
	}
	data := c[8 : 8+length]
	wantCrc := crc32.ChecksumIEEE(c[4 : 8+length])
	if got := binary.LittleEndian.Uint32(c[8+padded:]); got != wantCrc {
		t.Fatalf("chunk CRC = %#08x, want %#08x", got, wantCrc)
	}

	var secs []rscvSection
	for off := 0; off < len(data); {
		if len(data)-off < 12 {
			t.Fatalf("truncated section header at offset %d", off)
		}
		h := data[off : off+12]
		if h[1] != 0 || h[2] != 0 || h[3] != 0 {
			t.Errorf("flags/reserved = % x, want zeros", h[1:4])
		}
		s := rscvSection{
			kind: h[0],
			addr: binary.LittleEndian.Uint32(h[4:8]),
			size: binary.LittleEndian.Uint32(h[8:12]),
		}
		off += 12
		if s.kind != rscvKindBss {
			n := (int(s.size) + 3) &^ 3
			s.data = data[off : off+n]
			off += n
		}
		secs = append(secs, s)
	}
	return secs
}

func TestChunkLayout(t *testing.T) {
	c := chunk("RSCV", []byte{1, 2, 3, 4, 5})
	if len(c) != 20 {
		t.Fatalf("chunk size = %d, want 20", len(c))
	}
	if got := binary.LittleEndian.Uint32(c[0:4]); got != 5 {
		t.Errorf("length = %d, want 5", got)
	}
	if string(c[4:8]) != "RSCV" {
		t.Errorf("type = %q, want RSCV", c[4:8])
	}
	if want := []byte{1, 2, 3, 4, 5, 0, 0, 0}; !bytes.Equal(c[8:16], want) {
		t.Errorf("data = % x, want % x", c[8:16], want)
	}
	want := crc32.ChecksumIEEE([]byte("RSCV\x01\x02\x03\x04\x05"))
	if got := binary.LittleEndian.Uint32(c[16:20]); got != want {
		t.Errorf("CRC = %#08x, want %#08x", got, want)
	}
}

func TestChunkEmpty(t *testing.T) {
	c := chunk("IEND", nil)
	if len(c) != 12 {
		t.Fatalf("chunk size = %d, want 12", len(c))
	}
	want := crc32.ChecksumIEEE([]byte("IEND"))
	if want != 0xAE426082 {
		t.Fatalf("IEND CRC = %#08x, want 0xAE426082", want)
	}
	if got := binary.LittleEndian.Uint32(c[8:12]); got != want {
		t.Errorf("CRC = %#08x, want %#08x", got, want)
	}
}

func TestBuildRscvSections(t *testing.T) {
	sections := newMappedSections()
	sections.Get(".text").Emit([]byte{0x13, 0, 0, 0})
	sections.Get(".data").Emit([]byte{1, 2, 3, 4, 5})
	sections.Get(".bss").Reserve(16)

	c, err := buildRscv(sections)
	if err != nil {
		t.Fatal(err)
	}
	if got := binary.LittleEndian.Uint32(c[0:4]); got != 48 {
		t.Errorf("chunk length = %d, want 48", got)
	}
	secs := parseRscv(t, c)
	want := []rscvSection{
		{rscvKindText, 0x01A00000, 4, []byte{0x13, 0, 0, 0}},
		{rscvKindData, 0x01400000, 5, []byte{1, 2, 3, 4, 5, 0, 0, 0}},
		{rscvKindBss, 0x01000000, 16, nil},
	}
	if len(secs) != len(want) {
		t.Fatalf("got %d sections, want %d: %+v", len(secs), len(want), secs)
	}
	for i, w := range want {
		s := secs[i]
		if s.kind != w.kind || s.addr != w.addr || s.size != w.size || !bytes.Equal(s.data, w.data) {
			t.Errorf("section %d = %+v, want %+v", i, s, w)
		}
	}
}

func TestBuildRscvNeedsText(t *testing.T) {
	sections := newMappedSections()
	sections.Get(".data").Emit([]byte{1, 2, 3, 4})
	_, err := buildRscv(sections)
	if err == nil || !strings.Contains(err.Error(), "no .text section") {
		t.Errorf("expected 'no .text section' error, got: %v", err)
	}
}

func TestBuildRscvRejectsRom(t *testing.T) {
	sections := newMappedSections()
	sections.Get(".text").Emit([]byte{0x13, 0, 0, 0})
	sections.Get(".rom").Emit([]byte{1, 2, 3, 4})
	_, err := buildRscv(sections)
	if err == nil || !strings.Contains(err.Error(), ".rom section can't go in a cartridge") {
		t.Errorf("expected '.rom section can't go in a cartridge' error, got: %v", err)
	}
}

func TestBuildRscvRejectsBssData(t *testing.T) {
	sections := newMappedSections()
	sections.Get(".text").Emit([]byte{0x13, 0, 0, 0})
	sections.Get(".bss").Emit([]byte{0, 0, 0, 0, 0, 0, 7, 0})
	_, err := buildRscv(sections)
	if err == nil || err.Error() != ".bss section holds non-zero data at 0x01000006: use .data" {
		t.Errorf("expected '.bss section holds non-zero data at 0x01000006: use .data' error, got: %v", err)
	}
}

func TestBuildRscvAcceptsZeroBssData(t *testing.T) {
	sections := newMappedSections()
	sections.Get(".text").Emit([]byte{0x13, 0, 0, 0})
	sections.Get(".bss").Emit([]byte{0, 0, 0, 0})
	if _, err := buildRscv(sections); err != nil {
		t.Errorf("zeros in .bss: %v", err)
	}
}

// dsymEntry is one decoded entry of a dSYM chunk
type dsymEntry struct {
	addr  uint32
	kind  byte
	flags byte
	name  string
	pad   []byte
}

// parseDsym checks the chunk framing and decodes the dSYM entries
func parseDsym(t *testing.T, c []byte) []dsymEntry {
	t.Helper()
	if len(c) < 16 {
		t.Fatalf("chunk too short: %d bytes", len(c))
	}
	length := binary.LittleEndian.Uint32(c[0:4])
	if string(c[4:8]) != "dSYM" {
		t.Fatalf("chunk type = %q, want dSYM", c[4:8])
	}
	padded := (int(length) + 3) &^ 3
	if len(c) != 12+padded {
		t.Fatalf("chunk size = %d, want %d", len(c), 12+padded)
	}
	wantCrc := crc32.ChecksumIEEE(c[4 : 8+length])
	if got := binary.LittleEndian.Uint32(c[8+padded:]); got != wantCrc {
		t.Fatalf("chunk CRC = %#08x, want %#08x", got, wantCrc)
	}

	data := c[8 : 8+length]
	count := binary.LittleEndian.Uint32(data[0:4])
	var entries []dsymEntry
	off := 4
	for i := uint32(0); i < count; i++ {
		if len(data)-off < 8 {
			t.Fatalf("truncated entry %d at offset %d", i, off)
		}
		e := dsymEntry{
			addr:  binary.LittleEndian.Uint32(data[off : off+4]),
			kind:  data[off+4],
			flags: data[off+5],
		}
		n := int(binary.LittleEndian.Uint16(data[off+6 : off+8]))
		off += 8
		e.name = string(data[off : off+n])
		off += n
		e.pad = data[off : off+align4(n)-n]
		off += align4(n) - n
		entries = append(entries, e)
	}
	if off != len(data) {
		t.Errorf("dSYM data has %d trailing bytes", len(data)-off)
	}
	return entries
}

// dsymTestTable is the symbol table of TestBuildDsym
func dsymTestTable(t *testing.T) *symbol.Table {
	t.Helper()
	symbols := symbol.New()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(symbols.DefineLabel("zeta", 0x01A00008, ".text", 0))
	must(symbols.SetGlobal("zeta"))
	must(symbols.DefineLabel("alpha", 0x01A00008, ".text", 0))
	must(symbols.DefineLabel("buf", 0x01000000, ".bss", 0))
	must(symbols.DefineConstant("ANSWER", 42, 0))
	must(symbols.DefineLabel("boot", 0, ".rom", 0))
	return symbols
}

func TestBuildDsymNameTooLong(t *testing.T) {
	symbols := symbol.New()
	long := strings.Repeat("a", 65536)
	if err := symbols.DefineLabel(long, 0x01A00000, ".text", 0); err != nil {
		t.Fatal(err)
	}
	_, err := buildDsym(symbols)
	if err == nil || !strings.Contains(err.Error(), "is 65536 bytes: dSYM names are at most 65535 bytes") {
		t.Fatalf("err = %v, want a name length error", err)
	}

	symbols = symbol.New()
	if err := symbols.DefineLabel(long[:65535], 0x01A00000, ".text", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := buildDsym(symbols); err != nil {
		t.Fatalf("65535-byte name: %v", err)
	}
}

func mustBuildDsym(t *testing.T, symbols *symbol.Table) []byte {
	t.Helper()
	c, err := buildDsym(symbols)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestBuildDsym(t *testing.T) {
	c := mustBuildDsym(t, dsymTestTable(t))
	if string(c[4:8]) != "dSYM" {
		t.Fatalf("chunk type = %q, want dSYM", c[4:8])
	}
	if got := binary.LittleEndian.Uint32(c[8:12]); got != 3 {
		t.Fatalf("entry count = %d, want 3", got)
	}
	entries := parseDsym(t, c)
	want := []dsymEntry{
		{0x01000000, rscvKindBss, 0, "buf", []byte{0}},
		{0x01A00008, rscvKindText, 0, "alpha", []byte{0, 0, 0}},
		{0x01A00008, rscvKindText, 1, "zeta", []byte{}},
	}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for i, w := range want {
		e := entries[i]
		if e.addr != w.addr || e.kind != w.kind || e.flags != w.flags || e.name != w.name || !bytes.Equal(e.pad, w.pad) {
			t.Errorf("entry %d = %+v, want %+v", i, e, w)
		}
	}
	// alpha: 4 count + 12 buf entry, then its header; name length 5
	if got := binary.LittleEndian.Uint16(c[8+4+12+6:]); got != 5 {
		t.Errorf("alpha name length = %d, want 5", got)
	}
}

func TestBuildDsymDeterministic(t *testing.T) {
	symbols := dsymTestTable(t)
	first := mustBuildDsym(t, symbols)
	for i := 0; i < 20; i++ {
		if got := mustBuildDsym(t, symbols); !bytes.Equal(got, first) {
			t.Fatalf("call %d gave different bytes:\n% x\nwant\n% x", i, got, first)
		}
	}
}
