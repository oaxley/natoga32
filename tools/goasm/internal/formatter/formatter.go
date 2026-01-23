// Instruction formatter for the NATOGA32 disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package formatter

import (
	"fmt"
	"strings"

	"github.com/natoga32/goasm/internal/arch"
	"github.com/natoga32/goasm/internal/decoder"
)

// Options controls the output format of the disassembler
type Options struct {
	UseABINames bool // Use ABI register names (ra, sp, a0) instead of x0-x31
	ShowAddress bool // Prefix output with address
	ShowHex     bool // Show hex encoding as comment
	ShowPseudo  bool // Show pseudo-instructions when possible
}

// DefaultOptions returns the default formatting options
func DefaultOptions() Options {
	return Options{
		UseABINames: true,
		ShowAddress: true,
		ShowHex:     true,
	}
}

// Formatter formats decoded instructions as assembly text
type Formatter struct {
	opts   Options
	pseudo *decoder.PseudoMatcher
}

// New creates a new Formatter with the given options
func New(opts Options) *Formatter {
	return &Formatter{
		opts:   opts,
		pseudo: decoder.NewPseudoMatcher(),
	}
}

// Format formats a decoded instruction at the given address
func (f *Formatter) Format(addr uint32, instr *decoder.DecodedInstruction) string {
	var parts []string

	// Address prefix
	if f.opts.ShowAddress {
		parts = append(parts, fmt.Sprintf("0x%08X:", addr))
	}

	// Hex encoding
	if f.opts.ShowHex {
		parts = append(parts, fmt.Sprintf("%08X", instr.RawWord))
	}

	// Instruction text
	instrText := f.formatInstruction(addr, instr)
	parts = append(parts, instrText)

	return strings.Join(parts, "  ")
}

// FormatInstruction formats just the instruction mnemonic and operands
func (f *Formatter) FormatInstruction(addr uint32, instr *decoder.DecodedInstruction) string {
	return f.formatInstruction(addr, instr)
}

// formatInstruction formats the instruction mnemonic and operands
func (f *Formatter) formatInstruction(addr uint32, instr *decoder.DecodedInstruction) string {
	if instr.IsUnknown() {
		return fmt.Sprintf(".word 0x%08X", instr.RawWord)
	}

	// Try pseudo-instruction matching if enabled
	if f.opts.ShowPseudo {
		if mnemonic, operands, matched := f.pseudo.Match(instr, addr, f.opts.UseABINames); matched {
			if operands == "" {
				return mnemonic
			}
			return fmt.Sprintf("%s %s", mnemonic, operands)
		}
	}

	switch instr.Type {
	case arch.TypeR:
		return f.formatTypeR(instr)
	case arch.TypeR2:
		return f.formatTypeR2(instr)
	case arch.TypeI:
		return f.formatTypeI(instr)
	case arch.TypeI2:
		return f.formatTypeI2(instr)
	case arch.TypeS:
		return f.formatTypeS(instr)
	case arch.TypeB:
		return f.formatTypeB(addr, instr)
	case arch.TypeU:
		return f.formatTypeU(instr)
	case arch.TypeJ:
		return f.formatTypeJ(addr, instr)
	default:
		return fmt.Sprintf(".word 0x%08X", instr.RawWord)
	}
}

// formatTypeR formats R-type instructions: mnemonic rd, rs1, rs2
func (f *Formatter) formatTypeR(instr *decoder.DecodedInstruction) string {
	rd := f.regName(instr.Rd)
	rs1 := f.regName(instr.Rs1)
	rs2 := f.regName(instr.Rs2)
	return fmt.Sprintf("%s %s, %s, %s", instr.Mnemonic, rd, rs1, rs2)
}

// formatTypeR2 formats R2-type instructions: mnemonic rs1, rs2
func (f *Formatter) formatTypeR2(instr *decoder.DecodedInstruction) string {
	rs1 := f.regName(instr.Rs1)
	rs2 := f.regName(instr.Rs2)
	return fmt.Sprintf("%s %s, %s", instr.Mnemonic, rs1, rs2)
}

// formatTypeI formats I-type instructions
func (f *Formatter) formatTypeI(instr *decoder.DecodedInstruction) string {
	rd := f.regName(instr.Rd)
	rs1 := f.regName(instr.Rs1)

	// Load instructions use offset(base) format
	if isLoadInstruction(instr.Mnemonic) {
		return fmt.Sprintf("%s %s, %d(%s)", instr.Mnemonic, rd, instr.Immediate, rs1)
	}

	// CSR instructions have special formatting
	if isCSRInstruction(instr.Mnemonic) {
		return f.formatCSR(instr)
	}

	// System instructions with no operands
	if instr.Mnemonic == "ecall" || instr.Mnemonic == "mret" || instr.Mnemonic == "wfi" {
		return instr.Mnemonic
	}

	// JALR: jalr rd, rs1, offset (or jalr rd, offset(rs1) in some conventions)
	if instr.Mnemonic == "jalr" {
		return fmt.Sprintf("%s %s, %s, %d", instr.Mnemonic, rd, rs1, instr.Immediate)
	}

	// Regular I-type: mnemonic rd, rs1, imm
	return fmt.Sprintf("%s %s, %s, %d", instr.Mnemonic, rd, rs1, instr.Immediate)
}

// formatTypeI2 formats I2-type (shift immediate) instructions: mnemonic rd, rs1, shamt
func (f *Formatter) formatTypeI2(instr *decoder.DecodedInstruction) string {
	rd := f.regName(instr.Rd)
	rs1 := f.regName(instr.Rs1)
	return fmt.Sprintf("%s %s, %s, %d", instr.Mnemonic, rd, rs1, instr.Immediate)
}

// formatTypeS formats S-type (store) instructions: mnemonic rs2, offset(rs1)
func (f *Formatter) formatTypeS(instr *decoder.DecodedInstruction) string {
	rs1 := f.regName(instr.Rs1)
	rs2 := f.regName(instr.Rs2)
	return fmt.Sprintf("%s %s, %d(%s)", instr.Mnemonic, rs2, instr.Immediate, rs1)
}

// formatTypeB formats B-type (branch) instructions: mnemonic rs1, rs2, target
func (f *Formatter) formatTypeB(addr uint32, instr *decoder.DecodedInstruction) string {
	rs1 := f.regName(instr.Rs1)
	rs2 := f.regName(instr.Rs2)
	target := calculateTarget(addr, instr.Immediate)
	return fmt.Sprintf("%s %s, %s, 0x%X", instr.Mnemonic, rs1, rs2, target)
}

// formatTypeU formats U-type instructions: mnemonic rd, imm
func (f *Formatter) formatTypeU(instr *decoder.DecodedInstruction) string {
	rd := f.regName(instr.Rd)
	// U-type immediate is the upper 20 bits value
	return fmt.Sprintf("%s %s, 0x%X", instr.Mnemonic, rd, uint32(instr.Immediate)&0xFFFFF)
}

// formatTypeJ formats J-type (jump) instructions: mnemonic rd, target
func (f *Formatter) formatTypeJ(addr uint32, instr *decoder.DecodedInstruction) string {
	rd := f.regName(instr.Rd)
	target := calculateTarget(addr, instr.Immediate)
	return fmt.Sprintf("%s %s, 0x%X", instr.Mnemonic, rd, target)
}

// formatCSR formats CSR instructions
func (f *Formatter) formatCSR(instr *decoder.DecodedInstruction) string {
	rd := f.regName(instr.Rd)
	rs1 := f.regName(instr.Rs1)
	csr := uint32(instr.Immediate) & 0xFFF

	// CSR immediate variants use rs1 field as immediate
	if strings.HasSuffix(instr.Mnemonic, "i") {
		uimm := instr.Rs1
		return fmt.Sprintf("%s %s, 0x%X, %d", instr.Mnemonic, rd, csr, uimm)
	}

	return fmt.Sprintf("%s %s, 0x%X, %s", instr.Mnemonic, rd, csr, rs1)
}

// regName returns the register name based on formatter options
func (f *Formatter) regName(reg int) string {
	return arch.RegisterName(reg, f.opts.UseABINames)
}

// calculateTarget calculates the absolute target address from PC and offset
func calculateTarget(addr uint32, offset int32) uint32 {
	return uint32(int32(addr) + offset)
}

// isLoadInstruction returns true if the mnemonic is a load instruction
func isLoadInstruction(mnemonic string) bool {
	switch mnemonic {
	case "lb", "lh", "lw", "lbu", "lhu":
		return true
	default:
		return false
	}
}

// isCSRInstruction returns true if the mnemonic is a CSR instruction
func isCSRInstruction(mnemonic string) bool {
	switch mnemonic {
	case "csrrw", "csrrs", "csrrc", "csrrwi", "csrrsi", "csrrci":
		return true
	default:
		return false
	}
}

// FormatBranchTarget formats a branch/jump target as a string
func FormatBranchTarget(addr uint32, offset int32) string {
	target := calculateTarget(addr, offset)
	return fmt.Sprintf("0x%08X", target)
}
