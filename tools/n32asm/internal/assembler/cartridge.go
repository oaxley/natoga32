// Cartridge chunk output for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package assembler

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"sort"

	"github.com/natoga32/n32asm/internal/section"
	"github.com/natoga32/n32asm/internal/symbol"
)

// RSCV section kinds
const (
	rscvKindText   = 0x01
	rscvKindData   = 0x02
	rscvKindRodata = 0x03
	rscvKindBss    = 0x04
)

// rscvOrder lists the sections written to an RSCV chunk, in kind order
var rscvOrder = []struct {
	name string
	kind byte
}{
	{".text", rscvKindText},
	{".data", rscvKindData},
	{".rodata", rscvKindRodata},
	{".bss", rscvKindBss},
}

// align4 rounds n up to a multiple of 4
func align4(n int) int {
	return (n + 3) &^ 3
}

// chunk returns a complete cartridge chunk: length, type, data, zero padding to 4, CRC-32 of type+data
func chunk(fourcc string, data []byte) []byte {
	out := make([]byte, 0, 12+align4(len(data)))
	out = binary.LittleEndian.AppendUint32(out, uint32(len(data)))
	out = append(out, fourcc...)
	out = append(out, data...)
	out = append(out, make([]byte, align4(len(data))-len(data))...)
	crc := crc32.ChecksumIEEE(out[4 : 8+len(data)])
	return binary.LittleEndian.AppendUint32(out, crc)
}

// checkRscv rejects sections that can't make a cartridge
func checkRscv(sections *section.Manager) error {
	if text := sections.Get(".text"); text == nil || text.Size == 0 {
		return fmt.Errorf("no .text section: a cartridge needs code")
	}
	if rom := sections.Get(".rom"); rom != nil && rom.Size > 0 {
		return fmt.Errorf(".rom section can't go in a cartridge: use -f rom")
	}
	// RSCV holds only the .bss header: its data would be zeroed
	if bss := sections.Get(".bss"); bss != nil {
		for i, b := range bss.Data {
			if b != 0 {
				return fmt.Errorf(".bss section holds non-zero data at 0x%08X: use .data", uint32(bss.BaseAddr)+uint32(i))
			}
		}
	}
	return nil
}

// buildRscv returns the RSCV chunk of the assembled sections (.text, .data, .rodata, .bss order, non-empty only)
func buildRscv(sections *section.Manager) ([]byte, error) {
	if err := checkRscv(sections); err != nil {
		return nil, err
	}

	var data []byte
	for _, entry := range rscvOrder {
		sec := sections.Get(entry.name)
		if sec == nil || sec.Size == 0 {
			continue
		}
		data = append(data, entry.kind, 0, 0, 0)
		data = binary.LittleEndian.AppendUint32(data, uint32(sec.BaseAddr))
		data = binary.LittleEndian.AppendUint32(data, uint32(sec.Size))
		if entry.kind != rscvKindBss {
			data = append(data, sec.Data...)
			data = append(data, make([]byte, align4(len(sec.Data))-len(sec.Data))...)
		}
	}
	return chunk("RSCV", data), nil
}

// dsymFlagGlobal marks a dSYM entry as a global symbol
const dsymFlagGlobal = 0x01

// maxDsymNameLen is the longest name a dSYM entry can hold (u16 length field)
const maxDsymNameLen = 65535

// buildDsym returns the dSYM chunk: u32 count, then per label u32 address, u8 section kind,
// u8 flags (bit 0 = global), u16 name length, name, zero padding to 4; sorted by address then name.
// A name longer than 65535 bytes is an error (its length would not fit the u16 field).
func buildDsym(symbols *symbol.Table) ([]byte, error) {
	kinds := make(map[string]byte, len(rscvOrder))
	for _, entry := range rscvOrder {
		kinds[entry.name] = entry.kind
	}

	var labels []*symbol.Symbol
	for _, sym := range symbols.DefinedSymbols() {
		if _, ok := kinds[sym.Section]; ok && sym.Type == symbol.SymLabel && sym.Name != "" {
			labels = append(labels, sym)
		}
	}
	sort.Slice(labels, func(i, j int) bool {
		ai, aj := uint32(labels[i].Value), uint32(labels[j].Value)
		if ai != aj {
			return ai < aj
		}
		return labels[i].Name < labels[j].Name
	})

	data := binary.LittleEndian.AppendUint32(nil, uint32(len(labels)))
	for _, sym := range labels {
		if len(sym.Name) > maxDsymNameLen {
			return nil, fmt.Errorf("label %.32q... is %d bytes: dSYM names are at most %d bytes",
				sym.Name, len(sym.Name), maxDsymNameLen)
		}
		var flags byte
		if sym.Global {
			flags |= dsymFlagGlobal
		}
		data = binary.LittleEndian.AppendUint32(data, uint32(sym.Value))
		data = append(data, kinds[sym.Section], flags)
		data = binary.LittleEndian.AppendUint16(data, uint16(len(sym.Name)))
		data = append(data, sym.Name...)
		data = append(data, make([]byte, align4(len(sym.Name))-len(sym.Name))...)
	}
	return chunk("dSYM", data), nil
}
