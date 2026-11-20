// Cartridge manifest loading for the NATOGA32 cartridge packer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
)

// Asset is one asset chunk of the cartridge
type Asset struct {
	Name   string `json:"name"` // ^[A-Z][A-Z0-9_]*$, unique; becomes ASSET_<Name>
	Type   string `json:"type"` // one of the 14 asset chunk types
	Slot   uint16 `json:"slot"`
	Format uint8  `json:"format"`
	Count  uint16 `json:"count"`
	File   string `json:"file"` // payload only
	Boot   bool   `json:"boot"`
}

// Manifest describes the cartridge to build
type Manifest struct {
	Name         string   `json:"name"`          // required, 1-32 bytes of valid UTF-8
	Version      string   `json:"version"`       // "M.mm", e.g. "1.00" -> 0x0100 (BCD); default "1.00"
	MinRom       string   `json:"min_rom"`       // "major.minor", each 0-255; default "0.1"
	Regions      []string `json:"regions"`       // required, non-empty: "americas", "europe", "asia"
	Languages    []string `json:"languages"`     // "english"; may be empty
	Date         string   `json:"date"`          // RFC 3339; default: now
	GraphicsMode *int     `json:"graphics_mode"` // 0-3; absent = no requirement
	Rscv         string   `json:"rscv"`          // required
	Dsym         string   `json:"dsym"`          // optional
	Entry        string   `json:"entry"`         // label or "0x..."; default: first .text section
	Assets       []Asset  `json:"assets"`

	dir string // directory of the manifest; paths are relative to it
}

// IHDR region bits
var regionBit = map[string]uint8{
	"americas": 1 << 0,
	"europe":   1 << 1,
	"asia":     1 << 2,
}

// IHDR language bits
var languageBit = map[string]uint16{
	"english": 1 << 0,
}

// assetTypes are the asset chunk types
var assetTypes = map[string]bool{
	"TILS": true, "TMAP": true, "TPAL": true, "TMAT": true,
	"SPRS": true, "SOAM": true, "SPAL": true, "SMAT": true,
	"FNTS": true, "FPAL": true,
	"ASMD": true, "ATMD": true, "ATFM": true, "AIFM": true,
}

// dateEpoch is the origin of IHDR dates
var dateEpoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

var (
	versionRe   = regexp.MustCompile(`^([0-9]{1,2})\.([0-9]{2})$`)
	minRomRe    = regexp.MustCompile(`^([0-9]{1,3})\.([0-9]{1,3})$`)
	assetNameRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

const (
	defaultVersion = "1.00"
	defaultMinRom  = "0.1"
)

// loadManifest reads and validates the manifest (it does not read rscv/dsym/asset files)
func loadManifest(path string) (*Manifest, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	m := &Manifest{}
	dec := json.NewDecoder(bytes.NewReader(file))
	dec.DisallowUnknownFields()
	if err := dec.Decode(m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if dec.More() {
		return nil, fmt.Errorf("%s: unexpected data after the manifest", path)
	}
	// encoding/json replaces invalid UTF-8 with U+FFFD: check the raw name
	var raw struct {
		Name json.RawMessage `json:"name"`
	}
	if err := json.Unmarshal(file, &raw); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if !utf8.Valid(raw.Name) {
		return nil, fmt.Errorf("%s: name must be 1 to 32 bytes of UTF-8", path)
	}
	if m.Version == "" {
		m.Version = defaultVersion
	}
	if m.MinRom == "" {
		m.MinRom = defaultMinRom
	}
	m.dir = filepath.Dir(path)
	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return m, nil
}

// validate checks the manifest fields
func (m *Manifest) validate() error {
	if len(m.Name) < 1 || len(m.Name) > 32 || !utf8.ValidString(m.Name) {
		return errors.New("name must be 1 to 32 bytes of UTF-8")
	}
	if _, ok := parseVersion(m.Version); !ok {
		return errors.New("version must look like 1.00")
	}
	if _, _, ok := parseMinRom(m.MinRom); !ok {
		return errors.New("min_rom must look like 0.1")
	}
	if len(m.Regions) == 0 {
		return errors.New("regions: at least one region")
	}
	for _, r := range m.Regions {
		if _, ok := regionBit[r]; !ok {
			return fmt.Errorf("regions: unknown region %q", r)
		}
	}
	for _, l := range m.Languages {
		if _, ok := languageBit[l]; !ok {
			return fmt.Errorf("languages: unknown language %q", l)
		}
	}
	if m.Date != "" {
		date, err := time.Parse(time.RFC3339, m.Date)
		if err != nil {
			return errors.New("date must be RFC 3339")
		}
		if date.Before(dateEpoch) {
			return errors.New("date is before 2026-01-01")
		}
	}
	if m.GraphicsMode != nil && (*m.GraphicsMode < 0 || *m.GraphicsMode > 3) {
		return errors.New("graphics_mode must be 0 to 3")
	}
	if m.Rscv == "" {
		return errors.New("rscv is required")
	}
	names := map[string]bool{}
	for _, a := range m.Assets {
		if !assetNameRe.MatchString(a.Name) {
			return fmt.Errorf("asset name %q must match [A-Z][A-Z0-9_]*", a.Name)
		}
		if names[a.Name] {
			return fmt.Errorf("duplicate asset name %q", a.Name)
		}
		names[a.Name] = true
		if !assetTypes[a.Type] {
			return fmt.Errorf("asset %q: unknown type %q", a.Name, a.Type)
		}
		if a.File == "" {
			return fmt.Errorf("asset %q: file is required", a.Name)
		}
	}
	return nil
}

// parseVersion converts "M.mm" to BCD: major in the high byte, the two minor digits in the low byte
func parseVersion(s string) (uint16, bool) {
	match := versionRe.FindStringSubmatch(s)
	if match == nil {
		return 0, false
	}
	bcd := func(digits string) uint16 {
		var v uint16
		for _, d := range digits {
			v = v<<4 | uint16(d-'0')
		}
		return v
	}
	return bcd(match[1])<<8 | bcd(match[2]), true
}

// parseMinRom converts "major.minor" (each 0-255)
func parseMinRom(s string) (major, minor uint8, ok bool) {
	match := minRomRe.FindStringSubmatch(s)
	if match == nil {
		return 0, 0, false
	}
	ma, err1 := strconv.ParseUint(match[1], 10, 8)
	mi, err2 := strconv.ParseUint(match[2], 10, 8)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return uint8(ma), uint8(mi), true
}

// path returns p relative to the manifest's directory, unless p is absolute
func (m *Manifest) path(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(m.dir, p)
}

// versionBCD returns the game version in BCD
func (m *Manifest) versionBCD() uint16 {
	v, _ := parseVersion(m.Version)
	return v
}

// minRom returns the minimum ROM version
func (m *Manifest) minRom() (major, minor uint8) {
	major, minor, _ = parseMinRom(m.MinRom)
	return major, minor
}

// regionBits returns the IHDR region bits
func (m *Manifest) regionBits() uint8 {
	var bits uint8
	for _, r := range m.Regions {
		bits |= regionBit[r]
	}
	return bits
}

// languageBits returns the IHDR language bits
func (m *Manifest) languageBits() uint16 {
	var bits uint16
	for _, l := range m.Languages {
		bits |= languageBit[l]
	}
	return bits
}

// dateSeconds returns the IHDR date: seconds since 2026-01-01 of the manifest date, or of now when not given
func (m *Manifest) dateSeconds(now time.Time) (uint32, error) {
	date := now
	if m.Date != "" {
		var err error
		if date, err = time.Parse(time.RFC3339, m.Date); err != nil {
			return 0, errors.New("date must be RFC 3339")
		}
	}
	if date.Before(dateEpoch) {
		return 0, errors.New("date is before 2026-01-01")
	}
	seconds := int64(date.Sub(dateEpoch) / time.Second)
	if seconds > int64(^uint32(0)) {
		return 0, errors.New("date is too far in the future")
	}
	return uint32(seconds), nil
}
