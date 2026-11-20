// Cartridge chunk output for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"

	"github.com/natoga32/goasm/internal/section"
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
