// Decoded instruction types for the NATOGA32 disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package decoder

import "github.com/natoga32/goasm/internal/arch"

// DecodedInstruction represents a decoded RISC-V instruction
type DecodedInstruction struct {
	Mnemonic  string         // Instruction mnemonic (e.g., "add", "lw", "beq")
	Type      arch.InstrType // Instruction format type (R, I, S, B, U, J)
	Rd        int            // Destination register (-1 if not used)
	Rs1       int            // Source register 1 (-1 if not used)
	Rs2       int            // Source register 2 (-1 if not used)
	Immediate int32          // Immediate value (sign-extended)
	RawWord   uint32         // Original 32-bit instruction word
}

// NoRegister indicates a register field is not used
const NoRegister = -1

// HasRd returns true if the instruction uses a destination register
func (d *DecodedInstruction) HasRd() bool {
	return d.Rd != NoRegister
}

// HasRs1 returns true if the instruction uses source register 1
func (d *DecodedInstruction) HasRs1() bool {
	return d.Rs1 != NoRegister
}

// HasRs2 returns true if the instruction uses source register 2
func (d *DecodedInstruction) HasRs2() bool {
	return d.Rs2 != NoRegister
}

// HasImmediate returns true if the instruction uses an immediate value
func (d *DecodedInstruction) HasImmediate() bool {
	switch d.Type {
	case arch.TypeI, arch.TypeI2, arch.TypeS, arch.TypeB, arch.TypeU, arch.TypeJ:
		return true
	default:
		return false
	}
}

// IsUnknown returns true if the instruction could not be decoded
func (d *DecodedInstruction) IsUnknown() bool {
	return d.Mnemonic == ""
}
