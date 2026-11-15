package main

import (
	"fmt"
	"os"

	"github.com/natoga32/goasm/internal/section"
)

// writeOutput writes the assembled output to a file
func writeOutput(sections *section.Manager, filename, format string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	switch format {
	case "bin":
		// Raw binary - just write .text section
		textSec := sections.Get(".text")
		if textSec != nil {
			_, err = file.Write(textSec.Data)
		}

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

	default:
		return fmt.Errorf("unknown output format: %s", format)
	}

	return err
}
