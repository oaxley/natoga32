package preprocess

import (
	"fmt"
	"strings"

	"github.com/natoga32/goasm/internal/token"
)

// handlePaste processes ## token pasting
// Takes the previous token from result and concatenates it with the next token
func (p *Preprocessor) handlePaste(ts *tokenStream, result *[]token.Token) (token.Token, error) {
	pasteTok := ts.advance() // consume ##

	// Get left side (pop from result)
	if len(*result) == 0 {
		return token.Token{}, locationError(pasteTok, "## needs a token on its left")
	}
	left := (*result)[len(*result)-1]
	*result = (*result)[:len(*result)-1]

	// Get right side
	if ts.atEnd() || ts.peek().Type == token.EOL || ts.peek().Type == token.EOF {
		return token.Token{}, locationError(pasteTok, "## needs a token on its right")
	}
	right := ts.advance()

	// Expand left side if it's a defined symbol
	leftVal := left.Value
	if left.Type == token.IDENT {
		if val, ok := p.defines[left.Value]; ok {
			leftVal = fmt.Sprintf("%d", val)
		}
	}

	// Expand right side if it's a defined symbol
	rightVal := right.Value
	if right.Type == token.IDENT {
		if val, ok := p.defines[right.Value]; ok {
			rightVal = fmt.Sprintf("%d", val)
		}
	}

	// Handle LABEL on right side (has trailing colon)
	isLabel := right.Type == token.LABEL
	if isLabel {
		// Remove the colon for concatenation, we'll add it back
		rightVal = strings.TrimSuffix(rightVal, ":")
	}

	// Concatenate
	pastedValue := leftVal + rightVal

	// Determine result type
	resultType := token.IDENT
	if isLabel {
		pastedValue += ":"
		resultType = token.LABEL
	}

	return token.Token{
		Type:     resultType,
		Value:    pastedValue,
		Filename: left.Filename,
		Row:      left.Row,
		Col:      left.Col,
	}, nil
}
