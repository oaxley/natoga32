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

// zbbUnaryInstructions are Zbb instructions with fixed rs2/shamt values
// These need special handling because they share opcode/funct3/funct7 with
// shift-immediate instructions like binvi, but have specific rs2 values.
var zbbUnaryInstructions = map[string]bool{
	"clz":    true, // rs2=0b00000
	"ctz":    true, // rs2=0b00001
	"cpop":   true, // rs2=0b00010
	"sext.b": true, // rs2=0b00100
	"sext.h": true, // rs2=0b00101
	"rev8":   true, // rs2=0b11000
}

// shiftImmediateInstructions are instructions where rs2 field is actually
// a shift amount (shamt), not a fixed value. They should be matched as
// fallback when no Zbb unary instruction matches.
var shiftImmediateInstructions = map[string]bool{
	"binvi": true,
	"bseti": true,
	"bclri": true,
	"bexti": true,
	"rori":  true,
}

func init() {
	reverseOpcodeTable = make(map[OpcodeKey]*OpcodeEntry)

	for mnemonic, info := range arch.Opcodes {
		key := buildKeyForMnemonic(mnemonic, info)
		reverseOpcodeTable[key] = &OpcodeEntry{
			Mnemonic: mnemonic,
			Info:     info,
		}
	}
}

// buildKeyForMnemonic creates a lookup key from opcode info, with special
// handling for certain instruction classes
func buildKeyForMnemonic(mnemonic string, info arch.InstrOpcode) OpcodeKey {
	key := OpcodeKey{
		Opcode: info.Opcode,
		Funct3: info.Funct3,
	}

	// Include funct7 for opcodes that need it
	if useFunct7Opcodes[info.Opcode] {
		key.Funct7 = info.Funct7
	}

	// Zbb unary instructions always include Rs2 (even when 0)
	// because they have fixed rs2 values that distinguish them
	// from shift-immediate instructions
	if zbbUnaryInstructions[mnemonic] {
		key.Rs2 = info.Rs2
		return key
	}

	// Shift-immediate instructions (binvi, bseti, etc.) do NOT include
	// Rs2 in the key because it's a variable shift amount
	if shiftImmediateInstructions[mnemonic] {
		return key
	}

	// For other instructions, include rs2 if non-zero
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
