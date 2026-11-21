// Preprocessor for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package preprocess

import (
	"fmt"
	"os"

	"github.com/natoga32/goasm/internal/helpers"
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

	macroExpansions int // macro calls expanded so far, see maxMacroExpansions
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

// Limits on macro expansion, so that a recursive macro stops with an error.
// A pass expands one level of macros: maxMacroDepth is the number of nested
// levels, maxMacroExpansions the number of macro calls in a whole Process.
const (
	maxMacroDepth      = 1000
	maxMacroExpansions = 100000
)

// macroCall records a macro call, for the recursion error
type macroCall struct {
	name string
	tok  token.Token
}

// processTokens runs passes over tokens until no directive or macro is left to expand
func (p *Preprocessor) processTokens(tokens []token.Token) ([]token.Token, error) {
	depth := 0
	for {
		result, needsReprocess, call, err := p.processPass(tokens)
		if err != nil {
			return nil, err
		}
		if !needsReprocess {
			return result, nil
		}
		if call != nil {
			depth++
			if depth > maxMacroDepth {
				return nil, macroRecursionError(call, fmt.Sprintf("more than %d levels", maxMacroDepth))
			}
		}
		tokens = result
	}
}

// macroRecursionError reports a macro that expands without end
func macroRecursionError(call *macroCall, why string) error {
	loc := fmt.Sprintf("line %d", call.tok.Row)
	if call.tok.Filename != "" {
		loc = fmt.Sprintf("%s:%d", call.tok.Filename, call.tok.Row)
	}
	return fmt.Errorf("%s: error: macro '%s' expanded too deeply (%s): recursive macro without an end?",
		loc, call.name, why)
}

// processPass processes tokens once. It reports whether the result needs another
// pass, and the last macro it expanded (nil when none).
func (p *Preprocessor) processPass(tokens []token.Token) ([]token.Token, bool, *macroCall, error) {
	result := make([]token.Token, 0, len(tokens))
	needsReprocess := false
	var lastCall *macroCall

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
					return nil, false, nil, err
				}
				continue

			case ".include":
				included, err := p.handleInclude(ts)
				if err != nil {
					return nil, false, nil, err
				}
				result = append(result, included...)
				// handleInclude now processes tokens internally, no reprocess needed
				continue

			case ".ifdef", ".ifndef":
				conditional, err := p.handleConditional(ts)
				if err != nil {
					return nil, false, nil, err
				}
				result = append(result, conditional...)
				needsReprocess = true
				continue

			case ".if":
				conditional, err := p.handleIf(ts)
				if err != nil {
					return nil, false, nil, err
				}
				result = append(result, conditional...)
				needsReprocess = true
				continue

			case ".dup":
				expanded, err := p.handleDup(ts)
				if err != nil {
					return nil, false, nil, err
				}
				result = append(result, expanded...)
				continue

			case ".for":
				expanded, err := p.handleFor(ts)
				if err != nil {
					return nil, false, nil, err
				}
				result = append(result, expanded...)
				needsReprocess = true // Body may contain directives that need processing
				continue

			case ".macro":
				if err := p.handleMacro(ts); err != nil {
					return nil, false, nil, err
				}
				continue
			}
		}

		// Handle token pasting (##)
		if tok.Type == token.PASTE {
			pasted, err := p.handlePaste(ts, &result)
			if err != nil {
				return nil, false, nil, err
			}
			result = append(result, pasted)
			continue
		}

		// Expand environment variables
		if tok.Type == token.ENVVAR {
			envValue := os.Getenv(tok.Value)
			if envValue != "" {
				// Try to parse as number first
				if val, err := helpers.ParseNumber(envValue); err == nil {
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
				p.macroExpansions++
				lastCall = &macroCall{name: tok.Value, tok: tok}
				if p.macroExpansions > maxMacroExpansions {
					return nil, false, nil, macroRecursionError(lastCall,
						fmt.Sprintf("more than %d calls", maxMacroExpansions))
				}
				ts.advance() // consume macro name
				expanded, err := p.expandMacro(macro, ts)
				if err != nil {
					return nil, false, nil, err
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

	return result, needsReprocess, lastCall, nil
}

// evaluateExpr evaluates a preprocessor expression (.if, .define, .dup, .for)
// with the C rules: unary ! - + ~, then * / %, + -, << >>, < <= > >=, == !=,
// &, ^, |, && and || (both short-circuit), and parentheses. An empty
// expression is 0. A name that is not a .define name is 0.
func (p *Preprocessor) evaluateExpr(tokens []token.Token) (int64, error) {
	if len(tokens) == 0 {
		return 0, nil
	}
	e := &exprEval{p: p, toks: tokens}
	val, err := e.binary(0, true)
	if err != nil {
		return 0, err
	}
	if e.pos < len(e.toks) {
		tok := e.toks[e.pos]
		if tok.Type == token.RPAREN {
			return 0, fmt.Errorf("unexpected ')'")
		}
		return 0, fmt.Errorf("unexpected '%s' after the expression", tokenText(tok))
	}
	return val, nil
}

// exprEval walks the tokens of one preprocessor expression
type exprEval struct {
	p    *Preprocessor
	toks []token.Token
	pos  int
}

// binaryPrec gives the precedence of each binary operator (higher binds tighter)
var binaryPrec = map[token.Type]int{
	token.LOR: 1, token.LAND: 2,
	token.OR: 3, token.XOR: 4, token.AND: 5,
	token.EQ: 6, token.NEQ: 6,
	token.LT: 7, token.LTE: 7, token.GT: 7, token.GTE: 7,
	token.LSHIFT: 8, token.RSHIFT: 8,
	token.PLUS: 9, token.MINUS: 9,
	token.STAR: 10, token.SLASH: 10, token.MODULO: 10,
}

// binary parses operators of precedence above minPrec. When eval is false
// (the skipped side of && or ||) the value is not computed, so 0 && 1/0 is 0.
func (e *exprEval) binary(minPrec int, eval bool) (int64, error) {
	left, err := e.unary(eval)
	if err != nil {
		return 0, err
	}
	for e.pos < len(e.toks) {
		op := e.toks[e.pos]
		prec, ok := binaryPrec[op.Type]
		if !ok || prec <= minPrec {
			break
		}
		e.pos++
		evalRight := eval
		if op.Type == token.LAND && left == 0 || op.Type == token.LOR && left != 0 {
			evalRight = false
		}
		right, err := e.binary(prec, evalRight)
		if err != nil {
			return 0, err
		}
		if !eval {
			continue
		}
		if left, err = applyBinary(op, left, right); err != nil {
			return 0, err
		}
	}
	return left, nil
}

func boolValue(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// applyBinary computes "left op right"
func applyBinary(op token.Token, left, right int64) (int64, error) {
	switch op.Type {
	case token.PLUS:
		return left + right, nil
	case token.MINUS:
		return left - right, nil
	case token.STAR:
		return left * right, nil
	case token.SLASH:
		if right == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return left / right, nil
	case token.MODULO:
		if right == 0 {
			return 0, fmt.Errorf("modulo by zero")
		}
		return left % right, nil
	case token.LSHIFT, token.RSHIFT:
		if right < 0 || right > 63 {
			return 0, fmt.Errorf("shift amount %d out of range (0..63)", right)
		}
		if op.Type == token.LSHIFT {
			return left << uint(right), nil
		}
		return left >> uint(right), nil
	case token.AND:
		return left & right, nil
	case token.OR:
		return left | right, nil
	case token.XOR:
		return left ^ right, nil
	case token.EQ:
		return boolValue(left == right), nil
	case token.NEQ:
		return boolValue(left != right), nil
	case token.LT:
		return boolValue(left < right), nil
	case token.LTE:
		return boolValue(left <= right), nil
	case token.GT:
		return boolValue(left > right), nil
	case token.GTE:
		return boolValue(left >= right), nil
	case token.LAND:
		return boolValue(left != 0 && right != 0), nil
	case token.LOR:
		return boolValue(left != 0 || right != 0), nil
	}
	return 0, fmt.Errorf("unknown operator '%s'", tokenText(op))
}

// unary parses a value with its prefix operators: ! - + ~, ( ... ), a token
func (e *exprEval) unary(eval bool) (int64, error) {
	if e.pos >= len(e.toks) {
		return 0, fmt.Errorf("expected a value at the end of the expression")
	}
	tok := e.toks[e.pos]
	e.pos++
	switch tok.Type {
	case token.NOT, token.MINUS, token.PLUS, token.TILDE:
		val, err := e.unary(eval)
		if err != nil {
			return 0, err
		}
		switch tok.Type {
		case token.NOT:
			return boolValue(val == 0), nil
		case token.MINUS:
			return -val, nil
		case token.TILDE:
			return ^val, nil
		}
		return val, nil
	case token.LPAREN:
		val, err := e.binary(0, eval)
		if err != nil {
			return 0, err
		}
		if e.pos >= len(e.toks) || e.toks[e.pos].Type != token.RPAREN {
			return 0, fmt.Errorf("missing ')'")
		}
		e.pos++
		return val, nil
	case token.NUMBER, token.IDENT, token.ENVVAR, token.CHAR:
		return e.p.evalToken(tok)
	}
	return 0, fmt.Errorf("expected a value, got '%s'", tokenText(tok))
}

// tokenText is the source text of a token, for error messages
func tokenText(tok token.Token) string {
	if tok.Value != "" {
		return tok.Value
	}
	return tok.Type.String()
}

// locationError prefixes an error with the file and line of tok
func locationError(tok token.Token, format string, args ...any) error {
	loc := fmt.Sprintf("line %d", tok.Row)
	if tok.Filename != "" {
		loc = fmt.Sprintf("%s:%d", tok.Filename, tok.Row)
	}
	return fmt.Errorf("%s: error: %s", loc, fmt.Sprintf(format, args...))
}

// evalToken evaluates a single token to a value
func (p *Preprocessor) evalToken(tok token.Token) (int64, error) {
	switch tok.Type {
	case token.NUMBER:
		return helpers.ParseNumber(tok.Value)

	case token.IDENT:
		if val, ok := p.defines[tok.Value]; ok {
			return val, nil
		}
		return 0, nil // Undefined symbols are 0

	case token.ENVVAR:
		envValue := os.Getenv(tok.Value)
		if envValue != "" {
			return helpers.ParseNumber(envValue)
		}
		return 0, nil // Undefined env var is 0

	case token.CHAR:
		return helpers.CharValue(tok.Value)
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
