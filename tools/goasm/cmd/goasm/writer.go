package main

import (
	"fmt"
	"os"

	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

// maxRomSize is the size of the ROM address window (2 MiB)
const maxRomSize = 2 * 1024 * 1024

// writeOutput writes the assembled output to a file
func writeOutput(sections *section.Manager, filename, format string) error {
	// validate the format and the ROM image before creating the output file
	switch format {
	case "rscv", "bin", "rom":
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
	if format == "rom" {
		romSec := sections.Get(".rom")
		if romSec == nil || len(romSec.Data) == 0 {
			return fmt.Errorf("no .rom section: -f rom needs code or data in a .rom section")
		}
		if len(romSec.Data) > maxRomSize {
			return fmt.Errorf("ROM image exceeds 2 MiB (%d bytes)", len(romSec.Data))
		}
	}
	var rscv []byte
	if format == "rscv" {
		var err error
		if rscv, err = buildRscv(sections); err != nil {
			return err
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// .rodata is not part of the bin output yet: never drop it silently
	if format == "bin" {
		if ro := sections.Get(".rodata"); ro != nil && ro.Size > 0 {
			fmt.Fprintf(os.Stderr, "warning: .rodata section (%d bytes) is not written to %s output yet\n", ro.Size, format)
		}
	}

	switch format {
	case "rscv":
		// Cartridge RSCV chunk - every loadable section
		_, err = file.Write(rscv)

	case "bin":
		// Raw binary - just write .text section
		textSec := sections.Get(".text")
		if textSec != nil {
			_, err = file.Write(textSec.Data)
		}

	case "rom":
		// ROM image - the .rom section bytes, starting at address 0
		_, err = file.Write(sections.Get(".rom").Data)
	}

	return err
}

// writeSymbols writes buildDsym's chunk to filename
func writeSymbols(symbols *symbol.Table, filename string) error {
	return os.WriteFile(filename, buildDsym(symbols), 0o644)
}
