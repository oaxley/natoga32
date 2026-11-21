// Tests of the dSYM symbol checks
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package packer

import (
	"encoding/binary"
	"testing"
)

type dsymEntry struct {
	addr  uint32
	kind  uint8
	flags uint8
	name  string
}

// dsymData builds dSYM chunk data with the n32asm layout: u32 count, then per entry u32 address,
// u8 section kind, u8 flags, u16 name length, name, zero padding to 4
func dsymData(count uint32, entries ...dsymEntry) []byte {
	b := binary.LittleEndian.AppendUint32(nil, count)
	for _, e := range entries {
		b = binary.LittleEndian.AppendUint32(b, e.addr)
		b = append(b, e.kind, e.flags)
		b = binary.LittleEndian.AppendUint16(b, uint16(len(e.name)))
		b = append(b, e.name...)
		b = append(b, make([]byte, align4(uint64(len(e.name)))-uint64(len(e.name)))...)
	}
	return b
}

func TestParseDsym(t *testing.T) {
	data := dsymData(3,
		dsymEntry{0x01400000, 2, 0, "counter"},
		dsymEntry{0x01A00000, 1, 1, "main"},
		dsymEntry{0x01A00010, 1, 0, "loop"},
	)
	got, err := parseDsym(data)
	if err != nil {
		t.Fatalf("parseDsym: %v", err)
	}
	want := map[string]uint32{"counter": 0x01400000, "main": 0x01A00000, "loop": 0x01A00010}
	if len(got) != len(want) {
		t.Fatalf("symbols = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("symbol %q = 0x%08X, want 0x%08X", k, got[k], v)
		}
	}
}

func TestParseDsymEmpty(t *testing.T) {
	got, err := parseDsym(dsymData(0))
	if err != nil || len(got) != 0 {
		t.Fatalf("parseDsym = %v, %v; want empty map", got, err)
	}
}

func TestParseDsymErrors(t *testing.T) {
	main := dsymEntry{0x01A00000, 1, 0, "main"}
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"empty data", nil, "dSYM: chunk too short for the entry count"},
		{"count too big", dsymData(2, main), "dSYM: entry 1 runs past the end of the chunk"},
		{"trailing bytes", append(dsymData(1, main), 0, 0, 0, 0), "dSYM: 4 trailing bytes after 1 entries"},
		{"zero name length", dsymData(1, dsymEntry{0x01A00000, 1, 0, ""}), "dSYM: entry 0: empty name"},
		{"kind 0", dsymData(1, dsymEntry{0x01A00000, 0, 0, "main"}), "dSYM: entry 0: unknown section kind 0x00"},
		{"kind 5", dsymData(1, dsymEntry{0x01A00000, 5, 0, "main"}), "dSYM: entry 0: unknown section kind 0x05"},
		{"header past end", dsymData(1, main)[:10], "dSYM: entry 0 runs past the end of the chunk"},
		{"name past end", dsymData(1, main)[:13], "dSYM: entry 0: name runs past the end of the chunk"},
		{"duplicate name", dsymData(2, main, dsymEntry{0x01A00004, 1, 0, "main"}), `dSYM: duplicate symbol "main"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parseDsym(c.data)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
		})
	}
	// 16 bytes: count (4), entry header (8), name "main" (4)
	for _, c := range []struct {
		cut  int
		want string
	}{
		{10, "dSYM: entry 0 runs past the end of the chunk"},
		{13, "dSYM: entry 0: name runs past the end of the chunk"},
		{15, "dSYM: entry 0: name runs past the end of the chunk"},
	} {
		if _, err := parseDsym(dsymData(1, main)[:c.cut]); err == nil || err.Error() != c.want {
			t.Errorf("cut to %d bytes: err = %v, want %q", c.cut, err, c.want)
		}
	}
	_, err := parseDsym(dsymData(2, main, dsymEntry{0x01A00004, 1, 0, "main"}))
	if err == nil || err.Error() != `dSYM: duplicate symbol "main"` {
		t.Fatalf("err = %v, want duplicate symbol \"main\"", err)
	}
}
