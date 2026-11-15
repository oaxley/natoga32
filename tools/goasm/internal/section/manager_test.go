package section

import "testing"

func TestManagerDefault(t *testing.T) {
	m := NewManager()

	if m.CurrentName() != ".text" {
		t.Errorf("expected current section '.text', got '%s'", m.CurrentName())
	}

	// Check default sections exist
	sections := []string{".text", ".data", ".bss", ".rom"}
	for _, name := range sections {
		if m.Get(name) == nil {
			t.Errorf("default section '%s' not found", name)
		}
	}
}

func TestManagerSetSection(t *testing.T) {
	m := NewManager()

	err := m.SetSection(".data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.CurrentName() != ".data" {
		t.Errorf("expected current section '.data', got '%s'", m.CurrentName())
	}

	err = m.SetSection(".nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent section")
	}
}

func TestManagerEmit(t *testing.T) {
	m := NewManager()

	m.EmitWord(0xDEADBEEF)
	m.EmitByte(0xFF)

	sec := m.Get(".text")
	if sec.Size != 5 {
		t.Errorf("expected size 5, got %d", sec.Size)
	}
}

func TestManagerSwitchAndEmit(t *testing.T) {
	m := NewManager()

	// Emit to .text
	m.EmitWord(0x11111111)

	// Switch to .data and emit
	m.SetSection(".data")
	m.EmitWord(0x22222222)

	// Check sizes
	textSec := m.Get(".text")
	dataSec := m.Get(".data")

	if textSec.Size != 4 {
		t.Errorf("expected .text size 4, got %d", textSec.Size)
	}
	if dataSec.Size != 4 {
		t.Errorf("expected .data size 4, got %d", dataSec.Size)
	}
}

func TestManagerAll(t *testing.T) {
	m := NewManager()

	all := m.All()
	if len(all) != 4 {
		t.Errorf("expected 4 sections, got %d", len(all))
	}

	// Check order
	expectedOrder := []string{".text", ".data", ".bss", ".rom"}
	for i, sec := range all {
		if sec.Name != expectedOrder[i] {
			t.Errorf("section %d: expected '%s', got '%s'", i, expectedOrder[i], sec.Name)
		}
	}
}

func TestSectionReserve(t *testing.T) {
	bss := NewSection(".bss", SectionBss)
	bss.Reserve(100)

	if bss.Size != 100 {
		t.Errorf("expected size 100, got %d", bss.Size)
	}
	// BSS should not have data
	if len(bss.Data) != 0 {
		t.Errorf("BSS should not have data, got %d bytes", len(bss.Data))
	}

	// Data section should have zero-filled data
	data := NewSection(".data", SectionData)
	data.Reserve(10)
	if len(data.Data) != 10 {
		t.Errorf("data section should have 10 bytes, got %d", len(data.Data))
	}
}

func TestManagerSetBaseAddresses(t *testing.T) {
	m := NewManager()

	m.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000)

	if m.Get(".text").BaseAddr != 0x1000 {
		t.Errorf("expected .text base 0x1000, got 0x%X", m.Get(".text").BaseAddr)
	}
	if m.Get(".data").BaseAddr != 0x2000 {
		t.Errorf("expected .data base 0x2000, got 0x%X", m.Get(".data").BaseAddr)
	}
	if m.Get(".bss").BaseAddr != 0x3000 {
		t.Errorf("expected .bss base 0x3000, got 0x%X", m.Get(".bss").BaseAddr)
	}
	if m.Get(".rom").BaseAddr != 0x4000 {
		t.Errorf("expected .rom base 0x4000, got 0x%X", m.Get(".bss").BaseAddr)
	}

}
