// Section management for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package section

// SectionType represents the type of section
type SectionType int

const (
	SectionText   SectionType = iota // Executable code
	SectionData                      // Initialized data
	SectionRodata                    // Read-only data
	SectionBss                       // Uninitialized data
)

// String returns the string representation of a section type
func (t SectionType) String() string {
	names := [...]string{
		SectionText:   ".text",
		SectionData:   ".data",
		SectionRodata: ".rodata",
		SectionBss:    ".bss",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return ".unknown"
}

// Section represents a section in the object file
type Section struct {
	Name      string      // Section name (e.g., ".text", ".data")
	Type      SectionType // Section type
	BaseAddr  int64       // Base address of the section
	Size      int64       // Current size in bytes
	Alignment int         // Required alignment (power of 2)
	Data      []byte      // Section contents
}

// NewSection creates a new section with the given name and type
func NewSection(name string, typ SectionType) *Section {
	return &Section{
		Name:      name,
		Type:      typ,
		BaseAddr:  0,
		Size:      0,
		Alignment: 4, // Default to 4-byte alignment
		Data:      make([]byte, 0),
	}
}

// CurrentAddress returns the current address (base + size)
func (s *Section) CurrentAddress() int64 {
	return s.BaseAddr + s.Size
}

// Emit appends data to the section
func (s *Section) Emit(data []byte) {
	s.Data = append(s.Data, data...)
	s.Size += int64(len(data))
}

// EmitByte appends a single byte
func (s *Section) EmitByte(b byte) {
	s.Data = append(s.Data, b)
	s.Size++
}

// EmitWord appends a 32-bit word (little-endian)
func (s *Section) EmitWord(w uint32) {
	s.Data = append(s.Data, byte(w), byte(w>>8), byte(w>>16), byte(w>>24))
	s.Size += 4
}

// EmitHalf appends a 16-bit halfword (little-endian)
func (s *Section) EmitHalf(h uint16) {
	s.Data = append(s.Data, byte(h), byte(h>>8))
	s.Size += 2
}

// Align pads the section to the specified alignment
func (s *Section) Align(alignment int) {
	if alignment <= 0 {
		return
	}
	current := s.CurrentAddress()
	aligned := (current + int64(alignment) - 1) & ^(int64(alignment) - 1)
	padding := aligned - current
	for i := int64(0); i < padding; i++ {
		s.EmitByte(0)
	}
}

// Reserve allocates space without initializing it (for .bss)
func (s *Section) Reserve(size int64) {
	s.Size += size
	// Don't add to Data for BSS sections
	if s.Type != SectionBss {
		for i := int64(0); i < size; i++ {
			s.Data = append(s.Data, 0)
		}
	}
}
