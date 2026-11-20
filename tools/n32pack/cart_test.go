// Test reader of packed cartridges for the NATOGA32 cartridge packer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"
)

// indxEntry is one decoded INDX entry
type indxEntry struct {
	typ    string
	id     uint16
	flags  uint16
	offset uint32
	length uint32
}

// cartChunk is one chunk of a cartridge: its type, the file offset of its data and its data
type cartChunk struct {
	typ    string
	offset uint32
	data   []byte
}

// cart is a decoded cartridge
type cart struct {
	flags  uint16
	ihdr   []byte
	indx   []indxEntry
	chunks []cartChunk // every chunk, IHDR to IEND
}

// readCart checks the header and every chunk of file and decodes it
func readCart(t *testing.T, file []byte) cart {
	t.Helper()
	if len(file) < 16 {
		t.Fatalf("cartridge is %d bytes, too short", len(file))
	}
	if !bytes.Equal(file[0:4], []byte{0x7F, 'N', '3', '2'}) {
		t.Fatalf("bad magic % X", file[0:4])
	}
	if file[4] != 0x10 || file[5] != 0 {
		t.Fatalf("bad version/reserved %02X %02X", file[4], file[5])
	}
	if size := binary.LittleEndian.Uint32(file[8:]); size != uint32(len(file)) {
		t.Fatalf("size field %d, file is %d bytes", size, len(file))
	}
	if crc := binary.LittleEndian.Uint32(file[12:]); crc != crc32.ChecksumIEEE(file[0:12]) {
		t.Fatalf("bad header CRC")
	}
	c := cart{flags: binary.LittleEndian.Uint16(file[6:])}

	for off := uint64(16); off < uint64(len(file)); {
		if off+12 > uint64(len(file)) {
			t.Fatalf("chunk at %d runs past the end", off)
		}
		length := uint64(binary.LittleEndian.Uint32(file[off:]))
		typ := string(file[off+4 : off+8])
		end := off + 12 + align4(length)
		if end > uint64(len(file)) {
			t.Fatalf("%s chunk at %d runs past the end", typ, off)
		}
		data := file[off+8 : off+8+length]
		for _, b := range file[off+8+length : end-4] {
			if b != 0 {
				t.Fatalf("%s chunk at %d: padding is not zero", typ, off)
			}
		}
		if crc := binary.LittleEndian.Uint32(file[end-4:]); crc != crc32.ChecksumIEEE(file[off+4:off+8+length]) {
			t.Fatalf("%s chunk at %d: bad CRC", typ, off)
		}
		c.chunks = append(c.chunks, cartChunk{typ, uint32(off + 8), data})
		off = end
	}

	if len(c.chunks) < 2 || c.chunks[0].typ != "IHDR" || c.chunks[1].typ != "INDX" {
		t.Fatalf("cartridge does not start with IHDR, INDX")
	}
	c.ihdr = c.chunks[0].data
	indx := c.chunks[1].data
	if len(indx)%16 != 0 {
		t.Fatalf("INDX length %d is not a multiple of 16", len(indx))
	}
	for i := 0; i < len(indx); i += 16 {
		e := indx[i : i+16]
		c.indx = append(c.indx, indxEntry{
			typ:    string(e[0:4]),
			id:     binary.LittleEndian.Uint16(e[4:]),
			flags:  binary.LittleEndian.Uint16(e[6:]),
			offset: binary.LittleEndian.Uint32(e[8:]),
			length: binary.LittleEndian.Uint32(e[12:]),
		})
	}
	return c
}
