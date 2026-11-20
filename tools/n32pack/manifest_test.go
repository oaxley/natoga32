// Tests of the cartridge manifest loading
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeManifest writes a manifest in a new temporary directory and returns its path
func writeManifest(t *testing.T, json string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "game.json")
	if err := os.WriteFile(path, []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadManifestFull(t *testing.T) {
	path := writeManifest(t, `{
		"name": "Test Game",
		"version": "1.00",
		"min_rom": "0.1",
		"regions": ["americas", "europe", "asia"],
		"languages": ["english"],
		"date": "2026-01-02T00:00:00Z",
		"graphics_mode": 2,
		"rscv": "game.rscv",
		"dsym": "game.dsym",
		"entry": "main",
		"assets": [
			{"name": "TILES_1", "type": "TILS", "slot": 3, "format": 1, "count": 16, "file": "tiles.bin", "boot": true},
			{"name": "PAL", "type": "TPAL", "file": "pal.bin"}
		]
	}`)
	m, err := loadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "Test Game" || m.Rscv != "game.rscv" || m.Dsym != "game.dsym" || m.Entry != "main" {
		t.Errorf("manifest = %+v", m)
	}
	if m.GraphicsMode == nil || *m.GraphicsMode != 2 {
		t.Errorf("graphics_mode = %v, want 2", m.GraphicsMode)
	}
	if got := m.versionBCD(); got != 0x0100 {
		t.Errorf("versionBCD() = %#04x, want 0x0100", got)
	}
	if major, minor := m.minRom(); major != 0 || minor != 1 {
		t.Errorf("minRom() = (%d, %d), want (0, 1)", major, minor)
	}
	if got := m.regionBits(); got != 0x07 {
		t.Errorf("regionBits() = %#02x, want 0x07", got)
	}
	if got := m.languageBits(); got != 1 {
		t.Errorf("languageBits() = %#04x, want 1", got)
	}
	date, err := m.dateSeconds(time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if date != 86400 {
		t.Errorf("dateSeconds() = %d, want 86400", date)
	}
	if len(m.Assets) != 2 {
		t.Fatalf("assets = %+v", m.Assets)
	}
	want := Asset{Name: "TILES_1", Type: "TILS", Slot: 3, Format: 1, Count: 16, File: "tiles.bin", Boot: true}
	if m.Assets[0] != want {
		t.Errorf("asset 0 = %+v, want %+v", m.Assets[0], want)
	}
}

func TestLoadManifestDefaults(t *testing.T) {
	m, err := loadManifest(writeManifest(t, `{"name": "G", "regions": ["europe"], "rscv": "g.rscv"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := m.versionBCD(); got != 0x0100 {
		t.Errorf("versionBCD() = %#04x, want 0x0100", got)
	}
	if major, minor := m.minRom(); major != 0 || minor != 1 {
		t.Errorf("minRom() = (%d, %d), want (0, 1)", major, minor)
	}
	if got := m.regionBits(); got != 0x02 {
		t.Errorf("regionBits() = %#02x, want 0x02", got)
	}
	if got := m.languageBits(); got != 0 {
		t.Errorf("languageBits() = %#04x, want 0", got)
	}
	if m.GraphicsMode != nil {
		t.Errorf("graphics_mode = %d, want absent", *m.GraphicsMode)
	}
	now := time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)
	date, err := m.dateSeconds(now)
	if err != nil {
		t.Fatal(err)
	}
	if date != 3600 {
		t.Errorf("dateSeconds(now) = %d, want 3600", date)
	}
	if _, err := m.dateSeconds(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Error("dateSeconds(2025) gave no error")
	}
}

func TestManifestValues(t *testing.T) {
	m, err := loadManifest(writeManifest(t, `{"name": "G", "version": "2.15", "min_rom": "3.255",
		"regions": ["asia"], "date": "2026-01-01T01:00:00+01:00", "rscv": "g.rscv"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := m.versionBCD(); got != 0x0215 {
		t.Errorf("versionBCD() = %#04x, want 0x0215", got)
	}
	if major, minor := m.minRom(); major != 3 || minor != 255 {
		t.Errorf("minRom() = (%d, %d), want (3, 255)", major, minor)
	}
	if got := m.regionBits(); got != 0x04 {
		t.Errorf("regionBits() = %#02x, want 0x04", got)
	}
	if date, err := m.dateSeconds(time.Now()); err != nil || date != 0 {
		t.Errorf("dateSeconds() = %d, %v, want 0", date, err)
	}
}

func TestLoadManifestErrors(t *testing.T) {
	const base = `"regions": ["europe"], "rscv": "g.rscv"`
	tests := []struct {
		name, json, want string
	}{
		{"unknown field", `{"name": "G", ` + base + `, "colour": 1}`, "unknown field"},
		{"missing name", `{` + base + `}`, "name must be 1 to 32 bytes of UTF-8"},
		{"name too long", `{"name": "` + strings.Repeat("A", 33) + `", ` + base + `}`, "name must be 1 to 32 bytes of UTF-8"},
		{"name bad UTF-8", "{\"name\": \"A\xff\", " + base + `}`, "name must be 1 to 32 bytes of UTF-8"},
		{"bad version", `{"name": "G", "version": "1.0", ` + base + `}`, "version must look like 1.00"},
		{"bad version letters", `{"name": "G", "version": "1.0a", ` + base + `}`, "version must look like 1.00"},
		{"bad min_rom", `{"name": "G", "min_rom": "0.256", ` + base + `}`, "min_rom must look like 0.1"},
		{"no regions", `{"name": "G", "rscv": "g.rscv"}`, "regions: at least one region"},
		{"unknown region", `{"name": "G", "regions": ["mars"], "rscv": "g.rscv"}`, `regions: unknown region "mars"`},
		{"unknown language", `{"name": "G", "languages": ["klingon"], ` + base + `}`, `unknown language "klingon"`},
		{"bad date", `{"name": "G", "date": "yesterday", ` + base + `}`, "date must be RFC 3339"},
		{"date before 2026", `{"name": "G", "date": "2025-12-31T23:59:59Z", ` + base + `}`, "date is before 2026-01-01"},
		{"graphics_mode out of range", `{"name": "G", "graphics_mode": 4, ` + base + `}`, "graphics_mode must be 0 to 3"},
		{"graphics_mode negative", `{"name": "G", "graphics_mode": -1, ` + base + `}`, "graphics_mode must be 0 to 3"},
		{"missing rscv", `{"name": "G", "regions": ["europe"]}`, "rscv is required"},
		{"bad asset name", `{"name": "G", ` + base + `, "assets": [{"name": "x", "type": "TILS", "file": "a"}]}`, `asset name "x" must match [A-Z][A-Z0-9_]*`},
		{"duplicate asset name", `{"name": "G", ` + base + `, "assets": [{"name": "X", "type": "TILS", "file": "a"}, {"name": "X", "type": "TMAP", "file": "b"}]}`, `duplicate asset name "X"`},
		{"bad asset type", `{"name": "G", ` + base + `, "assets": [{"name": "X", "type": "ABCD", "file": "a"}]}`, `asset "X": unknown type "ABCD"`},
		{"missing asset file", `{"name": "G", ` + base + `, "assets": [{"name": "X", "type": "TILS"}]}`, `asset "X": file is required`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := writeManifest(t, tc.json)
			_, err := loadManifest(path)
			if err == nil {
				t.Fatal("no error")
			}
			if !strings.HasPrefix(err.Error(), path+": ") || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want %q prefixed with the path", err, tc.want)
			}
		})
	}
}

func TestManifestPathsRelative(t *testing.T) {
	path := writeManifest(t, `{"name": "G", "regions": ["europe"], "rscv": "g.rscv"}`)
	m, err := loadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	if got := m.path("a.rscv"); got != filepath.Join(dir, "a.rscv") {
		t.Errorf("path(a.rscv) = %q, want %q", got, filepath.Join(dir, "a.rscv"))
	}
	abs := filepath.Join(t.TempDir(), "b.rscv")
	if got := m.path(abs); got != abs {
		t.Errorf("path(%q) = %q, want unchanged", abs, got)
	}
}
