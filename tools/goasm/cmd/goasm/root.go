// goasm - NATOGA32 RISC-V Assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

var (
	outputFile   string
	outputFormat string
	debugLevel   int
	baseDir      string // Base directory for resolving relative paths
)

var rootCmd = &cobra.Command{
	Use:   "goasm [input file]",
	Short: "NATOGA32 RISC-V Assembler",
	Long: `goasm is a RISC-V assembler for the NATOGA32 virtual console.

It assembles RISC-V assembly files into binary format for ROM or
cartridge loading.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		if err := assemble(inputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "a.out", "Output file path")
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "on32", "Output format: on32, bin")
	rootCmd.Flags().IntVarP(&debugLevel, "debug", "d", 0, "Debug level (1-8 to stop after specific phase)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// semanticPass1 performs the first semantic analysis pass:
// - Calculate addresses for labels
// - Build symbol table
// - Track section sizes
func semanticPass1(program *ast.Program, symbols *symbol.Table, sections *section.Manager) error {
	// Set base addresses according to NATOGA32 memory map
	sections.SetBaseAddresses(
		MMAP_TEXT_SEGMENT,
		MMAP_DATA_SEGMENT,
		MMAP_DATA_SEGMENT, // same for data for now
		MMAP_BSS_SEGMENT,
	)

	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.Label:
			addr := sections.CurrentAddress()
			if err := symbols.DefineLabel(s.Name, addr, sections.CurrentName(), 0); err != nil {
				return err
			}

		case *ast.Instruction:
			// Each instruction is 4 bytes
			sections.Current().Size += 4

		case *ast.Directive:
			if err := handleDirective(s, symbols, sections); err != nil {
				return err
			}
		}
	}

	return nil
}

// semanticPass3 encodes instructions to binary
func semanticPass3(program *ast.Program, symbols *symbol.Table, sections *section.Manager) error {
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.Instruction:
			if err := encodeInstruction(s, symbols, sections); err != nil {
				return err
			}

		case *ast.Directive:
			if err := emitDirective(s, symbols, sections); err != nil {
				return err
			}
		}
	}
	return nil
}

// printHexDump prints a hex dump of data
func printHexDump(data []byte, baseAddr int64) {
	for i := 0; i < len(data); i += 16 {
		fmt.Printf("  %08X: ", baseAddr+int64(i))
		for j := 0; j < 16 && i+j < len(data); j++ {
			fmt.Printf("%02X ", data[i+j])
		}
		// Padding for short lines
		for j := len(data) - i; j < 16; j++ {
			fmt.Print("   ")
		}
		fmt.Print(" |")
		for j := 0; j < 16 && i+j < len(data); j++ {
			ch := data[i+j]
			if ch >= 32 && ch < 127 {
				fmt.Printf("%c", ch)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}
}
