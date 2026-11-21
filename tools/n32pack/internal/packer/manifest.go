// Cartridge manifest loading for the NATOGA32 cartridge packer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package packer

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

	file string // path of the manifest; names it in pack errors
	dir  string // directory of the manifest; paths are relative to it
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
	// RFC 3339 time offset: Z, or +-hh:mm with hh 00-23 and mm 00-59
	dateOffsetRe = regexp.MustCompile(`(Z|[+-]([01][0-9]|2[0-3]):[0-5][0-9])$`)
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

	if err := checkTypes(file); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
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
	m.file = path
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
		date, err := parseDate(m.Date)
		if err != nil {
			return err
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
		if date, err = parseDate(m.Date); err != nil {
			return 0, err
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

// parseDate parses an RFC 3339 date. time.Parse also takes offsets such as
// +24:00 or +01:60, which RFC 3339 does not allow: they are refused here.
func parseDate(s string) (time.Time, error) {
	date, err := time.Parse(time.RFC3339, s)
	if err != nil || !dateOffsetRe.MatchString(s) {
		return time.Time{}, errors.New("date must be RFC 3339")
	}
	return date, nil
}

// checkTypes checks the JSON type of every manifest field, and the range of
// the asset numbers, so that a wrong value gets a plain message instead of
// the decoder's "json: cannot unmarshal ..." text. A null value is the same
// as an absent field. Syntax errors and unknown fields are left to the decoder.
func checkTypes(file []byte) error {
	dec := json.NewDecoder(bytes.NewReader(file))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return nil // reported by the strict decoding
	}
	top, ok := doc.(map[string]any)
	if !ok {
		return errors.New("the manifest must be a JSON object")
	}
	for _, name := range []string{"name", "version", "min_rom", "date", "rscv", "dsym", "entry"} {
		if v, ok := top[name]; ok && v != nil {
			if _, ok := v.(string); !ok {
				return fmt.Errorf("%s must be a string", name)
			}
		}
	}
	for _, name := range []string{"regions", "languages"} {
		if v, ok := top[name]; ok && v != nil && !isStringList(v) {
			return fmt.Errorf("%s must be a list of strings", name)
		}
	}
	if v, ok := top["graphics_mode"]; ok && v != nil {
		if _, ok := jsonInt(v); !ok {
			return errors.New("graphics_mode must be 0 to 3")
		}
	}
	v, ok := top["assets"]
	if !ok || v == nil {
		return nil
	}
	assets, ok := v.([]any)
	if !ok {
		return errors.New("assets must be a list of objects")
	}
	for i, item := range assets {
		asset, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("assets[%d] must be an object", i)
		}
		if err := checkAssetTypes(asset); err != nil {
			// name the asset when it has a name, else its index
			if name, ok := asset["name"].(string); ok {
				return fmt.Errorf("asset %q: %w", name, err)
			}
			return fmt.Errorf("assets[%d]: %w", i, err)
		}
	}
	return nil
}

// checkAssetTypes checks the fields of one asset
func checkAssetTypes(asset map[string]any) error {
	for _, name := range []string{"name", "type", "file"} {
		if v, ok := asset[name]; ok && v != nil {
			if _, ok := v.(string); !ok {
				return fmt.Errorf("%s must be a string", name)
			}
		}
	}
	for _, f := range []struct {
		name string
		max  int64
	}{{"slot", 0xFFFF}, {"format", 0xFF}, {"count", 0xFFFF}} {
		if v, ok := asset[f.name]; ok && v != nil {
			if n, ok := jsonInt(v); !ok || n < 0 || n > f.max {
				return fmt.Errorf("%s must be an integer from 0 to %d", f.name, f.max)
			}
		}
	}
	if v, ok := asset["boot"]; ok && v != nil {
		if _, ok := v.(bool); !ok {
			return errors.New("boot must be true or false")
		}
	}
	return nil
}

// jsonInt returns the value of a JSON integer (decoded with UseNumber)
func jsonInt(v any) (int64, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, false
	}
	i, err := strconv.ParseInt(n.String(), 10, 64)
	return i, err == nil
}

// isStringList reports whether v is a JSON list of strings
func isStringList(v any) bool {
	list, ok := v.([]any)
	if !ok {
		return false
	}
	for _, item := range list {
		if _, ok := item.(string); !ok {
			return false
		}
	}
	return true
}
