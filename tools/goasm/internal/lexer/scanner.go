package lexer

import (
	"github.com/natoga32/goasm/internal/token"
)

// scanToken scans the next token
func (l *Lexer) scanToken() {
	startRow := l.row
	startCol := l.col
	ch := l.advance()

	switch ch {
	case '\n':
		// Only add EOL if we have tokens on this line
		if len(l.tokens) > 0 && l.tokens[len(l.tokens)-1].Type != token.EOL {
			l.addToken(token.EOL, "", startRow, startCol)
		}

	case ' ', '\t', '\r':
		// Skip whitespace

	case ';':
		// Comment - skip to end of line
		for !l.atEnd() && l.peek() != '\n' {
			l.advance()
		}

	case '#':
		if l.match('#') {
			// Token pasting operator
			l.addToken(token.PASTE, "##", startRow, startCol)
		} else {
			// Comment - skip to end of line
			for !l.atEnd() && l.peek() != '\n' {
				l.advance()
			}
		}

	case '(':
		l.addToken(token.LPAREN, "(", startRow, startCol)

	case ')':
		l.addToken(token.RPAREN, ")", startRow, startCol)

	case ',':
		l.addToken(token.COMMA, ",", startRow, startCol)

	case '+':
		l.addToken(token.PLUS, "+", startRow, startCol)

	case '-':
		l.addToken(token.MINUS, "-", startRow, startCol)

	case '*':
		l.addToken(token.STAR, "*", startRow, startCol)

	case '/':
		if l.match('/') {
			// C++ style comment
			for !l.atEnd() && l.peek() != '\n' {
				l.advance()
			}
		} else {
			l.addToken(token.SLASH, "/", startRow, startCol)
		}

	case '%':
		// Could be modulo or relocation modifier
		if isAlpha(l.peek()) {
			l.scanRelocation(startRow, startCol)
		} else {
			l.addToken(token.MODULO, "%", startRow, startCol)
		}

	case '&':
		l.addToken(token.AND, "&", startRow, startCol)

	case '|':
		l.addToken(token.OR, "|", startRow, startCol)

	case '^':
		l.addToken(token.XOR, "^", startRow, startCol)

	case '$':
		if l.peek() == '[' {
			// Environment variable: $[VARNAME]
			l.scanEnvVar(startRow, startCol)
		} else {
			l.addToken(token.DOLLAR, "$", startRow, startCol)
		}

	case '=':
		if l.match('=') {
			l.addToken(token.EQ, "==", startRow, startCol)
		} else {
			l.addToken(token.ASSIGN, "=", startRow, startCol)
		}

	case '!':
		if l.match('=') {
			l.addToken(token.NEQ, "!=", startRow, startCol)
		}

	case '<':
		if l.match('<') {
			l.addToken(token.LSHIFT, "<<", startRow, startCol)
		} else if l.match('=') {
			l.addToken(token.LTE, "<=", startRow, startCol)
		} else {
			l.addToken(token.LT, "<", startRow, startCol)
		}

	case '>':
		if l.match('>') {
			l.addToken(token.RSHIFT, ">>", startRow, startCol)
		} else if l.match('=') {
			l.addToken(token.GTE, ">=", startRow, startCol)
		} else {
			l.addToken(token.GT, ">", startRow, startCol)
		}

	case '"':
		l.scanString(startRow, startCol)

	case '\'':
		l.scanChar(startRow, startCol)

	case '.':
		l.scanDirective(startRow, startCol)

	case '\\':
		// Macro parameter reference: \name
		if isAlpha(l.peek()) {
			l.scanMacroParam(startRow, startCol)
		}

	default:
		if isDigit(ch) {
			l.scanNumber(ch, startRow, startCol)
		} else if isAlpha(ch) || ch == '_' {
			l.scanIdentifier(ch, startRow, startCol)
		}
	}
}
