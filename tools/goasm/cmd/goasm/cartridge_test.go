// Tests for the cartridge chunk output of the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"strings"
	"testing"
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
