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
	optimizeFlag bool   // -O/--optimize: Enable peephole optimizations

	// Disassembler flags
	disasmMode     bool   // -disasm: Enable disassembly mode
	disasmBaseAddr uint32 // -base: Base address for raw binary (default: MMAP_TEXT_SEGMENT)
	disasmPseudo   bool   // -pseudo: Enable pseudo-instruction recognition (Step 4)
	disasmShowHex  bool   // -hex: Show hex encoding (default: true)
	disasmNoABI    bool   // -no-abi: Use x0-x31 instead of ABI names
)

var rootCmd = &cobra.Command{
	Use:   "goasm [input file]",
	Short: "NATOGA32 RISC-V Assembler/Disassembler",
	Long: `goasm is a RISC-V assembler and disassembler for the NATOGA32 virtual console.

It assembles RISC-V assembly files into binary format for ROM or
cartridge loading. Use -disasm flag to disassemble binary files.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]

		// Check for disassembly mode
		if disasmMode {
			return runDisasm(inputFile)
		}

		// Default: assembly mode
		return assemble(inputFile)
	},
}

func init() {
	// Assembly flags
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "a.out", "Output file path")
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "on32", "Output format: on32, bin")
	rootCmd.Flags().IntVarP(&debugLevel, "debug", "d", 0, "Debug level (1-8 to stop after specific phase)")
	rootCmd.Flags().BoolVarP(&optimizeFlag, "optimize", "O", false, "Enable peephole optimizations")

	// Disassembly flags
	rootCmd.Flags().BoolVar(&disasmMode, "disasm", false, "Disassemble input file")
	rootCmd.Flags().Uint32Var(&disasmBaseAddr, "base", MMAP_TEXT_SEGMENT, "Base address for disassembly (raw binary only)")
	rootCmd.Flags().BoolVar(&disasmPseudo, "pseudo", false, "Show pseudo-instructions (nop, li, mv, ret, etc.)")
	rootCmd.Flags().BoolVar(&disasmShowHex, "hex", true, "Show hex encoding in disassembly output")
	rootCmd.Flags().BoolVar(&disasmNoABI, "no-abi", false, "Use x0-x31 register names instead of ABI names")
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
