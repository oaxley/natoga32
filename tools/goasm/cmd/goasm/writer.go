package main

import (
	"fmt"
	"os"

	"github.com/natoga32/goasm/internal/section"
)

// maxRomSize is the size of the ROM address window (2 MiB)
const maxRomSize = 2 * 1024 * 1024

// writeOutput writes the assembled output to a file
func writeOutput(sections *section.Manager, filename, format string) error {
	// validate the format and the ROM image before creating the output file
	switch format {
	case "bin", "rom", "on32":
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

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// .rodata is not part of the on32/bin output yet: never drop it silently
	if format == "on32" || format == "bin" {
		if ro := sections.Get(".rodata"); ro != nil && ro.Size > 0 {
			fmt.Fprintf(os.Stderr, "warning: .rodata section (%d bytes) is not written to %s output yet\n", ro.Size, format)
		}
	}

	switch format {
	case "bin":
		// Raw binary - just write .text section
		textSec := sections.Get(".text")
		if textSec != nil {
			_, err = file.Write(textSec.Data)
		}

	case "rom":
		// ROM image - the .rom section bytes, starting at address 0
		_, err = file.Write(sections.Get(".rom").Data)

	case "on32":
		// ON32 format with header
		// For now, just write raw sections with a simple header
		textSec := sections.Get(".text")
		dataSec := sections.Get(".data")

		// Write a simple header (magic + sizes)
		header := make([]byte, 16)
		header[0] = 'O'
		header[1] = 'N'
		header[2] = '3'
		header[3] = '2'

		textSize := uint32(0)
		dataSize := uint32(0)
		if textSec != nil {
			textSize = uint32(textSec.Size)
		}
		if dataSec != nil {
			dataSize = uint32(dataSec.Size)
		}

		// Text size (little-endian)
		header[4] = byte(textSize)
		header[5] = byte(textSize >> 8)
		header[6] = byte(textSize >> 16)
		header[7] = byte(textSize >> 24)

		// Data size (little-endian)
		header[8] = byte(dataSize)
		header[9] = byte(dataSize >> 8)
		header[10] = byte(dataSize >> 16)
		header[11] = byte(dataSize >> 24)

		_, err = file.Write(header)
		if err != nil {
			return err
		}

		// Write sections
		if textSec != nil {
			_, err = file.Write(textSec.Data)
			if err != nil {
				return err
			}
		}
		if dataSec != nil {
			_, err = file.Write(dataSec.Data)
			if err != nil {
				return err
			}
		}
	}

	return err
}
