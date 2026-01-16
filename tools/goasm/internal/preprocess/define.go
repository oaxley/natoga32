package preprocess

import (
	"fmt"

	"github.com/natoga32/goasm/internal/token"
)

// handleDefine processes .define NAME value
func (p *Preprocessor) handleDefine(ts *tokenStream) error {
	ts.advance() // consume .define

	// Get the name
	if !ts.check(token.IDENT) {
		return fmt.Errorf("expected identifier after .define at line %d", ts.peek().Row)
	}
	name := ts.advance().Value

	// Collect tokens until EOL
	var valueToks []token.Token
	for !ts.atEnd() && ts.peek().Type != token.EOL && ts.peek().Type != token.EOF {
		valueToks = append(valueToks, ts.advance())
	}

	// Skip EOL
	if ts.check(token.EOL) {
		ts.advance()
	}

	// Evaluate the expression
	value, err := p.evaluateExpr(valueToks)
	if err != nil {
		return fmt.Errorf("error evaluating .define expression: %v", err)
	}

	p.defines[name] = value
	return nil
}
