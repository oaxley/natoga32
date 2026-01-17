// Token types and Token struct for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package token

import "fmt"

// Type represents the type of a token
type Type int

const (
	// Control characters
	EOF Type = iota // End of file
	EOL             // End of line '\n'

	// Numbers & identifiers
	NUMBER    // Any numbers (decimal, hex, binary, octal)
	DIRECTIVE // Assembler directive (starts with '.')
	IDENT     // Identifier
	LABEL     // Identifier terminated with ':'
	STRING    // A "string"
	CHAR      // A single char 'A'
	ENVVAR    // Environment variable

	// Comparators
	EQ  // '=='
	NEQ // '!='
	GT  // '>'
	LT  // '<'
	GTE // '>='
	LTE // '<='

	// Operators & symbols
	ASSIGN // '='
	LPAREN // '('
	RPAREN // ')'
	COMMA  // ','
	PLUS   // '+'
	MINUS  // '-'
	STAR   // '*'
	SLASH  // '/'
	MODULO // '%'
	LSHIFT // '<<'
	RSHIFT // '>>'
	AND    // '&'
	OR     // '|'
	XOR    // '^'
	DOLLAR // '$' - alias for current PC

	// Misc
	PASTE       // Token pasting '##'
	MACRO_PARAM // Macro parameter reference '\name'
	SKIP        // Spaces and tabs (filtered out)
	COMMENT     // Comment (filtered out)

	// Relocation tokens
	ABS_LO   // %lo
	ABS_HI   // %hi
	PCREL_LO // %pcrel_lo
	PCREL_HI // %pcrel_hi
)

// String returns the string representation of a token type
func (t Type) String() string {
	names := [...]string{
		EOF:         "EOF",
		EOL:         "EOL",
		NUMBER:      "NUMBER",
		DIRECTIVE:   "DIRECTIVE",
		IDENT:       "IDENT",
		LABEL:       "LABEL",
		STRING:      "STRING",
		CHAR:        "CHAR",
		ENVVAR:      "ENVVAR",
		EQ:          "EQ",
		NEQ:         "NEQ",
		GT:          "GT",
		LT:          "LT",
		GTE:         "GTE",
		LTE:         "LTE",
		ASSIGN:      "ASSIGN",
		LPAREN:      "LPAREN",
		RPAREN:      "RPAREN",
		COMMA:       "COMMA",
		PLUS:        "PLUS",
		MINUS:       "MINUS",
		STAR:        "STAR",
		SLASH:       "SLASH",
		MODULO:      "MODULO",
		LSHIFT:      "LSHIFT",
		RSHIFT:      "RSHIFT",
		AND:         "AND",
		OR:          "OR",
		XOR:         "XOR",
		DOLLAR:      "DOLLAR",
		PASTE:       "PASTE",
		MACRO_PARAM: "MACRO_PARAM",
		SKIP:        "SKIP",
		COMMENT:     "COMMENT",
		ABS_LO:      "ABS_LO",
		ABS_HI:      "ABS_HI",
		PCREL_LO:    "PCREL_LO",
		PCREL_HI:    "PCREL_HI",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "UNKNOWN"
}

// Token represents a single token from the source code
type Token struct {
	Type     Type   // The type of the token
	Value    string // The literal value as seen in the source
	Filename string // Source filename
	Row      int    // Line number (1-based)
	Col      int    // Column number (1-based)
}

// String returns a string representation of the token
func (t Token) String() string {
	if t.Filename != "" {
		return fmt.Sprintf("%s;%s;%d;%d;%s", t.Type, t.Filename, t.Row, t.Col, t.Value)
	}
	return fmt.Sprintf("%s;%d;%d;%s", t.Type, t.Row, t.Col, t.Value)
}

// New creates a new token
func New(typ Type, value string, row, col int) Token {
	return Token{
		Type:  typ,
		Value: value,
		Row:   row,
		Col:   col,
	}
}

// NewWithFile creates a new token with a filename
func NewWithFile(typ Type, value, filename string, row, col int) Token {
	return Token{
		Type:     typ,
		Value:    value,
		Filename: filename,
		Row:      row,
		Col:      col,
	}
}

// IsOperator returns true if the token is a binary operator
func (t Token) IsOperator() bool {
	switch t.Type {
	case PLUS, MINUS, STAR, SLASH, MODULO, LSHIFT, RSHIFT, AND, OR, XOR:
		return true
	}
	return false
}

// IsRelocation returns true if the token is a relocation modifier
func (t Token) IsRelocation() bool {
	switch t.Type {
	case ABS_LO, ABS_HI, PCREL_LO, PCREL_HI:
		return true
	}
	return false
}
