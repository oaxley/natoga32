package preprocess

import (
	"fmt"

	"github.com/natoga32/goasm/internal/token"
)

// handleFor processes .for VAR = START, END [, STEP] ... .endf
// Expands the loop body for each iteration, substituting the loop variable
func (p *Preprocessor) handleFor(ts *tokenStream) ([]token.Token, error) {
	row := ts.peek().Row
	ts.advance() // consume .for

	// Parse variable name
	if !ts.check(token.IDENT) {
		return nil, fmt.Errorf("expected identifier after .for at line %d", row)
	}
	varName := ts.advance().Value

	// Expect '='
	if !ts.check(token.ASSIGN) {
		return nil, fmt.Errorf("expected '=' after variable name in .for at line %d", row)
	}
	ts.advance()

	// Parse start value (can be expression)
	var startToks []token.Token
	for !ts.atEnd() && ts.peek().Type != token.COMMA && ts.peek().Type != token.EOL && ts.peek().Type != token.EOF {
		startToks = append(startToks, ts.advance())
	}
	start, err := p.evaluateExpr(startToks)
	if err != nil {
		return nil, fmt.Errorf("error evaluating .for start value at line %d: %v", row, err)
	}

	// Expect ','
	if !ts.check(token.COMMA) {
		return nil, fmt.Errorf("expected ',' after start value in .for at line %d", row)
	}
	ts.advance()

	// Parse end value (can be expression)
	var endToks []token.Token
	for !ts.atEnd() && ts.peek().Type != token.COMMA && ts.peek().Type != token.EOL && ts.peek().Type != token.EOF {
		endToks = append(endToks, ts.advance())
	}
	end, err := p.evaluateExpr(endToks)
	if err != nil {
		return nil, fmt.Errorf("error evaluating .for end value at line %d: %v", row, err)
	}

	// Parse optional step value
	step := int64(1)
	if start > end {
		step = -1 // Default to -1 if counting down
	}
	if ts.check(token.COMMA) {
		ts.advance()
		var stepToks []token.Token
		for !ts.atEnd() && ts.peek().Type != token.EOL && ts.peek().Type != token.EOF {
			stepToks = append(stepToks, ts.advance())
		}
		step, err = p.evaluateExpr(stepToks)
		if err != nil {
			return nil, fmt.Errorf("error evaluating .for step value at line %d: %v", row, err)
		}
	}

	// Skip EOL after .for line
	if ts.check(token.EOL) {
		ts.advance()
	}

	// Validate step
	if step == 0 {
		return nil, fmt.Errorf(".for step cannot be zero at line %d", row)
	}

	// Collect loop body until .endf
	body, err := p.collectForBody(ts)
	if err != nil {
		return nil, err
	}

	// Expand the loop
	var result []token.Token
	for i := start; (step > 0 && i < end) || (step < 0 && i > end); i += step {
		// Expand body with variable substitution
		expanded := p.expandForBody(body, varName, i)
		result = append(result, expanded...)
	}

	return result, nil
}

// collectForBody collects tokens between .for and .endf, handling nesting
func (p *Preprocessor) collectForBody(ts *tokenStream) ([]token.Token, error) {
	var body []token.Token
	depth := 1

	for !ts.atEnd() && depth > 0 {
		tok := ts.peek()

		if tok.Type == token.EOF {
			break
		}

		if tok.Type == token.DIRECTIVE {
			switch tok.Value {
			case ".for":
				depth++
			case ".endf":
				depth--
				if depth == 0 {
					ts.advance() // consume .endf
					// Skip EOL after .endf
					if ts.check(token.EOL) {
						ts.advance()
					}
					return body, nil
				}
			}
		}

		body = append(body, ts.advance())
	}

	if depth != 0 {
		return nil, fmt.Errorf("unterminated .for loop (missing .endf)")
	}

	return body, nil
}

// expandForBody expands the loop body, substituting the loop variable
func (p *Preprocessor) expandForBody(body []token.Token, varName string, value int64) []token.Token {
	result := make([]token.Token, 0, len(body))

	for _, tok := range body {
		if tok.Type == token.IDENT && tok.Value == varName {
			// Substitute loop variable with its current value
			result = append(result, token.Token{
				Type:  token.NUMBER,
				Value: fmt.Sprintf("%d", value),
				Row:   tok.Row,
				Col:   tok.Col,
			})
		} else {
			result = append(result, tok)
		}
	}

	return result
}
