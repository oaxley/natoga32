// Binary file reader for the NATOGA32 disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// BinaryReader reads and parses binary files for disassembly
type BinaryReader struct {
	Data     []byte // Raw file data
	Format   string // "bin" or "on32"
	TextBase uint32 // Base address for .text section
	TextSize uint32 // Size of .text section
	DataBase uint32 // Base address for .data section
	DataSize uint32 // Size of .data section
}

const (
	on32HeaderSize = 16
	on32Magic      = "ON32"
)

// NewBinaryReader creates a new BinaryReader from a file
// For raw binary files, baseAddr specifies the base address
// For ON32 files, the base address is determined from the format
func NewBinaryReader(filename string, baseAddr uint32) (*BinaryReader, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	reader := &BinaryReader{
		Data: data,
	}

	// Check for ON32 magic header
	if reader.isOn32Format() {
		if err := reader.parseOn32Header(); err != nil {
			return nil, err
		}
	} else {
		// Raw binary format
		reader.Format = "bin"
		reader.TextBase = baseAddr
		reader.TextSize = uint32(len(data))
		reader.DataBase = 0
		reader.DataSize = 0
	}

	return reader, nil
}

// isOn32Format checks if the data starts with ON32 magic header
func (r *BinaryReader) isOn32Format() bool {
	if len(r.Data) < on32HeaderSize {
		return false
	}
	return string(r.Data[0:4]) == on32Magic
}

// parseOn32Header parses the ON32 header and extracts section information
// Header format (16 bytes):
//   - Bytes 0-3: Magic "ON32"
//   - Bytes 4-7: Text section size (little-endian uint32)
//   - Bytes 8-11: Data section size (little-endian uint32)
//   - Bytes 12-15: Reserved (unused)
func (r *BinaryReader) parseOn32Header() error {
	if len(r.Data) < on32HeaderSize {
		return fmt.Errorf("file too small for ON32 header")
	}

	r.Format = "on32"
	r.TextSize = binary.LittleEndian.Uint32(r.Data[4:8])
	r.DataSize = binary.LittleEndian.Uint32(r.Data[8:12])

	// Use standard memory map addresses
	r.TextBase = MMAP_TEXT_SEGMENT
	r.DataBase = MMAP_DATA_SEGMENT

	// Validate section sizes
	expectedSize := uint32(on32HeaderSize) + r.TextSize + r.DataSize
	if uint32(len(r.Data)) < expectedSize {
		return fmt.Errorf("file truncated: expected %d bytes, got %d", expectedSize, len(r.Data))
	}

	return nil
}

// TextSection returns the .text section data
func (r *BinaryReader) TextSection() []byte {
	if r.Format == "on32" {
		start := on32HeaderSize
		end := on32HeaderSize + int(r.TextSize)
		if end > len(r.Data) {
			end = len(r.Data)
		}
		return r.Data[start:end]
	}
	// Raw binary: entire file is .text
	return r.Data
}

// DataSection returns the .data section data
func (r *BinaryReader) DataSection() []byte {
	if r.Format == "on32" && r.DataSize > 0 {
		start := on32HeaderSize + int(r.TextSize)
		end := start + int(r.DataSize)
		if end > len(r.Data) {
			end = len(r.Data)
		}
		return r.Data[start:end]
	}
	return nil
}

// HasDataSection returns true if the file contains a .data section
func (r *BinaryReader) HasDataSection() bool {
	return r.DataSize > 0
}
