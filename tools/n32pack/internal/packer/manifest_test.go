// Tests of the cartridge manifest loading
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package packer

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
	if m.file != path || m.dir != filepath.Dir(path) {
		t.Errorf("file, dir = %q, %q; want %q, %q", m.file, m.dir, path, filepath.Dir(path))
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

// A value of the wrong JSON type, or a number out of range, names the field
// in plain words: no "json:" or Go struct text
func TestLoadManifestTypeErrors(t *testing.T) {
	const base = `"name": "G", "regions": ["europe"], "rscv": "g.rscv"`
	asset := func(fields string) string {
		return `{` + base + `, "assets": [{"name": "TILES", "type": "TILS", "file": "a", ` + fields + `}]}`
	}
	tests := []struct {
		name, json, want string
	}{
		{"slot too big", asset(`"slot": 70000`), `asset "TILES": slot must be an integer from 0 to 65535`},
		{"slot negative", asset(`"slot": -1`), `asset "TILES": slot must be an integer from 0 to 65535`},
		{"slot string", asset(`"slot": "3"`), `asset "TILES": slot must be an integer from 0 to 65535`},
		{"format too big", asset(`"format": 256`), `asset "TILES": format must be an integer from 0 to 255`},
		{"count fraction", asset(`"count": 1.5`), `asset "TILES": count must be an integer from 0 to 65535`},
		{"count too big", asset(`"count": 70000`), `asset "TILES": count must be an integer from 0 to 65535`},
		{"boot number", asset(`"boot": 1`), `asset "TILES": boot must be true or false`},
		{"asset file number", asset(`"slot": 0, "file": 5`), `asset "TILES": file must be a string`},
		{"asset name number", `{` + base + `, "assets": [{"name": 5, "type": "TILS", "file": "a"}]}`, `assets[0]: name must be a string`},
		{"second asset unnamed", `{` + base + `, "assets": [{"name": "A", "type": "TILS", "file": "a"}, {"type": "TILS", "file": "b", "slot": -2}]}`, `assets[1]: slot must be an integer from 0 to 65535`},
		{"asset not an object", `{` + base + `, "assets": [1]}`, `assets[0] must be an object`},
		{"assets not a list", `{` + base + `, "assets": "x"}`, `assets must be a list of objects`},
		{"version number", `{"version": 1, ` + base + `}`, `version must be a string`},
		{"date number", `{"date": 2026, ` + base + `}`, `date must be a string`},
		{"regions string", `{"name": "G", "regions": "europe", "rscv": "g.rscv"}`, `regions must be a list of strings`},
		{"languages number item", `{"languages": ["english", 1], ` + base + `}`, `languages must be a list of strings`},
		{"graphics_mode fraction", `{"graphics_mode": 9.5, ` + base + `}`, `graphics_mode must be 0 to 3`},
		{"graphics_mode string", `{"graphics_mode": "2", ` + base + `}`, `graphics_mode must be 0 to 3`},
		{"rscv number", `{"name": "G", "regions": ["europe"], "rscv": 1}`, `rscv must be a string`},
		{"manifest not an object", `[1]`, `the manifest must be a JSON object`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := writeManifest(t, tc.json)
			_, err := loadManifest(path)
			if err == nil {
				t.Fatal("no error")
			}
			if err.Error() != path+": "+tc.want {
				t.Errorf("error = %q, want %q", err, path+": "+tc.want)
			}
		})
	}
}

// date takes Z or a +-hh:mm offset with hh 00-23 and mm 00-59 (RFC 3339)
func TestManifestDateOffset(t *testing.T) {
	const base = `"name": "G", "regions": ["europe"], "rscv": "g.rscv"`
	good := map[string]uint32{
		"2026-01-01T01:00:00Z":      3600,
		"2026-01-01T02:00:00+01:00": 3600,
		"2026-01-01T00:00:00-01:30": 5400,
		"2026-01-02T00:00:00+23:59": 86400 - 23*3600 - 59*60,
		"2026-01-01T01:00:00-00:00": 3600,
	}
	for date, want := range good {
		path := writeManifest(t, `{"date": "`+date+`", `+base+`}`)
		m, err := loadManifest(path)
		if err != nil {
			t.Errorf("%s: %v", date, err)
			continue
		}
		got, err := m.dateSeconds(time.Now())
		if err != nil || got != want {
			t.Errorf("%s: dateSeconds = %d, %v; want %d", date, got, err, want)
		}
	}
	for _, date := range []string{
		"2026-03-01T10:00:00+24:00",
		"2026-03-01T10:00:00+01:60",
		"2026-03-01T10:00:00+99:00",
		"2026-03-01T10:00:00+0100",
		"2026-03-01T10:00:00",
		"2026-03-01",
	} {
		path := writeManifest(t, `{"date": "`+date+`", `+base+`}`)
		_, err := loadManifest(path)
		if err == nil || err.Error() != path+": date must be RFC 3339" {
			t.Errorf("%s: error = %v, want 'date must be RFC 3339'", date, err)
		}
	}
}
