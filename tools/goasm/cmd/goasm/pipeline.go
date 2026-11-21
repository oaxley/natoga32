package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/lexer"
	"github.com/natoga32/goasm/internal/optimizer"
	"github.com/natoga32/goasm/internal/parser"
	"github.com/natoga32/goasm/internal/preprocess"
	"github.com/natoga32/goasm/internal/pseudo"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

func assemble(inputFile string) error {
	if debugLevel < 0 || debugLevel > maxDebugLevel {
		return fmt.Errorf("--debug must be 0 to %d, got %d", maxDebugLevel, debugLevel)
	}

	fmt.Printf("Assembling '%s' to '%s' (format: %s)...\n", inputFile, outputFile, outputFormat)

	if outputFormat != "rscv" && outputFormat != "rom" {
		return fmt.Errorf("unknown output format: %s (use rscv or rom)", outputFormat)
	}

	// the dSYM chunk goes with a cartridge: refuse it before anything is written
	if symFile != "" && outputFormat != "rscv" {
		return fmt.Errorf("--sym needs -f rscv")
	}

	// Read input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer file.Close()

	// Phase 1: Lexer
	fmt.Println("Phase 1: Lexer")
	lex, err := lexer.NewFromReaderWithFile(file, inputFile)
	if err != nil {
		return fmt.Errorf("lexer error: %v", err)
	}
	tokens := lex.Tokenize()

	if debugLevel == 1 {
		fmt.Println("=== TOKENS ===")
		for i, tok := range tokens {
			fmt.Printf("%4d: %s\n", i, tok)
		}
		return nil
	}

	// Phase 2: Preprocessor
	fmt.Println("Phase 2: Preprocessor")
	baseDir = filepath.Dir(inputFile)
	pp := preprocess.New(baseDir)
	tokens, err = pp.Process(tokens)
	if err != nil {
		return fmt.Errorf("preprocessor error: %v", err)
	}

	if debugLevel == 2 {
		fmt.Println("=== PREPROCESSED TOKENS ===")
		for i, tok := range tokens {
			fmt.Printf("%4d: %s\n", i, tok)
		}
		return nil
	}

	// Phase 3: Parser
	fmt.Println("Phase 3: Parser")
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parser error: %v", err)
	}

	if debugLevel == 3 {
		fmt.Println("=== AST ===")
		debugProgram(program)
		return nil
	}

	// Phase 4: Pseudo-instruction expansion
	fmt.Println("Phase 4: Pseudo-instruction expansion")
	for _, stmt := range program.Statements {
		if instr, ok := stmt.(*ast.Instruction); ok {
			if err := pseudo.CheckOperands(instr); err != nil {
				return err
			}
		}
	}
	expander := pseudo.New()
	program = expander.ExpandProgram(program)

	if debugLevel == 4 {
		fmt.Println("=== EXPANDED AST ===")
		debugProgram(program)
		return nil
	}

	// Phase 4.5: Peephole Optimization (optional)
	if optimizeFlag {
		fmt.Println("Phase 4.5: Peephole Optimization")
		opt := optimizer.New()
		program = opt.OptimizeProgram(program)
	}

	// Initialize symbol table and section manager
	symbols := symbol.New()
	sections := section.NewManager()
	// Set base addresses according to NATOGA32 memory map
	sections.SetBaseAddresses(
		MMAP_TEXT_SEGMENT,
		MMAP_DATA_SEGMENT,
		MMAP_BSS_SEGMENT,
		MMAP_ROM_SEGMENT,
		MMAP_RODATA_SEGMENT,
	)

	// Phase 5: Semantic Analysis - Pass 1 (calculate addresses)
	fmt.Println("Phase 5: Semantic Analyzer - pass 1")
	if err := semanticPass1(program, symbols, sections); err != nil {
		return fmt.Errorf("semantic pass 1 error: %v", err)
	}
	if err := checkSectionSizes(sections); err != nil {
		return err
	}

	if debugLevel == 5 {
		fmt.Println("=== SYMBOLS ===")
		for _, sym := range symbols.AllSymbols() {
			fmt.Printf("  %s: 0x%08X (%s)\n", sym.Name, sym.Value, sym.Type)
		}
		fmt.Println("=== SECTIONS ===")
		for _, sec := range sections.All() {
			fmt.Printf("  %s: base=0x%08X size=%d\n", sec.Name, sec.BaseAddr, sec.Size)
		}
		return nil
	}

	// Phase 6: Semantic Analysis - Pass 2 (encode instructions, symbols resolved)
	fmt.Println("Phase 6: Semantic Analyzer - pass 2")
	sections = section.NewManager() // Reset sections for encoding
	sections.SetBaseAddresses(
		MMAP_TEXT_SEGMENT,
		MMAP_DATA_SEGMENT,
		MMAP_BSS_SEGMENT,
		MMAP_ROM_SEGMENT,
		MMAP_RODATA_SEGMENT,
	)

	if err := semanticPass2(program, symbols, sections); err != nil {
		return fmt.Errorf("semantic pass 2 error: %v", err)
	}

	if debugLevel == 6 {
		fmt.Println("=== ENCODED SECTIONS ===")
		for _, sec := range sections.All() {
			if sec.Size > 0 {
				fmt.Printf("Section %s (0x%08X, %d bytes):\n", sec.Name, sec.BaseAddr, sec.Size)
				printHexDump(sec.Data, sec.BaseAddr)
			}
		}
		return nil
	}

	// Phase 7: Output
	fmt.Println("Phase 7: Output")
	// build the dSYM chunk first: a bad label leaves neither output file behind
	var dsym []byte
	if symFile != "" {
		var err error
		if dsym, err = buildDsym(symbols); err != nil {
			return fmt.Errorf("symbol output error: %v", err)
		}
	}
	if err := writeOutput(sections, outputFile, outputFormat); err != nil {
		return fmt.Errorf("output error: %v", err)
	}
	if symFile != "" {
		if err := writeSymbols(dsym, symFile); err != nil {
			return fmt.Errorf("symbol output error: %v", err)
		}
		fmt.Printf("Symbols written to '%s'\n", symFile)
	}

	fmt.Printf("Output written to '%s'\n", outputFile)

	if debugLevel == 7 {
		printOutputSummary(sections, dsym)
	}
	return nil
}

// maxDebugLevel is the last phase -d can print (7: output)
const maxDebugLevel = 7

// printOutputSummary prints what phase 7 wrote: the format, the sections with
// data, the size of each file and the number of dSYM labels
func printOutputSummary(sections *section.Manager, dsym []byte) {
	fmt.Println("=== OUTPUT ===")
	fmt.Printf("  format: %s\n", outputFormat)
	for _, sec := range sections.All() {
		if sec.Size > 0 {
			fmt.Printf("  %s: base=0x%08X size=%d\n", sec.Name, sec.BaseAddr, sec.Size)
		}
	}
	if info, err := os.Stat(outputFile); err == nil {
		fmt.Printf("  %s: %d bytes\n", outputFile, info.Size())
	}
	if symFile != "" && len(dsym) >= 12 {
		// chunk: u32 length, "dSYM", then u32 label count
		fmt.Printf("  %s: %d labels\n", symFile, binary.LittleEndian.Uint32(dsym[8:12]))
	}
}
