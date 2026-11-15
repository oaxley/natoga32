package preprocess

import (
	"fmt"

	"github.com/natoga32/goasm/internal/token"
)

// handleDup processes .dup COUNT VALUE
// Expands to VALUE repeated COUNT times, separated by commas
func (p *Preprocessor) handleDup(ts *tokenStream) ([]token.Token, error) {
	row := ts.peek().Row
	ts.advance() // consume .dup

	// Parse count - can be a simple number/ident or expression in parentheses
	var countToks []token.Token
	if ts.check(token.LPAREN) {
		ts.advance() // consume (
		depth := 1
		for !ts.atEnd() && depth > 0 {
			tok := ts.peek()
			if tok.Type == token.LPAREN {
				depth++
			} else if tok.Type == token.RPAREN {
				depth--
				if depth == 0 {
					ts.advance() // consume closing )
					break
				}
			}
			countToks = append(countToks, ts.advance())
		}
	} else {
		// Single token count
		if ts.atEnd() || ts.peek().Type == token.EOL || ts.peek().Type == token.EOF {
			return nil, fmt.Errorf("expected count after .dup at line %d", row)
		}
		countToks = append(countToks, ts.advance())
	}

	// Evaluate count
	count, err := p.evaluateExpr(countToks)
	if err != nil {
		return nil, fmt.Errorf("error evaluating .dup count at line %d: %v", row, err)
	}

	if count < 0 {
		return nil, fmt.Errorf(".dup count cannot be negative at line %d", row)
	}

	// Parse value - can be a simple token or expression in parentheses
	var valueToks []token.Token
	if ts.check(token.LPAREN) {
		ts.advance() // consume (
		depth := 1
		for !ts.atEnd() && depth > 0 {
			tok := ts.peek()
			if tok.Type == token.LPAREN {
				depth++
			} else if tok.Type == token.RPAREN {
				depth--
				if depth == 0 {
					ts.advance() // consume closing )
					break
				}
			}
			valueToks = append(valueToks, ts.advance())
		}
	} else {
		// Single token value
		if ts.atEnd() || ts.peek().Type == token.EOL || ts.peek().Type == token.EOF {
			return nil, fmt.Errorf("expected value after .dup count at line %d", row)
		}
		valueToks = append(valueToks, ts.advance())
	}

	// Evaluate value
	value, err := p.evaluateExpr(valueToks)
	if err != nil {
		return nil, fmt.Errorf("error evaluating .dup value at line %d: %v", row, err)
	}

	// Generate repeated values with commas
	var result []token.Token
	for i := int64(0); i < count; i++ {
		if i > 0 {
			result = append(result, token.Token{
				Type:  token.COMMA,
				Value: ",",
				Row:   row,
			})
		}
		result = append(result, token.Token{
			Type:  token.NUMBER,
			Value: fmt.Sprintf("%d", value),
			Row:   row,
		})
	}

	return result, nil
}
