package lexer

import (
	"unicode"
)

// Helper functions
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

func isAlpha(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || isDigit(ch) || ch == '.'
}
