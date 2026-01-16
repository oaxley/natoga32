// RISC-V register definitions for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package arch

import (
	"fmt"
	"strings"
)

// Registers maps register names to their numeric values
var Registers = make(map[string]uint32)

func init() {
	// Standard registers x0 ... x31
	for i := uint32(0); i < 32; i++ {
		Registers[regName("x", i)] = i
	}

	// ABI aliases
	// x0 = zero (hardwired zero)
	// x1 = ra (return address)
	// x2 = sp (stack pointer)
	// x3 = gp (global pointer)
	// x4 = tp (thread pointer)
	// x5-x7 = t0-t2 (temporaries)
	// x8 = s0/fp (saved register / frame pointer)
	// x9 = s1 (saved register)
	// x10-x17 = a0-a7 (function arguments / return values)
	// x18-x27 = s2-s11 (saved registers)
	// x28-x31 = t3-t6 (temporaries)

	aliases := []struct {
		name string
		reg  uint32
	}{
		{"zero", 0},
		{"ra", 1},
		{"sp", 2},
		{"gp", 3},
		{"tp", 4},
		{"t0", 5},
		{"t1", 6},
		{"t2", 7},
		{"s0", 8},
		{"fp", 8}, // fp is alias for s0
		{"s1", 9},
		{"a0", 10},
		{"a1", 11},
		{"a2", 12},
		{"a3", 13},
		{"a4", 14},
		{"a5", 15},
		{"a6", 16},
		{"a7", 17},
		{"s2", 18},
		{"s3", 19},
		{"s4", 20},
		{"s5", 21},
		{"s6", 22},
		{"s7", 23},
		{"s8", 24},
		{"s9", 25},
		{"s10", 26},
		{"s11", 27},
		{"t3", 28},
		{"t4", 29},
		{"t5", 30},
		{"t6", 31},
	}

	for _, a := range aliases {
		Registers[a.name] = a.reg
	}
}

// regName creates a register name like "x0", "x1", etc.
func regName(prefix string, n uint32) string {
	return fmt.Sprintf("%s%d", prefix, n)
}

// LookupRegister returns the register number for a given name.
// Returns (reg, true) if found, (0, false) if not a valid register.
func LookupRegister(name string) (uint32, bool) {
	// Normalize to lowercase
	name = strings.ToLower(name)
	reg, ok := Registers[name]
	return reg, ok
}

// IsRegister checks if the given name is a valid register
func IsRegister(name string) bool {
	_, ok := LookupRegister(name)
	return ok
}
