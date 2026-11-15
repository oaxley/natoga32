package lexer

import (
	"strings"

	"github.com/natoga32/goasm/internal/token"
)

// scanNumber scans a number literal
func (l *Lexer) scanNumber(first byte, row, col int) {
	var sb strings.Builder
	sb.WriteByte(first)

	// Check for hex (0x), binary (0b), or octal (0o)
	if first == '0' && !l.atEnd() {
		next := l.peek()
		if next == 'x' || next == 'X' {
			sb.WriteByte(l.advance())
			for !l.atEnd() && isHexDigit(l.peek()) {
				sb.WriteByte(l.advance())
			}
			l.addToken(token.NUMBER, sb.String(), row, col)
			return
		} else if next == 'b' || next == 'B' {
			sb.WriteByte(l.advance())
			for !l.atEnd() && (l.peek() == '0' || l.peek() == '1') {
				sb.WriteByte(l.advance())
			}
			l.addToken(token.NUMBER, sb.String(), row, col)
			return
		} else if next == 'o' || next == 'O' {
			sb.WriteByte(l.advance())
			for !l.atEnd() && isOctalDigit(l.peek()) {
				sb.WriteByte(l.advance())
			}
			l.addToken(token.NUMBER, sb.String(), row, col)
			return
		}
	}

	// Decimal number
	for !l.atEnd() && isDigit(l.peek()) {
		sb.WriteByte(l.advance())
	}

	l.addToken(token.NUMBER, sb.String(), row, col)
}

// scanIdentifier scans an identifier or label
func (l *Lexer) scanIdentifier(first byte, row, col int) {
	var sb strings.Builder
	sb.WriteByte(first)

	for !l.atEnd() && isAlphaNumeric(l.peek()) {
		sb.WriteByte(l.advance())
	}

	value := sb.String()

	// Check if it's a label (ends with ':')
	if !l.atEnd() && l.peek() == ':' {
		l.advance()
		l.addToken(token.LABEL, value+":", row, col)
	} else {
		l.addToken(token.IDENT, value, row, col)
	}
}

// scanDirective scans a directive (starts with '.')
func (l *Lexer) scanDirective(row, col int) {
	var sb strings.Builder
	sb.WriteByte('.')

	for !l.atEnd() && isAlphaNumeric(l.peek()) {
		sb.WriteByte(l.advance())
	}

	l.addToken(token.DIRECTIVE, sb.String(), row, col)
}

// scanString scans a string literal
func (l *Lexer) scanString(row, col int) {
	var sb strings.Builder
	sb.WriteByte('"')

	for !l.atEnd() && l.peek() != '"' && l.peek() != '\n' {
		ch := l.advance()
		sb.WriteByte(ch)
		// Handle escape sequences
		if ch == '\\' && !l.atEnd() {
			sb.WriteByte(l.advance())
		}
	}

	if !l.atEnd() && l.peek() == '"' {
		sb.WriteByte(l.advance())
	}

	l.addToken(token.STRING, sb.String(), row, col)
}

// scanChar scans a character literal
func (l *Lexer) scanChar(row, col int) {
	var sb strings.Builder
	sb.WriteByte('\'')

	for !l.atEnd() && l.peek() != '\'' && l.peek() != '\n' {
		ch := l.advance()
		sb.WriteByte(ch)
		if ch == '\\' && !l.atEnd() {
			sb.WriteByte(l.advance())
		}
	}

	if !l.atEnd() && l.peek() == '\'' {
		sb.WriteByte(l.advance())
	}

	l.addToken(token.CHAR, sb.String(), row, col)
}

// scanMacroParam scans a macro parameter reference (\name)
func (l *Lexer) scanMacroParam(row, col int) {
	var sb strings.Builder

	for !l.atEnd() && isAlphaNumeric(l.peek()) {
		sb.WriteByte(l.advance())
	}

	l.addToken(token.MACRO_PARAM, sb.String(), row, col)
}

// scanEnvVar scans an environment variable reference: $[VARNAME]
func (l *Lexer) scanEnvVar(row, col int) {
	// Skip the opening '['
	l.advance()

	var sb strings.Builder
	for !l.atEnd() && l.peek() != ']' && l.peek() != '\n' {
		sb.WriteByte(l.advance())
	}

	// Skip the closing ']' if present
	if !l.atEnd() && l.peek() == ']' {
		l.advance()
	}

	l.addToken(token.ENVVAR, sb.String(), row, col)
}

// scanRelocation scans a relocation modifier (%lo, %hi, %pcrel_lo, %pcrel_hi)
func (l *Lexer) scanRelocation(row, col int) {
	var sb strings.Builder
	sb.WriteByte('%')

	for !l.atEnd() && (isAlphaNumeric(l.peek()) || l.peek() == '_') {
		sb.WriteByte(l.advance())
	}

	value := sb.String()
	switch strings.ToLower(value) {
	case "%lo":
		l.addToken(token.ABS_LO, value, row, col)
	case "%hi":
		l.addToken(token.ABS_HI, value, row, col)
	case "%pcrel_lo":
		l.addToken(token.PCREL_LO, value, row, col)
	case "%pcrel_hi":
		l.addToken(token.PCREL_HI, value, row, col)
	default:
		// Unknown relocation, treat as modulo followed by identifier
		l.addToken(token.MODULO, "%", row, col)
		// Re-scan the identifier part
		if len(value) > 1 {
			l.addToken(token.IDENT, value[1:], row, col+1)
		}
	}
}
