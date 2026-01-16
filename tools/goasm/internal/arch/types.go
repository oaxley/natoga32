// RISC-V instruction types for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package arch

// InstrType represents the RISC-V instruction format type
type InstrType int

const (
	TypeR  InstrType = iota // R-type: register-register operations
	TypeI                   // I-type: immediate operations, loads
	TypeI2                  // I-type variant: shift immediate (uses funct7)
	TypeS                   // S-type: stores
	TypeB                   // B-type: conditional branches
	TypeU                   // U-type: upper immediate (lui, auipc)
	TypeJ                   // J-type: unconditional jumps (jal)
)

// String returns the string representation of an instruction type
func (t InstrType) String() string {
	names := [...]string{
		TypeR:  "R",
		TypeI:  "I",
		TypeI2: "I2",
		TypeS:  "S",
		TypeB:  "B",
		TypeU:  "U",
		TypeJ:  "J",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "UNKNOWN"
}

// InstrOpcode contains the encoding information for an instruction
type InstrOpcode struct {
	Type   InstrType // Instruction format type
	Opcode uint32    // 7-bit opcode field
	Rd     uint32    // Default rd value (usually 0)
	Funct3 uint32    // 3-bit function field
	Rs1    uint32    // Default rs1 value (usually 0)
	Rs2    uint32    // Default rs2 value (usually 0)
	Funct7 uint32    // 7-bit function field (R-type, some I-type)
}

// InstrSize is the size of a RISC-V instruction in bytes
const InstrSize = 4

// Endianness for NATOGA32 (little-endian)
const LittleEndian = true
