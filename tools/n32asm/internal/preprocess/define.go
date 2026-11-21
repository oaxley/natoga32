package preprocess

import (
	"github.com/natoga32/n32asm/internal/token"
)

// handleDefine processes ".define NAME value" and ".define NAME, value".
// With no value, NAME is defined as 0 (a flag for .ifdef).
func (p *Preprocessor) handleDefine(ts *tokenStream) error {
	defTok := ts.advance() // consume .define

	// Get the name
	if !ts.check(token.IDENT) {
		return locationError(defTok, "expected a name after .define")
	}
	name := ts.advance().Value

	// an optional comma between the name and the value
	comma := false
	if ts.check(token.COMMA) {
		ts.advance()
		comma = true
	}

	// Collect tokens until EOL
	var valueToks []token.Token
	for !ts.atEnd() && ts.peek().Type != token.EOL && ts.peek().Type != token.EOF {
		valueToks = append(valueToks, ts.advance())
	}

	// Skip EOL
	if ts.check(token.EOL) {
		ts.advance()
	}

	if comma && len(valueToks) == 0 {
		return locationError(defTok, ".define %s: expected a value after ','", name)
	}

	// Evaluate the expression
	value, err := p.evaluateExpr(valueToks)
	if err != nil {
		return locationError(defTok, ".define %s: %v", name, err)
	}

	p.defines[name] = value
	return nil
}
