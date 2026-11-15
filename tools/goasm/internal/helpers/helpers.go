package helpers

import (
	"strconv"
	"strings"
)

// ParseNumber parses a number string into an int64
func ParseNumber(s string) (int64, error) {
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
