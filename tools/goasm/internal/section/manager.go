package section

import "fmt"

// Manager manages multiple sections
type Manager struct {
	sections map[string]*Section
	current  *Section
	order    []string // Maintains section order
}

// NewManager creates a new section manager with default sections
func NewManager() *Manager {
	m := &Manager{
		sections: make(map[string]*Section),
		order:    make([]string, 0),
	}

	// Create default sections
	m.AddSection(SectionText.String(), SectionText)
	m.AddSection(SectionData.String(), SectionData)
	m.AddSection(SectionBss.String(), SectionBss)

	// Start in .text section
	m.current = m.sections[".text"]

	return m
}

// AddSection adds a new section
func (m *Manager) AddSection(name string, typ SectionType) *Section {
	if _, exists := m.sections[name]; exists {
		return m.sections[name]
	}
	sec := NewSection(name, typ)
	m.sections[name] = sec
	m.order = append(m.order, name)
	return sec
}

// SetSection switches to the specified section
func (m *Manager) SetSection(name string) error {
	sec, ok := m.sections[name]
	if !ok {
		return fmt.Errorf("undefined section: %s", name)
	}
	m.current = sec
	return nil
}

// Current returns the current section
func (m *Manager) Current() *Section {
	return m.current
}

// CurrentName returns the name of the current section
func (m *Manager) CurrentName() string {
	if m.current == nil {
		return ""
	}
	return m.current.Name
}

// Get returns the section with the given name
func (m *Manager) Get(name string) *Section {
	return m.sections[name]
}

// All returns all sections in order
func (m *Manager) All() []*Section {
	result := make([]*Section, 0, len(m.order))
	for _, name := range m.order {
		result = append(result, m.sections[name])
	}
	return result
}

// Emit emits data to the current section
func (m *Manager) Emit(data []byte) {
	if m.current != nil {
		m.current.Emit(data)
	}
}

// EmitByte emits a byte to the current section
func (m *Manager) EmitByte(b byte) {
	if m.current != nil {
		m.current.EmitByte(b)
	}
}

// EmitWord emits a 32-bit word to the current section
func (m *Manager) EmitWord(w uint32) {
	if m.current != nil {
		m.current.EmitWord(w)
	}
}

// EmitHalf emits a 16-bit halfword to the current section
func (m *Manager) EmitHalf(h uint16) {
	if m.current != nil {
		m.current.EmitHalf(h)
	}
}

// Align aligns the current section
func (m *Manager) Align(alignment int) {
	if m.current != nil {
		m.current.Align(alignment)
	}
}

// CurrentAddress returns the current address in the current section
func (m *Manager) CurrentAddress() int64 {
	if m.current == nil {
		return 0
	}
	return m.current.CurrentAddress()
}

// SetBaseAddresses sets base addresses for all sections
// This should be called after all code/data has been emitted
func (m *Manager) SetBaseAddresses(textBase, dataBase, bssBase int64) {
	if sec := m.sections[".text"]; sec != nil {
		sec.BaseAddr = textBase
	}
	if sec := m.sections[".data"]; sec != nil {
		sec.BaseAddr = dataBase
	}
	if sec := m.sections[".bss"]; sec != nil {
		sec.BaseAddr = bssBase
	}
}

// TotalSize returns the total size of all sections
func (m *Manager) TotalSize() int64 {
	var total int64
	for _, sec := range m.sections {
		total += sec.Size
	}
	return total
}
