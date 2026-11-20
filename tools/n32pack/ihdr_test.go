// Tests of the IHDR and the file header flags
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestBuildIhdr(t *testing.T) {
	m := &Manifest{
		Name:      "Test Game",
		Version:   "1.23",
		MinRom:    "2.5",
		Regions:   []string{"americas", "asia"},
		Languages: []string{"english"},
		Date:      "2026-01-02T00:00:01Z",
	}
	got, err := buildIhdr(m, 0x01A00010, time.Now())
	if err != nil {
		t.Fatalf("buildIhdr: %v", err)
	}
	want := make([]byte, 64)
	copy(want[0x00:], []byte{0x23, 0x01})             // version 1.23 BCD
	copy(want[0x02:], []byte{2, 5})                   // min ROM 2.5
	want[0x04] = 0x05                                 // americas | asia
	copy(want[0x06:], []byte{0x01, 0x00})             // english
	copy(want[0x08:], []byte{0x81, 0x51, 0x01, 0x00}) // 86401 seconds
	copy(want[0x0C:], []byte{0x10, 0x00, 0xA0, 0x01}) // entry
	copy(want[0x18:], "Test Game")
	if !bytes.Equal(got, want) {
		t.Errorf("IHDR:\n got % X\nwant % X", got, want)
	}
}

func TestBuildIhdrName32(t *testing.T) {
	name := strings.Repeat("N", 32)
	m := &Manifest{Name: name, Version: "1.00", MinRom: "0.1", Regions: []string{"europe"}, Date: "2026-01-01T00:00:00Z"}
	got, err := buildIhdr(m, 0x01A00000, time.Now())
	if err != nil {
		t.Fatalf("buildIhdr: %v", err)
	}
	if len(got) != 64 {
		t.Fatalf("IHDR is %d bytes", len(got))
	}
	if string(got[0x18:0x38]) != name {
		t.Errorf("name % X", got[0x18:0x38])
	}
	if !bytes.Equal(got[0x38:], make([]byte, 8)) {
		t.Errorf("reserved % X", got[0x38:])
	}
}

func TestBuildIhdrNow(t *testing.T) {
	m := &Manifest{Name: "X", Version: "1.00", MinRom: "0.1", Regions: []string{"europe"}, file: "game.json"}
	got, err := buildIhdr(m, 0x01A00000, time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("buildIhdr: %v", err)
	}
	if !bytes.Equal(got[0x08:0x0C], []byte{60, 0, 0, 0}) {
		t.Errorf("date % X", got[0x08:0x0C])
	}
	if _, err := buildIhdr(m, 0x01A00000, time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)); err == nil ||
		err.Error() != "game.json: date is before 2026-01-01" {
		t.Errorf("buildIhdr before the epoch: %v", err)
	}
}

func TestHeaderFlags(t *testing.T) {
	mode := func(v int) *int { return &v }
	for _, tc := range []struct {
		mode *int
		want uint16
	}{{nil, 0}, {mode(0), 0x0008}, {mode(3), 0x000B}} {
		if got := headerFlags(&Manifest{GraphicsMode: tc.mode}); got != tc.want {
			t.Errorf("headerFlags(%v) = 0x%04X, want 0x%04X", tc.mode, got, tc.want)
		}
	}
}
