package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseNumber parses a number string into an int64. A '_' can separate two
// digits (0x0020_6800, 1_000_000); it cannot start, end or follow another '_'.
func ParseNumber(s string) (int64, error) {
	s = strings.ToLower(s)

	base, digits := 10, s
	switch {
	case strings.HasPrefix(s, "0x"):
		base, digits = 16, s[2:]
	case strings.HasPrefix(s, "0b"):
		base, digits = 2, s[2:]
	case strings.HasPrefix(s, "0o"):
		base, digits = 8, s[2:]
	}
	if strings.Contains(digits, "_") {
		if strings.HasPrefix(digits, "_") || strings.HasSuffix(digits, "_") || strings.Contains(digits, "__") {
			return 0, fmt.Errorf("a '_' must be between two digits")
		}
		digits = strings.ReplaceAll(digits, "_", "")
	}
	val, err := strconv.ParseInt(digits, base, 64)
	if err != nil {
		if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
			return 0, fmt.Errorf("does not fit 64 bits")
		}
		return 0, fmt.Errorf("not a base-%d number", base)
	}
	return val, nil
}

// Unescape translates the escapes of a string or character literal (without
// its quotes): \n \t \r \0 \\ \' \" and \xHH. The result is bytes.
func Unescape(s string) (string, error) {
	if !strings.Contains(s, "\\") {
		return s, nil
	}
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			sb.WriteByte(s[i])
			continue
		}
		if i+1 >= len(s) {
			return "", fmt.Errorf("'\\' at the end of the text")
		}
		i++
		switch s[i] {
		case 'n':
			sb.WriteByte('\n')
		case 't':
			sb.WriteByte('\t')
		case 'r':
			sb.WriteByte('\r')
		case '0':
			sb.WriteByte(0)
		case '\\', '\'', '"':
			sb.WriteByte(s[i])
		case 'x':
			if i+2 >= len(s) {
				return "", fmt.Errorf("escape '\\x' needs 2 hex digits")
			}
			v, err := strconv.ParseUint(s[i+1:i+3], 16, 8)
			if err != nil {
				return "", fmt.Errorf("escape '\\x' needs 2 hex digits")
			}
			sb.WriteByte(byte(v))
			i += 2
		default:
			return "", fmt.Errorf("unknown escape '\\%c'", s[i])
		}
	}
	return sb.String(), nil
}

// CharValue returns the value of a character literal token ('A', '\\n', '\\x41')
func CharValue(lit string) (int64, error) {
	s := lit
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		s = s[1 : len(s)-1]
	}
	text, err := Unescape(s)
	if err != nil {
		return 0, err
	}
	if len(text) != 1 {
		return 0, fmt.Errorf("character %s must be one character", lit)
	}
	return int64(text[0]), nil
}
