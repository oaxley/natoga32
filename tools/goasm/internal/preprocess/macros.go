package preprocess

import (
	"fmt"

	"github.com/natoga32/goasm/internal/token"
)

// handleMacro processes .macro NAME [params...] ... .endmacro
func (p *Preprocessor) handleMacro(ts *tokenStream) error {
	row := ts.peek().Row
	ts.advance() // consume .macro

	// Parse macro name
	if !ts.check(token.IDENT) {
		return fmt.Errorf("expected macro name after .macro at line %d", row)
	}
	name := ts.advance().Value

	// Parse optional parameters
	var params []string
	for !ts.atEnd() && ts.peek().Type != token.EOL && ts.peek().Type != token.EOF {
		if ts.check(token.IDENT) {
			params = append(params, ts.advance().Value)
		} else if ts.check(token.COMMA) {
			ts.advance() // skip comma between params
		} else {
			break
		}
	}

	// Skip EOL after .macro line
	if ts.check(token.EOL) {
		ts.advance()
	}

	// Collect macro body until .endmacro
	body, err := p.collectMacroBody(ts)
	if err != nil {
		return err
	}

	// Store the macro
	p.macros[name] = &Macro{
		Name:   name,
		Params: params,
		Body:   body,
	}

	return nil
}

// collectMacroBody collects tokens between .macro and .endmacro, handling nesting
func (p *Preprocessor) collectMacroBody(ts *tokenStream) ([]token.Token, error) {
	var body []token.Token
	depth := 1

	for !ts.atEnd() && depth > 0 {
		tok := ts.peek()

		if tok.Type == token.EOF {
			break
		}

		if tok.Type == token.DIRECTIVE {
			switch tok.Value {
			case ".macro":
				depth++
			case ".endmacro", ".endm":
				depth--
				if depth == 0 {
					ts.advance() // consume .endmacro
					// Skip EOL after .endmacro
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
		return nil, fmt.Errorf("unterminated .macro (missing .endmacro)")
	}

	return body, nil
}

// expandMacro expands a macro invocation with the given arguments
func (p *Preprocessor) expandMacro(macro *Macro, ts *tokenStream) ([]token.Token, error) {
	// Collect arguments (tokens until EOL)
	var args [][]token.Token
	var currentArg []token.Token

	for !ts.atEnd() && ts.peek().Type != token.EOL && ts.peek().Type != token.EOF {
		tok := ts.peek()
		if tok.Type == token.COMMA {
			ts.advance()
			if len(currentArg) > 0 {
				args = append(args, currentArg)
				currentArg = nil
			}
		} else {
			currentArg = append(currentArg, ts.advance())
		}
	}
	if len(currentArg) > 0 {
		args = append(args, currentArg)
	}

	// Skip EOL
	if ts.check(token.EOL) {
		ts.advance()
	}

	// Build parameter map
	paramMap := make(map[string][]token.Token)
	for i, param := range macro.Params {
		if i < len(args) {
			paramMap[param] = args[i]
		} else {
			paramMap[param] = nil // Missing argument
		}
	}

	// Expand the body, substituting parameters
	var result []token.Token
	for i := 0; i < len(macro.Body); i++ {
		tok := macro.Body[i]

		// Check for parameter reference: \param (MACRO_PARAM token)
		if tok.Type == token.MACRO_PARAM {
			paramName := tok.Value
			if replacement, ok := paramMap[paramName]; ok && len(replacement) > 0 {
				result = append(result, replacement...)
			}
			continue
		}

		result = append(result, tok)
	}

	return result, nil
}
