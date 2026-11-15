package lexer

import (
	"testing"

	"github.com/natoga32/goasm/internal/token"
)

func TestLexerBasic(t *testing.T) {
	input := `
.text
main:
    addi x1, x0, 10
    add x2, x1, x0
`
	l := New(input)
	tokens := l.Tokenize()

	// Expected token types (simplified check)
	expected := []token.Type{
		token.DIRECTIVE, // .text
		token.EOL,
		token.LABEL, // main:
		token.EOL,
		token.IDENT, // addi
		token.IDENT, // x1
		token.COMMA,
		token.IDENT, // x0
		token.COMMA,
		token.NUMBER, // 10
		token.EOL,
		token.IDENT, // add
		token.IDENT, // x2
		token.COMMA,
		token.IDENT, // x1
		token.COMMA,
		token.IDENT, // x0
		token.EOL,
		token.EOF,
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		for i, tok := range tokens {
			t.Logf("Token %d: %v", i, tok)
		}
		return
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token %d: expected %v, got %v (%s)", i, expected[i], tok.Type, tok.Value)
		}
	}
}

func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"0x1A", "0x1A"},
		{"0b1010", "0b1010"},
		{"0o17", "0o17"},
	}

	for _, test := range tests {
		l := New(test.input)
		tokens := l.Tokenize()
		if len(tokens) < 1 {
			t.Errorf("No tokens for input %q", test.input)
			continue
		}
		if tokens[0].Type != token.NUMBER {
			t.Errorf("Expected NUMBER, got %v for input %q", tokens[0].Type, test.input)
		}
		if tokens[0].Value != test.expected {
			t.Errorf("Expected value %q, got %q", test.expected, tokens[0].Value)
		}
	}
}

func TestLexerRelocation(t *testing.T) {
	input := "%hi(symbol) %lo(symbol) %pcrel_hi(label) %pcrel_lo(label)"
	l := New(input)
	tokens := l.Tokenize()

	expected := []token.Type{
		token.ABS_HI,
		token.LPAREN,
		token.IDENT,
		token.RPAREN,
		token.ABS_LO,
		token.LPAREN,
		token.IDENT,
		token.RPAREN,
		token.PCREL_HI,
		token.LPAREN,
		token.IDENT,
		token.RPAREN,
		token.PCREL_LO,
		token.LPAREN,
		token.IDENT,
		token.RPAREN,
		token.EOF,
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		for i, tok := range tokens {
			t.Logf("Token %d: %v", i, tok)
		}
		return
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token %d: expected %v, got %v", i, expected[i], tok.Type)
		}
	}
}

func TestLexerDirectives(t *testing.T) {
	input := ".text .data .bss .cpu .define"
	l := New(input)
	tokens := l.Tokenize()

	for i := 0; i < 5; i++ {
		if tokens[i].Type != token.DIRECTIVE {
			t.Errorf("Token %d: expected DIRECTIVE, got %v", i, tokens[i].Type)
		}
	}
}

func TestLexerString(t *testing.T) {
	input := `"hello world"`
	l := New(input)
	tokens := l.Tokenize()

	if tokens[0].Type != token.STRING {
		t.Errorf("Expected STRING, got %v", tokens[0].Type)
	}
	if tokens[0].Value != `"hello world"` {
		t.Errorf("Expected %q, got %q", `"hello world"`, tokens[0].Value)
	}
}

func TestLexerOperators(t *testing.T) {
	input := "+ - * / % << >> & | ^ == != < > <= >="
	l := New(input)
	tokens := l.Tokenize()

	expected := []token.Type{
		token.PLUS,
		token.MINUS,
		token.STAR,
		token.SLASH,
		token.MODULO,
		token.LSHIFT,
		token.RSHIFT,
		token.AND,
		token.OR,
		token.XOR,
		token.EQ,
		token.NEQ,
		token.LT,
		token.GT,
		token.LTE,
		token.GTE,
		token.EOF,
	}

	for i, exp := range expected {
		if i >= len(tokens) {
			t.Errorf("Missing token %d: expected %v", i, exp)
			continue
		}
		if tokens[i].Type != exp {
			t.Errorf("Token %d: expected %v, got %v", i, exp, tokens[i].Type)
		}
	}
}

func TestTokenPaste(t *testing.T) {
	// Test that ## is recognized as PASTE token
	input := "x##i"
	l := New(input)
	tokens := l.Tokenize()

	expected := []token.Type{
		token.IDENT, // x
		token.PASTE, // ##
		token.IDENT, // i
		token.EOF,
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		for i, tok := range tokens {
			t.Logf("Token %d: %v", i, tok)
		}
		return
	}

	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("Token %d: expected %v, got %v", i, exp, tokens[i].Type)
		}
	}

	// Verify values
	if tokens[0].Value != "x" {
		t.Errorf("Expected 'x', got '%s'", tokens[0].Value)
	}
	if tokens[1].Value != "##" {
		t.Errorf("Expected '##', got '%s'", tokens[1].Value)
	}
	if tokens[2].Value != "i" {
		t.Errorf("Expected 'i', got '%s'", tokens[2].Value)
	}
}

func TestTokenPasteWithComment(t *testing.T) {
	// Test that # alone still starts a comment
	input := "x # this is a comment"
	l := New(input)
	tokens := l.Tokenize()

	expected := []token.Type{
		token.IDENT, // x
		token.EOF,
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		for i, tok := range tokens {
			t.Logf("Token %d: %v", i, tok)
		}
		return
	}

	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("Token %d: expected %v, got %v", i, exp, tokens[i].Type)
		}
	}
}

func TestTokenPasteWithLabel(t *testing.T) {
	// Test handler##n:
	input := "handler##n:"
	l := New(input)
	tokens := l.Tokenize()

	expected := []token.Type{
		token.IDENT, // handler
		token.PASTE, // ##
		token.LABEL, // n:
		token.EOF,
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		for i, tok := range tokens {
			t.Logf("Token %d: %v", i, tok)
		}
		return
	}

	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("Token %d: expected %v, got %v", i, exp, tokens[i].Type)
		}
	}
}

func TestEnvVar(t *testing.T) {
	// Test $[VARNAME] environment variable syntax
	input := "$[HOME]"
	l := New(input)
	tokens := l.Tokenize()

	expected := []token.Type{
		token.ENVVAR, // HOME
		token.EOF,
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		for i, tok := range tokens {
			t.Logf("Token %d: %v", i, tok)
		}
		return
	}

	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("Token %d: expected %v, got %v", i, exp, tokens[i].Type)
		}
	}

	// Verify value is the variable name without brackets
	if tokens[0].Value != "HOME" {
		t.Errorf("Expected 'HOME', got '%s'", tokens[0].Value)
	}
}

func TestEnvVarInExpression(t *testing.T) {
	// Test environment variable in an expression
	input := ".define SIZE $[BUFFER_SIZE]"
	l := New(input)
	tokens := l.Tokenize()

	expected := []token.Type{
		token.DIRECTIVE, // .define
		token.IDENT,     // SIZE
		token.ENVVAR,    // BUFFER_SIZE
		token.EOF,
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		for i, tok := range tokens {
			t.Logf("Token %d: %v", i, tok)
		}
		return
	}

	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("Token %d: expected %v, got %v", i, exp, tokens[i].Type)
		}
	}
}

func TestDollarWithoutBracket(t *testing.T) {
	// Test that $ alone is still DOLLAR (current PC)
	input := "$ + 4"
	l := New(input)
	tokens := l.Tokenize()

	expected := []token.Type{
		token.DOLLAR, // $
		token.PLUS,   // +
		token.NUMBER, // 4
		token.EOF,
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		for i, tok := range tokens {
			t.Logf("Token %d: %v", i, tok)
		}
		return
	}

	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("Token %d: expected %v, got %v", i, exp, tokens[i].Type)
		}
	}
}
