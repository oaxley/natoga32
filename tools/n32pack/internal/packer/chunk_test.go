// Tests of the cartridge chunk helpers
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package packer

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendChunk(t *testing.T) {
	c := appendChunk(nil, "RSCV", []byte{1, 2, 3, 4, 5})
	if len(c) != 20 {
		t.Fatalf("chunk size = %d, want 20", len(c))
	}
	if got := binary.LittleEndian.Uint32(c[0:4]); got != 5 {
		t.Errorf("length = %d, want 5", got)
	}
	if string(c[4:8]) != "RSCV" {
		t.Errorf("type = %q, want RSCV", c[4:8])
	}
	if want := []byte{1, 2, 3, 4, 5, 0, 0, 0}; !bytes.Equal(c[8:16], want) {
		t.Errorf("data = % x, want % x", c[8:16], want)
	}
	want := crc32.ChecksumIEEE([]byte("RSCV\x01\x02\x03\x04\x05"))
	if got := binary.LittleEndian.Uint32(c[16:20]); got != want {
		t.Errorf("CRC = %#08x, want %#08x", got, want)
	}
}

func TestAppendChunkEmpty(t *testing.T) {
	c := appendChunk(nil, "IEND", nil)
	if len(c) != 12 {
		t.Fatalf("chunk size = %d, want 12", len(c))
	}
	want := crc32.ChecksumIEEE([]byte("IEND"))
	if want != 0xAE426082 {
		t.Fatalf("IEND CRC = %#08x, want 0xAE426082", want)
	}
	if got := binary.LittleEndian.Uint32(c[8:12]); got != want {
		t.Errorf("CRC = %#08x, want %#08x", got, want)
	}
}

func TestAppendChunkKeepsPrefix(t *testing.T) {
	c := appendChunk([]byte{0xAA, 0xBB}, "IEND", nil)
	if len(c) != 14 || c[0] != 0xAA || c[1] != 0xBB {
		t.Fatalf("prefix lost: % x", c)
	}
	if !bytes.Equal(c[2:], appendChunk(nil, "IEND", nil)) {
		t.Errorf("chunk after prefix = % x", c[2:])
	}
}

func TestAlign4(t *testing.T) {
	for _, tc := range []struct{ in, want uint64 }{{0, 0}, {1, 4}, {3, 4}, {4, 4}, {5, 8}, {0xFFFFFFFF, 0x100000000}} {
		if got := align4(tc.in); got != tc.want {
			t.Errorf("align4(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadChunkFile(t *testing.T) {
	payload := []byte{1, 2, 3, 4, 5}
	good := appendChunk(nil, "RSCV", payload)

	t.Run("round trip", func(t *testing.T) {
		data, err := readChunkFile(writeTemp(t, "a.rscv", good), "RSCV")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, payload) {
			t.Errorf("data = % x, want % x", data, payload)
		}
	})

	t.Run("empty chunk", func(t *testing.T) {
		data, err := readChunkFile(writeTemp(t, "a.dsym", appendChunk(nil, "dSYM", nil)), "dSYM")
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != 0 {
			t.Errorf("data = % x, want empty", data)
		}
	})

	bad := func(name string, file []byte, want, msg string) {
		t.Run(name, func(t *testing.T) {
			path := writeTemp(t, "a.bin", file)
			_, err := readChunkFile(path, want)
			if err == nil {
				t.Fatal("no error")
			}
			if !strings.HasPrefix(err.Error(), path+": ") || !strings.Contains(err.Error(), msg) {
				t.Errorf("error = %q, want %q prefixed with the path", err, msg)
			}
		})
	}

	bad("wrong type", good, "dSYM", "not a dSYM chunk")
	bad("truncated", good[:len(good)-1], "RSCV", "bad chunk length")
	bad("too short", good[:6], "RSCV", "bad chunk length")
	bad("trailing bytes", append(append([]byte{}, good...), 0, 0, 0, 0), "RSCV", "bad chunk length")

	// good holds 5 data bytes, so bytes 13-15 are padding (the CRC does not cover them)
	padded := append([]byte{}, good...)
	padded[13] = 0xFF
	bad("non-zero padding", padded, "RSCV", "chunk padding is not zero")

	flipped := append([]byte{}, good...)
	flipped[len(flipped)-1] ^= 0x01
	bad("flipped CRC", flipped, "RSCV", "bad chunk CRC")

	t.Run("missing file", func(t *testing.T) {
		if _, err := readChunkFile(filepath.Join(t.TempDir(), "none.rscv"), "RSCV"); err == nil {
			t.Fatal("no error")
		}
	})
}
