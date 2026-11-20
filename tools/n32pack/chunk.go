// Cartridge chunk helpers for the NATOGA32 cartridge packer
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

// align4 rounds n up to a multiple of 4
func align4(n uint64) uint64 {
	return (n + 3) &^ 3
}

// appendChunk appends a chunk to dst: u32 length, type, data, zero padding to 4, u32 CRC-32(type + data)
func appendChunk(dst []byte, typ string, data []byte) []byte {
	start := len(dst)
	dst = binary.LittleEndian.AppendUint32(dst, uint32(len(data)))
	dst = append(dst, typ...)
	dst = append(dst, data...)
	crc := crc32.ChecksumIEEE(dst[start+4 : start+8+len(data)])
	dst = append(dst, make([]byte, align4(uint64(len(data)))-uint64(len(data)))...)
	return binary.LittleEndian.AppendUint32(dst, crc)
}

// readChunkFile reads a file holding exactly one chunk of type want; checks length, padding size and CRC
func readChunkFile(path, want string) ([]byte, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(file) < 12 {
		return nil, fmt.Errorf("%s: bad chunk length", path)
	}
	if string(file[4:8]) != want {
		return nil, fmt.Errorf("%s: not a %s chunk", path, want)
	}
	length := uint64(binary.LittleEndian.Uint32(file[0:4]))
	if uint64(len(file)) != 12+align4(length) {
		return nil, fmt.Errorf("%s: bad chunk length", path)
	}
	crc := binary.LittleEndian.Uint32(file[len(file)-4:])
	if crc32.ChecksumIEEE(file[4:8+length]) != crc {
		return nil, fmt.Errorf("%s: bad chunk CRC", path)
	}
	return file[8 : 8+length], nil
}
