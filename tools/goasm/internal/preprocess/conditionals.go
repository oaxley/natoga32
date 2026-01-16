package preprocess

import (
	"fmt"

	"github.com/natoga32/goasm/internal/token"
)

// handleConditional processes .ifdef/.ifndef NAME ... .else ... .endif
func (p *Preprocessor) handleConditional(ts *tokenStream) ([]token.Token, error) {
	directive := ts.advance().Value // consume .ifdef or .ifndef

	// Get the name
	if !ts.check(token.IDENT) {
		return nil, fmt.Errorf("expected identifier after %s at line %d", directive, ts.peek().Row)
	}
	name := ts.advance().Value

	// Skip EOL
	if ts.check(token.EOL) {
		ts.advance()
	}

	// Check condition
	_, isDefined := p.defines[name]
	condition := isDefined
	if directive == ".ifndef" {
		condition = !condition
	}

	// Collect true and false blocks
	trueBlock, falseBlock, err := p.collectConditionalBlocks(ts)
	if err != nil {
		return nil, err
	}

	// Return appropriate block
	if condition {
		return trueBlock, nil
	}
	return falseBlock, nil
}

// handleIf processes .if EXPR ... .else ... .endif
func (p *Preprocessor) handleIf(ts *tokenStream) ([]token.Token, error) {
	ts.advance() // consume .if

	// Collect expression until EOL
	var exprToks []token.Token
	for !ts.atEnd() && ts.peek().Type != token.EOL && ts.peek().Type != token.EOF {
		exprToks = append(exprToks, ts.advance())
	}

	// Skip EOL
	if ts.check(token.EOL) {
		ts.advance()
	}

	// Evaluate condition
	value, err := p.evaluateExpr(exprToks)
	if err != nil {
		return nil, fmt.Errorf("error evaluating .if expression: %v", err)
	}
	condition := value != 0

	// Collect true and false blocks
	trueBlock, falseBlock, err := p.collectConditionalBlocks(ts)
	if err != nil {
		return nil, err
	}

	// Return appropriate block
	if condition {
		return trueBlock, nil
	}
	return falseBlock, nil
}

// collectConditionalBlocks collects tokens for true and false blocks
func (p *Preprocessor) collectConditionalBlocks(ts *tokenStream) ([]token.Token, []token.Token, error) {
	var trueBlock []token.Token
	var falseBlock []token.Token
	inFalseBlock := false
	depth := 1

	for !ts.atEnd() && depth > 0 {
		tok := ts.peek()

		if tok.Type == token.EOF {
			break
		}

		if tok.Type == token.DIRECTIVE {
			switch tok.Value {
			case ".if", ".ifdef", ".ifndef":
				depth++
				if inFalseBlock {
					falseBlock = append(falseBlock, ts.advance())
				} else {
					trueBlock = append(trueBlock, ts.advance())
				}
				continue

			case ".else":
				if depth == 1 {
					ts.advance() // consume .else
					// Skip EOL after .else
					if ts.check(token.EOL) {
						ts.advance()
					}
					inFalseBlock = true
					continue
				}

			case ".endif":
				depth--
				if depth == 0 {
					ts.advance() // consume .endif
					// Skip EOL after .endif
					if ts.check(token.EOL) {
						ts.advance()
					}
					continue
				}
			}
		}

		// Add token to appropriate block
		if inFalseBlock {
			falseBlock = append(falseBlock, ts.advance())
		} else {
			trueBlock = append(trueBlock, ts.advance())
		}
	}

	if depth != 0 {
		return nil, nil, fmt.Errorf("unterminated conditional block")
	}

	return trueBlock, falseBlock, nil
}
