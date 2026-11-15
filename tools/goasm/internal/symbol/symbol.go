// Symbol table for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package symbol

import "fmt"

// SymbolType represents the type of a symbol
type SymbolType int

const (
	SymLabel    SymbolType = iota // Label (code address)
	SymConstant                   // Constant value (.equ, .set)
	SymData                       // Data label
	SymExternal                   // External symbol (not defined in this file)
)

// String returns the string representation of a symbol type
func (t SymbolType) String() string {
	names := [...]string{
		SymLabel:    "LABEL",
		SymConstant: "CONST",
		SymData:     "DATA",
		SymExternal: "EXTERN",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "UNKNOWN"
}

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name    string     // Symbol name
	Type    SymbolType // Symbol type
	Value   int64      // Address or value
	Section string     // Section where the symbol is defined
	Defined bool       // True if symbol has been defined
	Global  bool       // True if symbol is exported (global)
	Line    int        // Source line where symbol was defined
}

// Table is the symbol table
type Table struct {
	symbols    map[string]*Symbol
	localCount int // Counter for generating local labels
}

// New creates a new symbol table
func New() *Table {
	return &Table{
		symbols:    make(map[string]*Symbol),
		localCount: 0,
	}
}

// Define defines a symbol with the given name and value.
// Returns an error if the symbol is already defined.
func (t *Table) Define(name string, typ SymbolType, value int64, section string, line int) error {
	if existing, ok := t.symbols[name]; ok && existing.Defined {
		return fmt.Errorf("symbol '%s' already defined at line %d", name, existing.Line)
	}

	t.symbols[name] = &Symbol{
		Name:    name,
		Type:    typ,
		Value:   value,
		Section: section,
		Defined: true,
		Global:  false,
		Line:    line,
	}
	return nil
}

// DefineLabel is a convenience function for defining a label
func (t *Table) DefineLabel(name string, address int64, section string, line int) error {
	return t.Define(name, SymLabel, address, section, line)
}

// DefineConstant is a convenience function for defining a constant
func (t *Table) DefineConstant(name string, value int64, line int) error {
	return t.Define(name, SymConstant, value, "", line)
}

// Reference creates a reference to an undefined symbol.
// This is used for forward references.
func (t *Table) Reference(name string) {
	if _, ok := t.symbols[name]; !ok {
		t.symbols[name] = &Symbol{
			Name:    name,
			Defined: false,
		}
	}
}

// Lookup returns the symbol with the given name, or nil if not found
func (t *Table) Lookup(name string) *Symbol {
	return t.symbols[name]
}

// IsDefined returns true if the symbol is defined
func (t *Table) IsDefined(name string) bool {
	sym := t.symbols[name]
	return sym != nil && sym.Defined
}

// SetGlobal marks a symbol as global (exported)
func (t *Table) SetGlobal(name string) error {
	sym := t.symbols[name]
	if sym == nil {
		// Create a reference for forward declaration
		t.symbols[name] = &Symbol{
			Name:    name,
			Defined: false,
			Global:  true,
		}
		return nil
	}
	sym.Global = true
	return nil
}

// UpdateValue updates the value of an existing symbol
func (t *Table) UpdateValue(name string, value int64) error {
	sym := t.symbols[name]
	if sym == nil {
		return fmt.Errorf("symbol '%s' not found", name)
	}
	sym.Value = value
	return nil
}

// AllSymbols returns all symbols in the table
func (t *Table) AllSymbols() []*Symbol {
	result := make([]*Symbol, 0, len(t.symbols))
	for _, sym := range t.symbols {
		result = append(result, sym)
	}
	return result
}

// DefinedSymbols returns only the defined symbols
func (t *Table) DefinedSymbols() []*Symbol {
	result := make([]*Symbol, 0)
	for _, sym := range t.symbols {
		if sym.Defined {
			result = append(result, sym)
		}
	}
	return result
}

// UndefinedSymbols returns symbols that were referenced but not defined
func (t *Table) UndefinedSymbols() []*Symbol {
	result := make([]*Symbol, 0)
	for _, sym := range t.symbols {
		if !sym.Defined {
			result = append(result, sym)
		}
	}
	return result
}

// GlobalSymbols returns symbols marked as global
func (t *Table) GlobalSymbols() []*Symbol {
	result := make([]*Symbol, 0)
	for _, sym := range t.symbols {
		if sym.Global {
			result = append(result, sym)
		}
	}
	return result
}

// Clear removes all symbols from the table
func (t *Table) Clear() {
	t.symbols = make(map[string]*Symbol)
	t.localCount = 0
}

// Size returns the number of symbols in the table
func (t *Table) Size() int {
	return len(t.symbols)
}
