package preprocess

import (
	"github.com/natoga32/goasm/internal/token"
)

// tokenStream is a simple wrapper for iterating over tokens
type tokenStream struct {
	tokens []token.Token
	pos    int
}

func newTokenStream(tokens []token.Token) *tokenStream {
	return &tokenStream{tokens: tokens, pos: 0}
}

func (ts *tokenStream) atEnd() bool {
	return ts.pos >= len(ts.tokens)
}

func (ts *tokenStream) peek() token.Token {
	if ts.atEnd() {
		return token.Token{Type: token.EOF}
	}
	return ts.tokens[ts.pos]
}

func (ts *tokenStream) advance() token.Token {
	tok := ts.peek()
	if !ts.atEnd() {
		ts.pos++
	}
	return tok
}

func (ts *tokenStream) check(typ token.Type) bool {
	return ts.peek().Type == typ
}
