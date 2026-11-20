// Tests for the RSCV chunk reader of the NATOGA32 disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"encoding/binary"
	"strings"
	"testing"
)

// rscvSectionBytes returns one RSCV section: header, then data padded to 4
func rscvSectionBytes(kind, flags byte, addr, size uint32, data []byte) []byte {
	out := []byte{kind, flags, 0, 0}
	out = binary.LittleEndian.AppendUint32(out, addr)
	out = binary.LittleEndian.AppendUint32(out, size)
	out = append(out, data...)
	return append(out, make([]byte, align4(len(data))-len(data))...)
}

func TestParseRscvChunkSections(t *testing.T) {
	data := rscvSectionBytes(rscvKindText, 0, 0x01A00000, 4, []byte{0x13, 0, 0, 0})
	data = append(data, rscvSectionBytes(rscvKindRodata, 0, 0x01200000, 3, []byte{1, 2, 3})...)
	data = append(data, rscvSectionBytes(rscvKindBss, 0, 0x01000000, 16, nil)...)

	sections, err := parseRscvChunk(chunk("RSCV", data))
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 3 {
		t.Fatalf("got %d sections, want 3", len(sections))
	}
	want := []struct {
		name string
		addr uint32
		size uint32
		data int
	}{
		{".text", 0x01A00000, 4, 4},
		{".rodata", 0x01200000, 3, 3},
		{".bss", 0x01000000, 16, 0},
	}
	for i, w := range want {
		s := sections[i]
		if s.Name != w.name || s.Addr != w.addr || s.Size != w.size || len(s.Data) != w.data {
			t.Errorf("section %d = %s 0x%08X %d (%d data bytes), want %s 0x%08X %d (%d)",
				i, s.Name, s.Addr, s.Size, len(s.Data), w.name, w.addr, w.size, w.data)
		}
	}
	if sections[2].Data != nil {
		t.Errorf(".bss must have nil data")
	}
}

func TestParseRscvChunkErrors(t *testing.T) {
	text := rscvSectionBytes(rscvKindText, 0, 0x01A00000, 4, []byte{0x13, 0, 0, 0})
	good := chunk("RSCV", text)

	tests := []struct {
		name  string
		chunk []byte
		want  string
	}{
		{"truncated file", good[:len(good)-1], "chunk length"},
		{"extra bytes", append(append([]byte{}, good...), 0, 0, 0, 0), "chunk length"},
		{"bad CRC", func() []byte { c := append([]byte{}, good...); c[8] ^= 1; return c }(), "CRC mismatch"},
		{"unknown kind", chunk("RSCV", rscvSectionBytes(0x05, 0, 0, 0, nil)), "unknown section kind 0x05"},
		{"non-zero flags", chunk("RSCV", rscvSectionBytes(rscvKindText, 1, 0x01A00000, 4, []byte{0x13, 0, 0, 0})), "non-zero flags"},
		{"non-zero reserved byte 2", func() []byte {
			s := append([]byte{}, text...)
			s[2] = 1
			return chunk("RSCV", s)
		}(), "non-zero flags or reserved bytes"},
		{"non-zero reserved byte 3", func() []byte {
			s := append([]byte{}, text...)
			s[3] = 1
			return chunk("RSCV", s)
		}(), "non-zero flags or reserved bytes"},
		{"section too big", chunk("RSCV", rscvSectionBytes(rscvKindText, 0, 0x01A00000, 8, []byte{0x13, 0, 0, 0})), "don't fit"},
		{"truncated header", chunk("RSCV", text[:8]), "truncated section header"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseRscvChunk(tt.chunk)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("got %v, want an error containing %q", err, tt.want)
			}
		})
	}
}
