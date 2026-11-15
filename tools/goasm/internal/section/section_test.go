package section

import "testing"

func TestNewSection(t *testing.T) {
	sec := NewSection(".text", SectionText)

	if sec.Name != ".text" {
		t.Errorf("expected name '.text', got '%s'", sec.Name)
	}
	if sec.Type != SectionText {
		t.Errorf("expected type SectionText, got %v", sec.Type)
	}
	if sec.Size != 0 {
		t.Errorf("expected size 0, got %d", sec.Size)
	}
}

func TestSectionEmit(t *testing.T) {
	sec := NewSection(".text", SectionText)

	sec.Emit([]byte{0x01, 0x02, 0x03, 0x04})
	if sec.Size != 4 {
		t.Errorf("expected size 4, got %d", sec.Size)
	}
	if len(sec.Data) != 4 {
		t.Errorf("expected data length 4, got %d", len(sec.Data))
	}
}

func TestSectionEmitWord(t *testing.T) {
	sec := NewSection(".text", SectionText)
	sec.BaseAddr = 0x100

	sec.EmitWord(0x12345678)

	if sec.Size != 4 {
		t.Errorf("expected size 4, got %d", sec.Size)
	}

	// Check little-endian encoding
	expected := []byte{0x78, 0x56, 0x34, 0x12}
	for i, b := range expected {
		if sec.Data[i] != b {
			t.Errorf("data[%d]: expected 0x%02X, got 0x%02X", i, b, sec.Data[i])
		}
	}
}

func TestSectionEmitHalf(t *testing.T) {
	sec := NewSection(".data", SectionData)

	sec.EmitHalf(0xABCD)

	if sec.Size != 2 {
		t.Errorf("expected size 2, got %d", sec.Size)
	}

	// Check little-endian encoding
	if sec.Data[0] != 0xCD || sec.Data[1] != 0xAB {
		t.Errorf("expected [0xCD, 0xAB], got [0x%02X, 0x%02X]", sec.Data[0], sec.Data[1])
	}
}

func TestSectionAlign(t *testing.T) {
	sec := NewSection(".text", SectionText)
	sec.BaseAddr = 0x100

	sec.EmitByte(0xFF) // Size = 1, address = 0x101
	sec.Align(4)       // Should pad to 0x104

	expectedSize := int64(4)
	if sec.Size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, sec.Size)
	}

	// Check padding bytes are zero
	for i := 1; i < 4; i++ {
		if sec.Data[i] != 0 {
			t.Errorf("padding byte %d: expected 0, got 0x%02X", i, sec.Data[i])
		}
	}
}

func TestSectionCurrentAddress(t *testing.T) {
	sec := NewSection(".text", SectionText)
	sec.BaseAddr = 0x1000

	if sec.CurrentAddress() != 0x1000 {
		t.Errorf("expected address 0x1000, got 0x%X", sec.CurrentAddress())
	}

	sec.EmitWord(0x12345678)
	if sec.CurrentAddress() != 0x1004 {
		t.Errorf("expected address 0x1004, got 0x%X", sec.CurrentAddress())
	}
}
