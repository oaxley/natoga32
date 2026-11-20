// dSYM symbol checks for the NATOGA32 cartridge packer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"encoding/binary"
	"fmt"
)

// dsymEntryHeaderSize is the size of a dSYM entry before its name
const dsymEntryHeaderSize = 8

// parseDsym checks the dSYM chunk data and returns label -> address
func parseDsym(data []byte) (map[string]uint32, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("dSYM: chunk too short for the entry count")
	}
	count := binary.LittleEndian.Uint32(data)
	symbols := make(map[string]uint32)
	off := uint64(4)
	size := uint64(len(data))
	for i := uint32(0); i < count; i++ {
		if size-off < dsymEntryHeaderSize {
			return nil, fmt.Errorf("dSYM: entry %d runs past the end of the chunk", i)
		}
		h := data[off : off+dsymEntryHeaderSize]
		addr := binary.LittleEndian.Uint32(h)
		kind := h[4]
		nameLen := uint64(binary.LittleEndian.Uint16(h[6:]))
		if kind < kindText || kind > kindBss {
			return nil, fmt.Errorf("dSYM: entry %d: unknown section kind 0x%02X", i, kind)
		}
		if nameLen == 0 {
			return nil, fmt.Errorf("dSYM: entry %d: empty name", i)
		}
		off += dsymEntryHeaderSize
		if size-off < align4(nameLen) {
			return nil, fmt.Errorf("dSYM: entry %d: name runs past the end of the chunk", i)
		}
		name := string(data[off : off+nameLen])
		if _, dup := symbols[name]; dup {
			return nil, fmt.Errorf("dSYM: duplicate symbol %q", name)
		}
		symbols[name] = addr
		off += align4(nameLen)
	}
	if off != size {
		return nil, fmt.Errorf("dSYM: %d trailing bytes after %d entries", size-off, count)
	}
	return symbols, nil
}
