// RSCV section checks and entry point resolution for the NATOGA32 cartridge packer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// Section is one RSCV section: its kind, load address and size
type Section struct {
	Kind uint8
	Addr uint32
	Size uint32
}

// RSCV section kinds
const (
	kindText   = 0x01
	kindData   = 0x02
	kindRodata = 0x03
	kindBss    = 0x04
)

// rscvHeaderSize is the size of an RSCV section header
const rscvHeaderSize = 12

// region is the memory region of a section kind (inclusive start, exclusive end)
type region struct {
	name       string
	start, end uint64
}

var regions = map[uint8]region{
	kindText:   {".text", 0x01A00000, 0x02000000},
	kindData:   {".data", 0x01400000, 0x01A00000},
	kindRodata: {".rodata", 0x01200000, 0x01400000},
	kindBss:    {".bss", 0x01000000, 0x01200000},
}

// parseRscv checks the RSCV chunk data and returns its sections
func parseRscv(data []byte) ([]Section, error) {
	var sections []Section
	hasText := false
	for off, n := uint64(0), 0; off < uint64(len(data)); n++ {
		rest := uint64(len(data)) - off
		if rest < rscvHeaderSize {
			return nil, fmt.Errorf("RSCV: section %d runs past the end of the chunk", n)
		}
		h := data[off : off+rscvHeaderSize]
		sec := Section{Kind: h[0], Addr: binary.LittleEndian.Uint32(h[4:]), Size: binary.LittleEndian.Uint32(h[8:])}
		reg, ok := regions[sec.Kind]
		if !ok {
			return nil, fmt.Errorf("RSCV: section %d: unknown kind 0x%02X", n, sec.Kind)
		}
		if h[1] != 0 || h[2] != 0 || h[3] != 0 {
			return nil, fmt.Errorf("RSCV: section %d: reserved fields must be 0", n)
		}
		total := uint64(rscvHeaderSize)
		if sec.Kind != kindBss {
			total += align4(uint64(sec.Size))
		}
		if total > rest {
			return nil, fmt.Errorf("RSCV: section %d runs past the end of the chunk", n)
		}
		end := uint64(sec.Addr) + uint64(sec.Size)
		if uint64(sec.Addr) < reg.start || end > reg.end {
			return nil, fmt.Errorf("RSCV: section %d (%s at 0x%08X, %d bytes) is outside its region 0x%08X-0x%08X",
				n, reg.name, sec.Addr, sec.Size, reg.start, reg.end-1)
		}
		for k, prev := range sections {
			if overlaps(prev, sec) {
				return nil, fmt.Errorf("RSCV: section %d overlaps section %d", n, k)
			}
		}
		if sec.Kind == kindText {
			hasText = true
		}
		sections = append(sections, sec)
		off += total
	}
	if !hasText {
		return nil, fmt.Errorf("RSCV: no .text section")
	}
	return sections, nil
}

// overlaps reports whether two sections share at least one byte; a size-0 section overlaps nothing
func overlaps(a, b Section) bool {
	if a.Size == 0 || b.Size == 0 {
		return false
	}
	return uint64(a.Addr) < uint64(b.Addr)+uint64(b.Size) && uint64(b.Addr) < uint64(a.Addr)+uint64(a.Size)
}

// entryPoint resolves the manifest entry against the sections and symbols (symbols may be nil)
func entryPoint(entry string, sections []Section, symbols map[string]uint32) (uint32, error) {
	var addr uint32
	switch {
	case entry == "":
		found := false
		for _, sec := range sections {
			if sec.Kind == kindText {
				addr, found = sec.Addr, true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("entry: no .text section")
		}
	case strings.HasPrefix(entry, "0x") || strings.HasPrefix(entry, "0X"):
		v, err := strconv.ParseUint(entry[2:], 16, 32)
		if err != nil {
			return 0, fmt.Errorf("entry %q: %v", entry, err)
		}
		addr = uint32(v)
	default:
		if symbols == nil {
			return 0, fmt.Errorf("entry %q: no dsym file given", entry)
		}
		v, ok := symbols[entry]
		if !ok {
			return 0, fmt.Errorf("entry %q: unknown symbol", entry)
		}
		addr = v
	}
	if addr%4 != 0 {
		return 0, fmt.Errorf("entry 0x%08X is not 4-byte aligned", addr)
	}
	for _, sec := range sections {
		if sec.Kind == kindText && addr >= sec.Addr && uint64(addr) < uint64(sec.Addr)+uint64(sec.Size) {
			return addr, nil
		}
	}
	return 0, fmt.Errorf("entry 0x%08X is not in a .text section", addr)
}
