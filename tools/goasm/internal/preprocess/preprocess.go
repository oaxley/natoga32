// Preprocessor for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package preprocess

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/natoga32/goasm/internal/token"
)

// Macro represents a defined macro
type Macro struct {
	Name   string        // Macro name
	Params []string      // Parameter names
	Body   []token.Token // Body tokens
}

// Preprocessor handles preprocessing directives
type Preprocessor struct {
	defines  map[string]int64  // Defined constants
	macros   map[string]*Macro // Defined macros
	includes map[string]bool   // Track included files to prevent recursion
	baseDir  string            // Base directory for resolving includes
}

// New creates a new preprocessor
func New(baseDir string) *Preprocessor {
	return &Preprocessor{
		defines:  make(map[string]int64),
		macros:   make(map[string]*Macro),
		includes: make(map[string]bool),
		baseDir:  baseDir,
	}
}

// Process preprocesses a list of tokens
func (p *Preprocessor) Process(tokens []token.Token) ([]token.Token, error) {
	return p.processTokens(tokens)
}

// processTokens processes tokens recursively
func (p *Preprocessor) processTokens(tokens []token.Token) ([]token.Token, error) {
	result := make([]token.Token, 0, len(tokens))
	needsReprocess := false

	ts := newTokenStream(tokens)

	for !ts.atEnd() {
		tok := ts.peek()

		if tok.Type == token.EOF {
			break
		}

		// Handle directives
		if tok.Type == token.DIRECTIVE {
			switch tok.Value {
			case ".define":
				if err := p.handleDefine(ts); err != nil {
					return nil, err
				}
				continue

			case ".include":
				included, err := p.handleInclude(ts)
				if err != nil {
					return nil, err
				}
				result = append(result, included...)
				// handleInclude now processes tokens internally, no reprocess needed
				continue

			case ".ifdef", ".ifndef":
				conditional, err := p.handleConditional(ts)
				if err != nil {
					return nil, err
				}
				result = append(result, conditional...)
				needsReprocess = true
				continue

			case ".if":
				conditional, err := p.handleIf(ts)
				if err != nil {
					return nil, err
				}
				result = append(result, conditional...)
				needsReprocess = true
				continue

			case ".dup":
				expanded, err := p.handleDup(ts)
				if err != nil {
					return nil, err
				}
				result = append(result, expanded...)
				continue

			case ".for":
				expanded, err := p.handleFor(ts)
				if err != nil {
					return nil, err
				}
				result = append(result, expanded...)
				needsReprocess = true // Body may contain directives that need processing
				continue

			case ".macro":
				if err := p.handleMacro(ts); err != nil {
					return nil, err
				}
				continue
			}
		}

		// Handle token pasting (##)
		if tok.Type == token.PASTE {
			pasted, err := p.handlePaste(ts, &result)
			if err != nil {
				return nil, err
			}
			result = append(result, pasted)
			continue
		}

		// Expand environment variables
		if tok.Type == token.ENVVAR {
			envValue := os.Getenv(tok.Value)
			if envValue != "" {
				// Try to parse as number first
				if val, err := parseNumber(envValue); err == nil {
					result = append(result, token.Token{
						Type:  token.NUMBER,
						Value: fmt.Sprintf("%d", val),
						Row:   tok.Row,
						Col:   tok.Col,
					})
				} else {
					// Treat as string literal
					result = append(result, token.Token{
						Type:  token.STRING,
						Value: "\"" + envValue + "\"",
						Row:   tok.Row,
						Col:   tok.Col,
					})
				}
			}
			ts.advance()
			continue
		}

		// Expand defined symbols
		if tok.Type == token.IDENT {
			if val, ok := p.defines[tok.Value]; ok {
				// Replace identifier with its defined value
				result = append(result, token.Token{
					Type:  token.NUMBER,
					Value: fmt.Sprintf("%d", val),
					Row:   tok.Row,
					Col:   tok.Col,
				})
				ts.advance()
				continue
			}

			// Expand macros
			if macro, ok := p.macros[tok.Value]; ok {
				ts.advance() // consume macro name
				expanded, err := p.expandMacro(macro, ts)
				if err != nil {
					return nil, err
				}
				result = append(result, expanded...)
				needsReprocess = true // Expanded content may need processing
				continue
			}
		}

		// Regular token - pass through
		result = append(result, tok)
		ts.advance()
	}

	// Add EOF if not present
	if len(result) == 0 || result[len(result)-1].Type != token.EOF {
		result = append(result, token.Token{Type: token.EOF})
	}

	// Recursively process if we made changes
	if needsReprocess {
		return p.processTokens(result)
	}

	return result, nil
}

// evaluateExpr evaluates a simple expression
func (p *Preprocessor) evaluateExpr(tokens []token.Token) (int64, error) {
	if len(tokens) == 0 {
		return 0, nil
	}

	// Simple case: single token
	if len(tokens) == 1 {
		return p.evalToken(tokens[0])
	}

	// Handle unary minus at start
	i := 0
	negate := false
	if tokens[0].Type == token.MINUS {
		negate = true
		i = 1
	}

	if i >= len(tokens) {
		return 0, fmt.Errorf("expected value after unary minus")
	}

	// For now, evaluate left to right with basic operators
	result, err := p.evalToken(tokens[i])
	if err != nil {
		return 0, err
	}

	if negate {
		result = -result
	}

	i++
	for i < len(tokens) {
		if i+1 >= len(tokens) {
			break
		}

		opTok := tokens[i]
		valTok := tokens[i+1]

		val, err := p.evalToken(valTok)
		if err != nil {
			return 0, err
		}

		switch opTok.Type {
		case token.PLUS:
			result += val
		case token.MINUS:
			result -= val
		case token.STAR:
			result *= val
		case token.SLASH:
			if val != 0 {
				result /= val
			}
		case token.AND:
			result &= val
		case token.OR:
			result |= val
		case token.XOR:
			result ^= val
		case token.LSHIFT:
			result <<= uint(val)
		case token.RSHIFT:
			result >>= uint(val)
		case token.EQ:
			if result == val {
				result = 1
			} else {
				result = 0
			}
		case token.NEQ:
			if result != val {
				result = 1
			} else {
				result = 0
			}
		case token.LT:
			if result < val {
				result = 1
			} else {
				result = 0
			}
		case token.GT:
			if result > val {
				result = 1
			} else {
				result = 0
			}
		case token.LTE:
			if result <= val {
				result = 1
			} else {
				result = 0
			}
		case token.GTE:
			if result >= val {
				result = 1
			} else {
				result = 0
			}
		default:
			return 0, fmt.Errorf("unexpected operator: %s", opTok.Value)
		}

		i += 2
	}

	return result, nil
}

// evalToken evaluates a single token to a value
func (p *Preprocessor) evalToken(tok token.Token) (int64, error) {
	switch tok.Type {
	case token.NUMBER:
		return parseNumber(tok.Value)

	case token.IDENT:
		if val, ok := p.defines[tok.Value]; ok {
			return val, nil
		}
		return 0, nil // Undefined symbols are 0

	case token.ENVVAR:
		envValue := os.Getenv(tok.Value)
		if envValue != "" {
			return parseNumber(envValue)
		}
		return 0, nil // Undefined env var is 0

	case token.CHAR:
		// Handle character literals
		s := tok.Value
		if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
			s = s[1 : len(s)-1]
		}
		if len(s) > 0 {
			if s[0] == '\\' && len(s) > 1 {
				switch s[1] {
				case 'n':
					return int64('\n'), nil
				case 't':
					return int64('\t'), nil
				case 'r':
					return int64('\r'), nil
				case '0':
					return 0, nil
				default:
					return int64(s[1]), nil
				}
			}
			return int64(s[0]), nil
		}
		return 0, nil
	}

	return 0, fmt.Errorf("cannot evaluate token: %s", tok.Value)
}

// IsDefined checks if a symbol is defined
func (p *Preprocessor) IsDefined(name string) bool {
	_, ok := p.defines[name]
	return ok
}

// Define adds a define
func (p *Preprocessor) Define(name string, value int64) {
	p.defines[name] = value
}

//----- helper functions

// parseNumber parses a number string
func parseNumber(s string) (int64, error) {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "0x") {
		return strconv.ParseInt(s[2:], 16, 64)
	}
	if strings.HasPrefix(s, "0b") {
		return strconv.ParseInt(s[2:], 2, 64)
	}
	if strings.HasPrefix(s, "0o") {
		return strconv.ParseInt(s[2:], 8, 64)
	}
	return strconv.ParseInt(s, 10, 64)
}
