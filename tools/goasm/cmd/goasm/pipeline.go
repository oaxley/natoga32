package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/natoga32/goasm/internal/lexer"
	"github.com/natoga32/goasm/internal/parser"
	"github.com/natoga32/goasm/internal/preprocess"
	"github.com/natoga32/goasm/internal/pseudo"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

func assemble(inputFile string) error {
	fmt.Printf("Assembling '%s' to '%s' (format: %s)...\n", inputFile, outputFile, outputFormat)

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
	expander := pseudo.New()
	program = expander.ExpandProgram(program)

	if debugLevel == 4 {
		fmt.Println("=== EXPANDED AST ===")
		debugProgram(program)
		return nil
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
	)

	// Phase 5: Semantic Analysis - Pass 1 (calculate addresses)
	fmt.Println("Phase 5: Semantic Analyzer - pass 1")
	if err := semanticPass1(program, symbols, sections); err != nil {
		return fmt.Errorf("semantic pass 1 error: %v", err)
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

	// Phase 6: Semantic Analysis - Pass 2 (resolve references)
	fmt.Println("Phase 6: Semantic Analyzer - pass 2")
	// Nothing to do here for now - symbols are resolved during encoding

	if debugLevel == 6 {
		fmt.Println("=== SYMBOLS (resolved) ===")
		for _, sym := range symbols.AllSymbols() {
			fmt.Printf("  %s: 0x%08X (%s)\n", sym.Name, sym.Value, sym.Type)
		}
		return nil
	}

	// Phase 7: Semantic Analysis - Pass 3 (encode instructions)
	fmt.Println("Phase 7: Semantic Analyzer - pass 3")
	sections = section.NewManager() // Reset sections for encoding
	sections.SetBaseAddresses(
		MMAP_TEXT_SEGMENT,
		MMAP_DATA_SEGMENT,
		MMAP_BSS_SEGMENT,
		MMAP_ROM_SEGMENT,
	)

	if err := semanticPass3(program, symbols, sections); err != nil {
		return fmt.Errorf("semantic pass 3 error: %v", err)
	}

	if debugLevel == 7 {
		fmt.Println("=== ENCODED SECTIONS ===")
		for _, sec := range sections.All() {
			if sec.Size > 0 {
				fmt.Printf("Section %s (0x%08X, %d bytes):\n", sec.Name, sec.BaseAddr, sec.Size)
				printHexDump(sec.Data, sec.BaseAddr)
			}
		}
		return nil
	}

	// Phase 8: Output
	fmt.Println("Phase 8: Output")
	if err := writeOutput(sections, outputFile, outputFormat); err != nil {
		return fmt.Errorf("output error: %v", err)
	}

	fmt.Printf("Output written to '%s'\n", outputFile)
	return nil
}
