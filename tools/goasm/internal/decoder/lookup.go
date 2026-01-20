// Reverse opcode lookup for RISC-V instruction decoding
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package decoder

import "github.com/natoga32/goasm/internal/arch"

// OpcodeKey is the key for reverse opcode lookup
type OpcodeKey struct {
	Opcode uint32
	Funct3 uint32
	Funct7 uint32
	Rs2    uint32 // Used for special instructions like mret, wfi
}

// OpcodeEntry contains the mnemonic and instruction info
type OpcodeEntry struct {
	Mnemonic string
	Info     arch.InstrOpcode
}

// reverseOpcodeTable maps encoding fields to instruction mnemonics
var reverseOpcodeTable map[OpcodeKey]*OpcodeEntry

// useFunct7Opcodes are opcodes where funct7 is needed for disambiguation
var useFunct7Opcodes = map[uint32]bool{
	0x33: true, // R-type ALU
	0x13: true, // I-type ALU (for shifts and Zbb)
	0x73: true, // System instructions
	0x0B: true, // Thread instructions
}

// useRs2Opcodes are opcodes where rs2 is needed for disambiguation
var useRs2Opcodes = map[uint32]bool{
	0x13: true, // For Zbb instructions like sext.b, sext.h, clz, ctz, cpop
	0x73: true, // For mret, wfi
}

func init() {
	reverseOpcodeTable = make(map[OpcodeKey]*OpcodeEntry)

	for mnemonic, info := range arch.Opcodes {
		key := buildKey(info)
		reverseOpcodeTable[key] = &OpcodeEntry{
			Mnemonic: mnemonic,
			Info:     info,
		}
	}
}

// buildKey creates a lookup key from opcode info
func buildKey(info arch.InstrOpcode) OpcodeKey {
	key := OpcodeKey{
		Opcode: info.Opcode,
		Funct3: info.Funct3,
	}

	// Include funct7 for opcodes that need it
	if useFunct7Opcodes[info.Opcode] {
		key.Funct7 = info.Funct7
	}

	// Include rs2 for special instructions
	if useRs2Opcodes[info.Opcode] && info.Rs2 != 0 {
		key.Rs2 = info.Rs2
	}

	return key
}

// LookupInstruction finds the instruction mnemonic from encoded fields
func LookupInstruction(word uint32) (*OpcodeEntry, bool) {
	opcode := extractOpcode(word)
	funct3 := extractFunct3(word)
	funct7 := extractFunct7(word)
	rs2 := extractRs2(word)

	// Try with all fields first (most specific)
	if useFunct7Opcodes[opcode] && useRs2Opcodes[opcode] {
		key := OpcodeKey{Opcode: opcode, Funct3: funct3, Funct7: funct7, Rs2: rs2}
		if entry, ok := reverseOpcodeTable[key]; ok {
			return entry, true
		}
	}

	// Try with funct7 but no rs2
	if useFunct7Opcodes[opcode] {
		key := OpcodeKey{Opcode: opcode, Funct3: funct3, Funct7: funct7}
		if entry, ok := reverseOpcodeTable[key]; ok {
			return entry, true
		}
	}

	// Try with just opcode and funct3
	key := OpcodeKey{Opcode: opcode, Funct3: funct3}
	if entry, ok := reverseOpcodeTable[key]; ok {
		return entry, true
	}

	// Try with just opcode (for U-type and J-type)
	key = OpcodeKey{Opcode: opcode}
	if entry, ok := reverseOpcodeTable[key]; ok {
		return entry, true
	}

	return nil, false
}

// GetInstructionType determines the instruction type from the opcode
func GetInstructionType(opcode uint32) arch.InstrType {
	switch opcode {
	case 0x03: // Load
		return arch.TypeI
	case 0x0B: // Thread (mixed types)
		return arch.TypeR // Will be refined by lookup
	case 0x13: // ALU immediate
		return arch.TypeI // Will be refined by lookup
	case 0x17: // AUIPC
		return arch.TypeU
	case 0x23: // Store
		return arch.TypeS
	case 0x33: // ALU register
		return arch.TypeR
	case 0x37: // LUI
		return arch.TypeU
	case 0x63: // Branch
		return arch.TypeB
	case 0x67: // JALR
		return arch.TypeI
	case 0x6F: // JAL
		return arch.TypeJ
	case 0x73: // System
		return arch.TypeI
	default:
		return arch.TypeR // Default
	}
}
