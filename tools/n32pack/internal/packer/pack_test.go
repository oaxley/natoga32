// Tests of the cartridge packing
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package packer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// packDir writes the files of a test cartridge in a new temporary directory and returns a manifest using them
func packDir(t *testing.T, files map[string][]byte) *Manifest {
	t.Helper()
	dir := t.TempDir()
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return &Manifest{
		Name:    "Pack Test",
		Version: "1.00",
		MinRom:  "0.1",
		Regions: []string{"europe"},
		Date:    "2026-03-01T12:00:00Z",
		Rscv:    "game.rscv",
		dir:     dir,
		file:    filepath.Join(dir, "game.json"),
	}
}

// testRscv returns an RSCV chunk file with a 6-byte .text section and a .bss section
func testRscv() []byte {
	return appendChunk(nil, "RSCV", concat(
		section(kindText, 0x01A00000, 6, []byte{1, 2, 3, 4, 5, 6}),
		section(kindBss, 0x01000000, 32, nil),
	))
}

// testDsym returns a dSYM chunk file with the label main at 0x01A00004
func testDsym() []byte {
	return appendChunk(nil, "dSYM", dsymData(1, dsymEntry{0x01A00004, kindText, 1, "main"}))
}

func layoutManifest(t *testing.T) *Manifest {
	m := packDir(t, map[string][]byte{
		"game.rscv": testRscv(),
		"game.dsym": testDsym(),
		"tiles.bin": {0xAA, 0xBB, 0xCC},
		"pal.bin":   {0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
	})
	m.Dsym = "game.dsym"
	m.Entry = "main"
	m.Assets = []Asset{
		{Name: "TILES", Type: "TILS", Slot: 0x0102, Format: 3, Count: 0x0405, File: "tiles.bin", Boot: true},
		{Name: "PAL", Type: "TPAL", Slot: 7, Format: 1, Count: 4, File: "pal.bin"},
	}
	return m
}

func TestPackLayout(t *testing.T) {
	m := layoutManifest(t)
	file, err := pack(m, time.Now())
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	c := readCart(t, file)

	var types []string
	for _, ch := range c.chunks {
		types = append(types, ch.typ)
	}
	if got := strings.Join(types, " "); got != "IHDR INDX RSCV TILS TPAL dSYM IEND" {
		t.Fatalf("chunk order %s", got)
	}
	if c.flags != 0 {
		t.Errorf("header flags 0x%04X", c.flags)
	}
	if len(c.ihdr) != 64 {
		t.Errorf("IHDR is %d bytes", len(c.ihdr))
	}
	if !bytes.Equal(c.ihdr[0x0C:0x10], []byte{0x04, 0x00, 0xA0, 0x01}) {
		t.Errorf("IHDR entry % X", c.ihdr[0x0C:0x10])
	}

	wantIndx := []struct {
		typ   string
		id    uint16
		flags uint16
	}{{"RSCV", 0, 1}, {"TILS", 1, 1}, {"TPAL", 2, 0}, {"dSYM", 3, 0}}
	if len(c.indx) != len(wantIndx) {
		t.Fatalf("INDX has %d entries", len(c.indx))
	}
	for i, w := range wantIndx {
		e := c.indx[i]
		ch := c.chunks[i+2]
		if e.typ != w.typ || e.id != w.id || e.flags != w.flags {
			t.Errorf("INDX %d = %s id %d flags %d, want %s id %d flags %d", i, e.typ, e.id, e.flags, w.typ, w.id, w.flags)
		}
		if e.typ != ch.typ || e.offset != ch.offset || e.length != uint32(len(ch.data)) {
			t.Errorf("INDX %d = %s at %d (%d bytes), chunk %s at %d (%d bytes)",
				i, e.typ, e.offset, e.length, ch.typ, ch.offset, len(ch.data))
		}
	}

	rscv, _ := readChunkFile(m.path("game.rscv"), "RSCV")
	if !bytes.Equal(c.chunks[2].data, rscv) {
		t.Errorf("RSCV data differs")
	}
	if want := []byte{0x02, 0x01, 3, 0, 0x05, 0x04, 0, 0, 0xAA, 0xBB, 0xCC}; !bytes.Equal(c.chunks[3].data, want) {
		t.Errorf("TILS data % X, want % X", c.chunks[3].data, want)
	}
	if want := []byte{7, 0, 1, 0, 4, 0, 0, 0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}; !bytes.Equal(c.chunks[4].data, want) {
		t.Errorf("TPAL data % X, want % X", c.chunks[4].data, want)
	}
	dsym, _ := readChunkFile(m.path("game.dsym"), "dSYM")
	if !bytes.Equal(c.chunks[5].data, dsym) {
		t.Errorf("dSYM data differs")
	}
	last := c.chunks[6]
	if len(last.data) != 0 || uint64(last.offset)+4 != uint64(len(file)) {
		t.Errorf("IEND at %d (%d bytes), file is %d bytes", last.offset, len(last.data), len(file))
	}
}

func TestPackNoDsym(t *testing.T) {
	m := packDir(t, map[string][]byte{"game.rscv": testRscv(), "empty.bin": {}})
	mode := 2
	m.GraphicsMode = &mode
	m.Assets = []Asset{{Name: "EMPTY", Type: "SPRS", File: "empty.bin"}}
	file, err := pack(m, time.Now())
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	c := readCart(t, file)
	if c.flags != 0x000A {
		t.Errorf("header flags 0x%04X", c.flags)
	}
	if len(c.indx) != 2 || c.indx[0].typ != "RSCV" || c.indx[1].typ != "SPRS" {
		t.Fatalf("INDX %v", c.indx)
	}
	if !bytes.Equal(c.chunks[3].data, make([]byte, 8)) {
		t.Errorf("SPRS data % X", c.chunks[3].data)
	}
	if !bytes.Equal(c.ihdr[0x0C:0x10], []byte{0x00, 0x00, 0xA0, 0x01}) {
		t.Errorf("IHDR entry % X", c.ihdr[0x0C:0x10])
	}
}

func TestPackDeterministic(t *testing.T) {
	m := layoutManifest(t)
	a, err := pack(m, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	b, err := pack(m, time.Date(2027, 5, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Errorf("two packs differ")
	}
}

func TestPackErrors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		edit  func(m *Manifest)
		files map[string][]byte
		want  string
	}{
		{"missing asset", func(m *Manifest) {
			m.Assets = []Asset{{Name: "GONE", Type: "TILS", File: "gone.bin"}}
		}, nil, `asset "GONE": `},
		{"invalid RSCV", nil, map[string][]byte{"game.rscv": appendChunk(nil, "RSCV", []byte{1, 2, 3})}, "RSCV: "},
		{"dsym holds RSCV", func(m *Manifest) { m.Dsym = "bad.dsym" }, map[string][]byte{"bad.dsym": testRscv()}, "not a dSYM chunk"},
		{"missing rscv", func(m *Manifest) { m.Rscv = "none.rscv" }, nil, "none.rscv"},
		{"bad entry", func(m *Manifest) { m.Entry = "0x01A00002" }, nil, "game.json: entry 0x01A00002 is not 4-byte aligned"},
		{"bad date", func(m *Manifest) { m.Date = "2025-01-01T00:00:00Z" }, nil, "game.json: date is before 2026-01-01"},
	} {
		files := map[string][]byte{"game.rscv": testRscv()}
		for k, v := range tc.files {
			files[k] = v
		}
		m := packDir(t, files)
		if tc.edit != nil {
			tc.edit(m)
		}
		_, err := pack(m, time.Now())
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", tc.name, err, tc.want)
		}
	}
}

func TestCheckLimits(t *testing.T) {
	if err := checkLimits(15872, 256<<20); err != nil {
		t.Errorf("at the limits: %v", err)
	}
	if err := checkLimits(15873, 100); err == nil || err.Error() != "too many chunks: 15873 (max 15872)" {
		t.Errorf("too many chunks: %v", err)
	}
	if err := checkLimits(1, 256<<20+4); err == nil || err.Error() != "cartridge is 268435460 bytes, over the 256 MiB limit" {
		t.Errorf("too big: %v", err)
	}
}

func TestPackTooManyChunks(t *testing.T) {
	m := packDir(t, map[string][]byte{"game.rscv": testRscv(), "a.bin": {1}})
	m.Assets = make([]Asset, 15872)
	for i := range m.Assets {
		m.Assets[i] = Asset{Name: fmt.Sprintf("A%d", i), Type: "TILS", File: "a.bin"}
	}
	_, err := pack(m, time.Now())
	if want := m.file + ": too many chunks: 15873 (max 15872)"; err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
	m.Assets = m.Assets[:15871]
	m.Dsym = ""
	if _, err := pack(m, time.Now()); err != nil {
		t.Errorf("15872 chunks: %v", err)
	}
}
