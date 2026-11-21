// IHDR and file header flags for the NATOGA32 cartridge packer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package packer

import (
	"encoding/binary"
	"fmt"
	"time"
)

// ihdrSize is the size of the IHDR chunk data
const ihdrSize = 64

// IHDR field offsets
const (
	ihdrOffVersion   = 0x00
	ihdrOffMinRom    = 0x02
	ihdrOffRegions   = 0x04
	ihdrOffLanguages = 0x06
	ihdrOffDate      = 0x08
	ihdrOffEntry     = 0x0C
	ihdrOffName      = 0x18
)

// flagModeRequired is the file header flag set when a graphics mode is required
const flagModeRequired = 0x0008

// buildIhdr returns the 64-byte IHDR data; a date error names the manifest
func buildIhdr(m *Manifest, entry uint32, now time.Time) ([]byte, error) {
	date, err := m.dateSeconds(now)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", m.file, err)
	}
	ihdr := make([]byte, ihdrSize)
	binary.LittleEndian.PutUint16(ihdr[ihdrOffVersion:], m.versionBCD())
	ihdr[ihdrOffMinRom], ihdr[ihdrOffMinRom+1] = m.minRom()
	ihdr[ihdrOffRegions] = m.regionBits()
	binary.LittleEndian.PutUint16(ihdr[ihdrOffLanguages:], m.languageBits())
	binary.LittleEndian.PutUint32(ihdr[ihdrOffDate:], date)
	binary.LittleEndian.PutUint32(ihdr[ihdrOffEntry:], entry)
	// DSP entry (+0x10) and save RAM (+0x14) stay 0
	copy(ihdr[ihdrOffName:ihdrOffName+32], m.Name)
	return ihdr, nil
}

// headerFlags returns the file header flags (graphics mode)
func headerFlags(m *Manifest) uint16 {
	if m.GraphicsMode == nil {
		return 0
	}
	return flagModeRequired | uint16(*m.GraphicsMode)
}
