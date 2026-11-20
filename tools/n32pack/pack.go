// Cartridge packing for the NATOGA32 cartridge packer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"time"
)

// Cartridge limits and sizes
const (
	headerSize     = 16
	chunkOverhead  = 12 // u32 length, type, u32 CRC
	indxEntrySize  = 16
	maxIndxEntries = 15872
	maxCartSize    = 256 << 20
	formatVersion  = 0x10
	assetHeadSize  = 8
)

// cartMagic starts every cartridge
var cartMagic = []byte{0x7F, 'N', '3', '2'}

// packChunk is one chunk listed in INDX
type packChunk struct {
	typ   string
	data  []byte
	flags uint16
}

// checkLimits checks the INDX entry count and the total file size
func checkLimits(entries int, size uint64) error {
	if entries > maxIndxEntries {
		return fmt.Errorf("too many chunks: %d (max %d)", entries, maxIndxEntries)
	}
	if size > maxCartSize {
		return fmt.Errorf("cartridge is %d bytes, over the 256 MiB limit", size)
	}
	return nil
}

// pack reads the rscv/dsym/asset files named by the manifest and returns the cartridge bytes;
// each error names its file once: the manifest, or the rscv/dsym/asset file
func pack(m *Manifest, now time.Time) ([]byte, error) {
	entries := 1 + len(m.Assets)
	if m.Dsym != "" {
		entries++
	}
	if err := checkLimits(entries, 0); err != nil {
		return nil, fmt.Errorf("%s: %w", m.file, err)
	}

	rscv, err := readChunkFile(m.path(m.Rscv), "RSCV")
	if err != nil {
		return nil, err
	}
	sections, err := parseRscv(rscv)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", m.path(m.Rscv), err)
	}
	var dsym []byte
	var symbols map[string]uint32
	if m.Dsym != "" {
		if dsym, err = readChunkFile(m.path(m.Dsym), "dSYM"); err != nil {
			return nil, err
		}
		if symbols, err = parseDsym(dsym); err != nil {
			return nil, fmt.Errorf("%s: %w", m.path(m.Dsym), err)
		}
	}
	entry, err := entryPoint(m.Entry, sections, symbols)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", m.file, err)
	}
	ihdr, err := buildIhdr(m, entry, now)
	if err != nil {
		return nil, err
	}

	chunks := make([]packChunk, 0, entries)
	chunks = append(chunks, packChunk{"RSCV", rscv, 1})
	for _, a := range m.Assets {
		payload, err := os.ReadFile(m.path(a.File))
		if err != nil {
			return nil, fmt.Errorf("asset %q: %w", a.Name, err)
		}
		data := make([]byte, assetHeadSize, assetHeadSize+len(payload))
		binary.LittleEndian.PutUint16(data[0:], a.Slot)
		data[2] = a.Format
		binary.LittleEndian.PutUint16(data[4:], a.Count)
		data = append(data, payload...)
		var flags uint16
		if a.Boot {
			flags = 1
		}
		chunks = append(chunks, packChunk{a.Type, data, flags})
	}
	if dsym != nil {
		chunks = append(chunks, packChunk{"dSYM", dsym, 0})
	}

	// total size: header, IHDR, INDX, the listed chunks, IEND
	first := uint64(headerSize) + chunkOverhead + ihdrSize + chunkOverhead + uint64(len(chunks))*indxEntrySize
	size := first
	for _, c := range chunks {
		size += chunkOverhead + align4(uint64(len(c.data)))
	}
	size += chunkOverhead
	if err := checkLimits(len(chunks), size); err != nil {
		return nil, fmt.Errorf("%s: %w", m.file, err)
	}

	// INDX: the data of each chunk starts 8 bytes after the chunk
	indx := make([]byte, 0, len(chunks)*indxEntrySize)
	offset := first
	for id, c := range chunks {
		indx = append(indx, c.typ...)
		indx = binary.LittleEndian.AppendUint16(indx, uint16(id))
		indx = binary.LittleEndian.AppendUint16(indx, c.flags)
		indx = binary.LittleEndian.AppendUint32(indx, uint32(offset+8))
		indx = binary.LittleEndian.AppendUint32(indx, uint32(len(c.data)))
		offset += chunkOverhead + align4(uint64(len(c.data)))
	}

	file := make([]byte, headerSize, size)
	copy(file, cartMagic)
	file[4] = formatVersion
	binary.LittleEndian.PutUint16(file[6:], headerFlags(m))
	binary.LittleEndian.PutUint32(file[8:], uint32(size))
	file = appendChunk(file, "IHDR", ihdr)
	file = appendChunk(file, "INDX", indx)
	for _, c := range chunks {
		file = appendChunk(file, c.typ, c.data)
	}
	file = appendChunk(file, "IEND", nil)
	binary.LittleEndian.PutUint32(file[12:], crc32.ChecksumIEEE(file[0:12]))
	return file, nil
}
