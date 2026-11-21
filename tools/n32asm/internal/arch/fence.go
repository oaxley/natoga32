// FENCE predecessor and successor sets for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package arch

import "strings"

// Bits of a fence set: device input, device output, memory read, memory write
const (
	FenceI uint32 = 0b1000
	FenceO uint32 = 0b0100
	FenceR uint32 = 0b0010
	FenceW uint32 = 0b0001

	// FenceAll is the set "iorw", used when fence has no operand
	FenceAll = FenceI | FenceO | FenceR | FenceW
)

// fenceLetters lists the letters of a set, in the order they are printed
var fenceLetters = []struct {
	letter byte
	bit    uint32
}{
	{'i', FenceI},
	{'o', FenceO},
	{'r', FenceR},
	{'w', FenceW},
}

// ParseFenceSet converts a set such as "rw" or "iorw" into its 4-bit value.
// The letters can come in any order, each at most once. "0" is the empty set.
func ParseFenceSet(s string) (uint32, bool) {
	if s == "0" {
		return 0, true
	}
	if s == "" {
		return 0, false
	}
	set := uint32(0)
	for i := 0; i < len(s); i++ {
		bit := uint32(0)
		for _, fl := range fenceLetters {
			if s[i] == fl.letter {
				bit = fl.bit
			}
		}
		if bit == 0 || set&bit != 0 {
			return 0, false
		}
		set |= bit
	}
	return set, true
}

// FenceSetString returns the letters of a 4-bit fence set in iorw order, "0" when empty
func FenceSetString(set uint32) string {
	var sb strings.Builder
	for _, fl := range fenceLetters {
		if set&fl.bit != 0 {
			sb.WriteByte(fl.letter)
		}
	}
	if sb.Len() == 0 {
		return "0"
	}
	return sb.String()
}
