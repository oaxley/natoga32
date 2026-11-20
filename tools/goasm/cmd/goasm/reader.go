// Binary file reader for the NATOGA32 disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
)

// ReadSection is one section of a binary file
type ReadSection struct {
	Kind uint8  // RSCV kind (rscvKindText for raw binaries)
	Name string // ".text", ".data", ".rodata", ".bss"
	Addr uint32
	Size uint32
	Data []byte // nil for .bss
}

// BinaryReader reads and parses binary files for disassembly
type BinaryReader struct {
	Format   string        // "bin" or "rscv"
	Sections []ReadSection // file order
}

// rscvSectionHeaderSize is the size of an RSCV section header: kind, flags, reserved, address, size
const rscvSectionHeaderSize = 12

// NewBinaryReader creates a new BinaryReader from a file
// A file whose bytes 4-7 are "RSCV" is read as a cartridge RSCV chunk;
// anything else is a raw binary loaded as .text at baseAddr
func NewBinaryReader(filename string, baseAddr uint32) (*BinaryReader, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	if len(data) >= 8 && string(data[4:8]) == "RSCV" {
		sections, err := parseRscvChunk(data)
		if err != nil {
			return nil, fmt.Errorf("invalid RSCV chunk: %w", err)
		}
		return &BinaryReader{Format: "rscv", Sections: sections}, nil
	}

	// Raw binary format: the whole file is .text
	return &BinaryReader{
		Format: "bin",
		Sections: []ReadSection{{
			Kind: rscvKindText,
			Name: ".text",
			Addr: baseAddr,
			Size: uint32(len(data)),
			Data: data,
		}},
	}, nil
}

// rscvKindName returns the section name of an RSCV kind, or "" for an unknown kind
func rscvKindName(kind uint8) string {
	for _, entry := range rscvOrder {
		if entry.kind == kind {
			return entry.name
		}
	}
	return ""
}

// parseRscvChunk checks a complete RSCV chunk (size, CRC) and returns its sections
func parseRscvChunk(c []byte) ([]ReadSection, error) {
	if len(c) < 12 {
		return nil, fmt.Errorf("file is %d bytes, too small for a chunk", len(c))
	}
	length := binary.LittleEndian.Uint32(c[0:4])
	if size := 12 + uint64(align4(int(length))); uint64(len(c)) != size {
		return nil, fmt.Errorf("file is %d bytes, chunk length %d needs %d", len(c), length, size)
	}
	want := binary.LittleEndian.Uint32(c[len(c)-4:])
	if got := crc32.ChecksumIEEE(c[4 : 8+length]); got != want {
		return nil, fmt.Errorf("CRC mismatch: computed 0x%08X, stored 0x%08X", got, want)
	}

	data := c[8 : 8+length]
	var sections []ReadSection
	for off := 0; off < len(data); {
		if len(data)-off < rscvSectionHeaderSize {
			return nil, fmt.Errorf("truncated section header at offset %d", off)
		}
		hdr := data[off : off+rscvSectionHeaderSize]
		kind := hdr[0]
		name := rscvKindName(kind)
		if name == "" {
			return nil, fmt.Errorf("unknown section kind 0x%02X at offset %d", kind, off)
		}
		if hdr[1] != 0 || hdr[2] != 0 || hdr[3] != 0 {
			return nil, fmt.Errorf("%s: non-zero flags or reserved bytes", name)
		}
		sec := ReadSection{
			Kind: kind,
			Name: name,
			Addr: binary.LittleEndian.Uint32(hdr[4:8]),
			Size: binary.LittleEndian.Uint32(hdr[8:12]),
		}
		off += rscvSectionHeaderSize
		if kind != rscvKindBss {
			padded := uint64(align4(int(sec.Size)))
			if uint64(sec.Size) > uint64(len(data)-off) || padded > uint64(len(data)-off) {
				return nil, fmt.Errorf("%s: %d bytes don't fit in the chunk", name, sec.Size)
			}
			sec.Data = data[off : off+int(sec.Size)]
			off += int(padded)
		}
		sections = append(sections, sec)
	}
	return sections, nil
}
