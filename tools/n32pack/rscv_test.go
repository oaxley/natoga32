// Tests of the RSCV section checks and the entry point
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"encoding/binary"
	"strings"
	"testing"
)

// section builds the bytes of one RSCV section: the 12-byte header, then the data padded to 4
func section(kind uint8, addr, size uint32, data []byte) []byte {
	b := []byte{kind, 0, 0, 0}
	b = binary.LittleEndian.AppendUint32(b, addr)
	b = binary.LittleEndian.AppendUint32(b, size)
	b = append(b, data...)
	return append(b, make([]byte, align4(uint64(len(data)))-uint64(len(data)))...)
}

func concat(parts ...[]byte) []byte {
	var b []byte
	for _, p := range parts {
		b = append(b, p...)
	}
	return b
}

func TestParseRscvValid(t *testing.T) {
	data := concat(
		section(0x01, 0x01A00000, 6, []byte{1, 2, 3, 4, 5, 6}),
		section(0x02, 0x01400000, 4, []byte{7, 8, 9, 10}),
		section(0x04, 0x01000000, 64, nil),
	)
	got, err := parseRscv(data)
	if err != nil {
		t.Fatalf("parseRscv: %v", err)
	}
	want := []Section{{1, 0x01A00000, 6}, {2, 0x01400000, 4}, {4, 0x01000000, 64}}
	if len(got) != len(want) {
		t.Fatalf("sections = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("section %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseRscvAdjacent(t *testing.T) {
	data := concat(
		section(0x01, 0x01A00000, 4, []byte{1, 2, 3, 4}),
		section(0x02, 0x01400000, 4, []byte{1, 2, 3, 4}),
		section(0x02, 0x01400004, 4, []byte{1, 2, 3, 4}),
	)
	if _, err := parseRscv(data); err != nil {
		t.Fatalf("parseRscv: %v", err)
	}
}

func TestParseRscvZeroSize(t *testing.T) {
	data := concat(
		section(0x01, 0x01A00000, 4, []byte{1, 2, 3, 4}),
		section(0x02, 0x01400000, 0, nil),
		section(0x02, 0x01400000, 4, []byte{1, 2, 3, 4}),
	)
	if _, err := parseRscv(data); err != nil {
		t.Fatalf("parseRscv: %v", err)
	}
}

func TestParseRscvErrors(t *testing.T) {
	text := section(0x01, 0x01A00000, 4, []byte{1, 2, 3, 4})
	reserved := section(0x02, 0x01400000, 4, []byte{1, 2, 3, 4})
	reserved[2] = 1
	flags := section(0x02, 0x01400000, 4, []byte{1, 2, 3, 4})
	flags[1] = 1
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"truncated header", concat(text, []byte{0x02, 0, 0, 0, 0, 0}), "RSCV: section 1 runs past the end of the chunk"},
		{"data past end", concat(text, section(0x02, 0x01400000, 8, []byte{1, 2, 3, 4})), "RSCV: section 1 runs past the end of the chunk"},
		{"unknown kind", concat(text, section(0x05, 0x01400000, 0, nil)), "RSCV: section 1: unknown kind 0x05"},
		{"reserved byte", concat(text, reserved), "RSCV: section 1: reserved fields must be 0"},
		{"flags byte", concat(text, flags), "RSCV: section 1: reserved fields must be 0"},
		{"text past end, no wrap", section(0x01, 0x01FFFFFC, 8, make([]byte, 8)),
			"RSCV: section 0 (.text at 0x01FFFFFC, 8 bytes) is outside its region 0x01A00000-0x01FFFFFF"},
		{"text data size huge", section(0x01, 0x01A00000, 0xFFFFFFF0, nil), "RSCV: section 0 runs past the end of the chunk"},
		{"bss wraps", concat(text, section(0x04, 0x01100000, 0xFFFFFFF0, nil)),
			"RSCV: section 1 (.bss at 0x01100000, 4294967280 bytes) is outside its region 0x01000000-0x011FFFFF"},
		{"data wrong region", concat(text, section(0x02, 0x01A00000, 4, []byte{1, 2, 3, 4})),
			"RSCV: section 1 (.data at 0x01A00000, 4 bytes) is outside its region 0x01400000-0x019FFFFF"},
		{"bss below region", concat(text, section(0x04, 0x00FFFFFC, 4, nil)),
			"RSCV: section 1 (.bss at 0x00FFFFFC, 4 bytes) is outside its region 0x01000000-0x011FFFFF"},
		{"overlap", concat(text,
			section(0x02, 0x01400000, 8, make([]byte, 8)),
			section(0x02, 0x01400004, 4, []byte{1, 2, 3, 4})), "RSCV: section 2 overlaps section 1"},
		{"no text", section(0x02, 0x01400000, 4, []byte{1, 2, 3, 4}), "RSCV: no .text section"},
		{"empty", nil, "RSCV: no .text section"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parseRscv(c.data)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
		})
	}
}

func TestEntryPoint(t *testing.T) {
	sections := []Section{
		{Kind: 0x02, Addr: 0x01400000, Size: 16},
		{Kind: 0x01, Addr: 0x01A00100, Size: 16},
		{Kind: 0x01, Addr: 0x01B00000, Size: 16},
	}
	symbols := map[string]uint32{"main": 0x01A00108, "odd": 0x01A00102, "var": 0x01400000}
	cases := []struct {
		name    string
		entry   string
		symbols map[string]uint32
		want    uint32
		err     string
	}{
		{"default", "", nil, 0x01A00100, ""},
		{"hex", "0x01B00004", nil, 0x01B00004, ""},
		{"label", "main", symbols, 0x01A00108, ""},
		{"missing dsym", "main", nil, 0, `entry "main": no dsym file given`},
		{"unknown label", "start", symbols, 0, `entry "start": unknown symbol`},
		{"bad hex", "0xZZ", nil, 0, `entry "0xZZ":`},
		{"unaligned", "odd", symbols, 0, "entry 0x01A00102 is not 4-byte aligned"},
		{"outside text", "var", symbols, 0, "entry 0x01400000 is not in a .text section"},
		{"past text end", "0x01A00110", nil, 0, "entry 0x01A00110 is not in a .text section"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := entryPoint(c.entry, sections, c.symbols)
			if c.err != "" {
				if err == nil || !strings.HasPrefix(err.Error(), c.err) {
					t.Fatalf("err = %v, want prefix %q", err, c.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("entryPoint: %v", err)
			}
			if got != c.want {
				t.Fatalf("entry = 0x%08X, want 0x%08X", got, c.want)
			}
		})
	}
}
