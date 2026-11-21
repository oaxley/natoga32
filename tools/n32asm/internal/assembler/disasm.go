// Disassembly logic for the NATOGA32 disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package assembler

import (
	"fmt"
	"io"
	"os"

	"github.com/natoga32/n32asm/internal/decoder"
	"github.com/natoga32/n32asm/internal/formatter"
)

// runDisasm disassembles the input file and outputs assembly text
func runDisasm(inputFile string) error {
	// Read and parse the input file
	reader, err := NewBinaryReader(inputFile, disasmBaseAddr)
	if err != nil {
		return err
	}

	// Create formatter with options from flags
	opts := formatter.Options{
		UseABINames: !disasmNoABI,
		ShowAddress: true,
		ShowHex:     disasmShowHex,
		ShowPseudo:  disasmPseudo,
	}
	fmtr := formatter.New(opts)

	// Determine output destination
	var out io.Writer
	if outputFile != "" && outputFile != "a.out" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		out = file
	} else {
		out = os.Stdout
	}

	// Print file info header
	fmt.Fprintf(out, "; Disassembly of %s (%s format)\n", inputFile, reader.Format)
	for _, sec := range reader.Sections {
		fmt.Fprintf(out, "; %s: 0x%08X - 0x%08X (%d bytes)\n", sec.Name, sec.Addr, sec.Addr+sec.Size, sec.Size)
	}

	// Each section in file order
	for _, sec := range reader.Sections {
		fmt.Fprintln(out)
		fmt.Fprintln(out, sec.Name)
		switch sec.Kind {
		case rscvKindText:
			disasmTextSection(out, fmtr, sec)
		case rscvKindBss:
			fmt.Fprintf(out, "    .space %d\n", sec.Size)
		default:
			if err := dumpDataSection(out, sec); err != nil {
				return err
			}
		}
	}

	return nil
}

// disasmTextSection outputs the instructions of a .text section
func disasmTextSection(out io.Writer, fmtr *formatter.Formatter, sec ReadSection) {
	textData := sec.Data
	addr := sec.Addr

	for i := 0; i+4 <= len(textData); i += 4 {
		instr := decoder.DecodeBytes(textData[i : i+4])
		fmt.Fprintln(out, fmtr.Format(addr, instr))
		addr += 4
	}

	// Handle any remaining bytes that don't form a complete instruction
	remaining := len(textData) % 4
	if remaining > 0 {
		offset := len(textData) - remaining
		fmt.Fprintf(out, "; WARNING: %d trailing bytes at offset 0x%X\n", remaining, offset)
		for j := offset; j < len(textData); j++ {
			fmt.Fprintf(out, ".byte 0x%02X\n", textData[j])
		}
	}
}

// dumpDataSection outputs a .data or .rodata section as hex bytes
func dumpDataSection(out io.Writer, sec ReadSection) error {
	dataSection := sec.Data
	addr := sec.Addr

	// Output in 16-byte lines
	for i := 0; i < len(dataSection); i += 16 {
		end := i + 16
		if end > len(dataSection) {
			end = len(dataSection)
		}
		chunk := dataSection[i:end]

		// Address prefix
		fmt.Fprintf(out, "0x%08X:  ", addr)

		// Hex bytes
		for j := 0; j < 16; j++ {
			if j < len(chunk) {
				fmt.Fprintf(out, "%02X ", chunk[j])
			} else {
				fmt.Fprint(out, "   ")
			}
		}

		// ASCII representation
		fmt.Fprint(out, " |")
		for _, b := range chunk {
			if b >= 32 && b < 127 {
				fmt.Fprintf(out, "%c", b)
			} else {
				fmt.Fprint(out, ".")
			}
		}
		fmt.Fprintln(out, "|")

		addr += 16
	}

	return nil
}
