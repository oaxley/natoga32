// Disassembly logic for the NATOGA32 disassembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/natoga32/goasm/internal/decoder"
	"github.com/natoga32/goasm/internal/formatter"
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
	fmt.Fprintf(out, "; .text: 0x%08X - 0x%08X (%d bytes)\n",
		reader.TextBase, reader.TextBase+reader.TextSize, reader.TextSize)
	if reader.HasDataSection() {
		fmt.Fprintf(out, "; .data: 0x%08X - 0x%08X (%d bytes)\n",
			reader.DataBase, reader.DataBase+reader.DataSize, reader.DataSize)
	}
	fmt.Fprintln(out)

	// Disassemble .text section
	fmt.Fprintln(out, ".text")
	textData := reader.TextSection()
	addr := reader.TextBase

	for i := 0; i+4 <= len(textData); i += 4 {
		chunk := textData[i : i+4]
		instr := decoder.DecodeBytes(chunk)
		line := fmtr.Format(addr, instr)
		fmt.Fprintln(out, line)
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

	// Dump .data section if present
	if reader.HasDataSection() {
		fmt.Fprintln(out)
		fmt.Fprintln(out, ".data")
		if err := dumpDataSection(out, reader); err != nil {
			return err
		}
	}

	return nil
}

// dumpDataSection outputs the .data section as hex bytes
func dumpDataSection(out io.Writer, reader *BinaryReader) error {
	dataSection := reader.DataSection()
	addr := reader.DataBase

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
