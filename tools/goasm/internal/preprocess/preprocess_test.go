package preprocess

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/natoga32/goasm/internal/token"
)

func TestDefine(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "FOO", Row: 1},
		{Type: token.NUMBER, Value: "42", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "FOO", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: NUMBER(42), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Type != token.NUMBER {
		t.Errorf("expected NUMBER, got %v", result[0].Type)
	}
	if result[0].Value != "42" {
		t.Errorf("expected '42', got '%s'", result[0].Value)
	}
}

func TestDefineWithExpression(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "A", Row: 1},
		{Type: token.NUMBER, Value: "10", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".define", Row: 2},
		{Type: token.IDENT, Value: "B", Row: 2},
		{Type: token.IDENT, Value: "A", Row: 2},
		{Type: token.PLUS, Value: "+", Row: 2},
		{Type: token.NUMBER, Value: "5", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.IDENT, Value: "B", Row: 3},
		{Type: token.EOF, Row: 3},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: NUMBER(15), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Value != "15" {
		t.Errorf("expected '15', got '%s'", result[0].Value)
	}
}

func TestIfdefTrue(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "DEBUG", Row: 1},
		{Type: token.NUMBER, Value: "1", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".ifdef", Row: 2},
		{Type: token.IDENT, Value: "DEBUG", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.IDENT, Value: "included", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.DIRECTIVE, Value: ".endif", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.EOF, Row: 5},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: IDENT(included), EOL, EOF
	found := false
	for _, tok := range result {
		if tok.Type == token.IDENT && tok.Value == "included" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'included' token in output")
	}
}

func TestIfdefFalse(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".ifdef", Row: 1},
		{Type: token.IDENT, Value: "UNDEFINED", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "should_not_appear", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endif", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.EOF, Row: 4},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should only have EOF
	for _, tok := range result {
		if tok.Type == token.IDENT && tok.Value == "should_not_appear" {
			t.Error("unexpected 'should_not_appear' token in output")
		}
	}
}

func TestIfndefTrue(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".ifndef", Row: 1},
		{Type: token.IDENT, Value: "UNDEFINED", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "included", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endif", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.EOF, Row: 4},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have 'included'
	found := false
	for _, tok := range result {
		if tok.Type == token.IDENT && tok.Value == "included" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'included' token in output")
	}
}

func TestIfdefWithElse(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".ifdef", Row: 1},
		{Type: token.IDENT, Value: "UNDEFINED", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "true_branch", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".else", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.IDENT, Value: "false_branch", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.DIRECTIVE, Value: ".endif", Row: 5},
		{Type: token.EOL, Row: 5},
		{Type: token.EOF, Row: 6},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have 'false_branch', not 'true_branch'
	foundTrue := false
	foundFalse := false
	for _, tok := range result {
		if tok.Type == token.IDENT {
			if tok.Value == "true_branch" {
				foundTrue = true
			}
			if tok.Value == "false_branch" {
				foundFalse = true
			}
		}
	}
	if foundTrue {
		t.Error("unexpected 'true_branch' token in output")
	}
	if !foundFalse {
		t.Error("expected 'false_branch' token in output")
	}
}

func TestIfExpr(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "VERSION", Row: 1},
		{Type: token.NUMBER, Value: "2", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".if", Row: 2},
		{Type: token.IDENT, Value: "VERSION", Row: 2},
		{Type: token.GT, Value: ">", Row: 2},
		{Type: token.NUMBER, Value: "1", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.IDENT, Value: "v2_code", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.DIRECTIVE, Value: ".endif", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.EOF, Row: 5},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have 'v2_code'
	found := false
	for _, tok := range result {
		if tok.Type == token.IDENT && tok.Value == "v2_code" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'v2_code' token in output")
	}
}

func TestNestedConditionals(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "A", Row: 1},
		{Type: token.NUMBER, Value: "1", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".ifdef", Row: 2},
		{Type: token.IDENT, Value: "A", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".ifndef", Row: 3},
		{Type: token.IDENT, Value: "B", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.IDENT, Value: "nested", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.DIRECTIVE, Value: ".endif", Row: 5},
		{Type: token.EOL, Row: 5},
		{Type: token.DIRECTIVE, Value: ".endif", Row: 6},
		{Type: token.EOL, Row: 6},
		{Type: token.EOF, Row: 7},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have 'nested'
	found := false
	for _, tok := range result {
		if tok.Type == token.IDENT && tok.Value == "nested" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'nested' token in output")
	}
}

func TestInclude(t *testing.T) {
	// Create a temporary directory and include file
	tmpDir, err := os.MkdirTemp("", "preprocess_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an include file
	includeContent := "; included file\n.define INCLUDED_VAL 123\n"
	includeFile := filepath.Join(tmpDir, "include.asm")
	if err := os.WriteFile(includeFile, []byte(includeContent), 0644); err != nil {
		t.Fatalf("failed to write include file: %v", err)
	}

	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".include", Row: 1},
		{Type: token.STRING, Value: "\"include.asm\"", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "INCLUDED_VAL", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New(tmpDir)
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have expanded INCLUDED_VAL to 123
	found := false
	for _, tok := range result {
		if tok.Type == token.NUMBER && tok.Value == "123" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected INCLUDED_VAL to be expanded to 123")
	}
}

func TestRecursiveIncludeDetection(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "preprocess_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a self-including file
	includeContent := ".include \"recursive.asm\"\n"
	includeFile := filepath.Join(tmpDir, "recursive.asm")
	if err := os.WriteFile(includeFile, []byte(includeContent), 0644); err != nil {
		t.Fatalf("failed to write include file: %v", err)
	}

	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".include", Row: 1},
		{Type: token.STRING, Value: "\"recursive.asm\"", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.EOF, Row: 2},
	}

	p := New(tmpDir)
	_, err = p.Process(tokens)
	if err == nil {
		t.Error("expected error for recursive include")
	}
}

func TestUnterminatedConditional(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".ifdef", Row: 1},
		{Type: token.IDENT, Value: "FOO", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "stuff", Row: 2},
		{Type: token.EOF, Row: 3},
	}

	p := New("")
	_, err := p.Process(tokens)
	if err == nil {
		t.Error("expected error for unterminated conditional")
	}
}

func TestHexNumbers(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "ADDR", Row: 1},
		{Type: token.NUMBER, Value: "0x1000", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "ADDR", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: NUMBER(4096), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Type != token.NUMBER {
		t.Errorf("expected NUMBER, got %v", result[0].Type)
	}
	if result[0].Value != "4096" {
		t.Errorf("expected '4096', got '%s'", result[0].Value)
	}
}

func TestBitwiseOperations(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "MASK", Row: 1},
		{Type: token.NUMBER, Value: "0xFF", Row: 1},
		{Type: token.AND, Value: "&", Row: 1},
		{Type: token.NUMBER, Value: "0x0F", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "MASK", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// 0xFF & 0x0F = 0x0F = 15
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Value != "15" {
		t.Errorf("expected '15', got '%s'", result[0].Value)
	}
}

func TestCharLiteral(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "NEWLINE", Row: 1},
		{Type: token.CHAR, Value: "'\\n'", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "NEWLINE", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// '\n' = 10
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Value != "10" {
		t.Errorf("expected '10', got '%s'", result[0].Value)
	}
}

func TestDefineMethod(t *testing.T) {
	p := New("")
	p.Define("TEST", 100)

	if !p.IsDefined("TEST") {
		t.Error("expected TEST to be defined")
	}

	tokens := []token.Token{
		{Type: token.IDENT, Value: "TEST", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if result[0].Value != "100" {
		t.Errorf("expected '100', got '%s'", result[0].Value)
	}
}

func TestDupSimple(t *testing.T) {
	// .dup 5 0 -> 0, 0, 0, 0, 0
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".byte", Row: 1},
		{Type: token.DIRECTIVE, Value: ".dup", Row: 1},
		{Type: token.NUMBER, Value: "5", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: .byte, 0, COMMA, 0, COMMA, 0, COMMA, 0, COMMA, 0, EOL, EOF
	// Count the NUMBER tokens (should be 5)
	numCount := 0
	commaCount := 0
	for _, tok := range result {
		if tok.Type == token.NUMBER && tok.Value == "0" {
			numCount++
		}
		if tok.Type == token.COMMA {
			commaCount++
		}
	}
	if numCount != 5 {
		t.Errorf("expected 5 NUMBER tokens, got %d", numCount)
	}
	if commaCount != 4 {
		t.Errorf("expected 4 COMMA tokens, got %d", commaCount)
	}
}

func TestDupWithHexValue(t *testing.T) {
	// .dup 3 0xFF -> 255, 255, 255
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".dup", Row: 1},
		{Type: token.NUMBER, Value: "3", Row: 1},
		{Type: token.NUMBER, Value: "0xFF", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: 255, COMMA, 255, COMMA, 255, EOF
	numCount := 0
	for _, tok := range result {
		if tok.Type == token.NUMBER && tok.Value == "255" {
			numCount++
		}
	}
	if numCount != 3 {
		t.Errorf("expected 3 NUMBER(255) tokens, got %d", numCount)
	}
}

func TestDupWithExpression(t *testing.T) {
	// .define SIZE 20
	// .dup (SIZE / 4) 0xAA -> 5 copies of 170
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "SIZE", Row: 1},
		{Type: token.NUMBER, Value: "20", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".dup", Row: 2},
		{Type: token.LPAREN, Value: "(", Row: 2},
		{Type: token.IDENT, Value: "SIZE", Row: 2},
		{Type: token.SLASH, Value: "/", Row: 2},
		{Type: token.NUMBER, Value: "4", Row: 2},
		{Type: token.RPAREN, Value: ")", Row: 2},
		{Type: token.NUMBER, Value: "0xAA", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have 5 copies of 170
	numCount := 0
	for _, tok := range result {
		if tok.Type == token.NUMBER && tok.Value == "170" {
			numCount++
		}
	}
	if numCount != 5 {
		t.Errorf("expected 5 NUMBER(170) tokens, got %d", numCount)
	}
}

func TestDupWithDefinedCount(t *testing.T) {
	// .define COUNT 4
	// .dup COUNT 1 -> 1, 1, 1, 1
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "COUNT", Row: 1},
		{Type: token.NUMBER, Value: "4", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".dup", Row: 2},
		{Type: token.IDENT, Value: "COUNT", Row: 2},
		{Type: token.NUMBER, Value: "1", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have 4 copies of 1
	numCount := 0
	for _, tok := range result {
		if tok.Type == token.NUMBER && tok.Value == "1" {
			numCount++
		}
	}
	if numCount != 4 {
		t.Errorf("expected 4 NUMBER(1) tokens, got %d", numCount)
	}
}

func TestDupZeroCount(t *testing.T) {
	// .dup 0 42 -> nothing
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".dup", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.NUMBER, Value: "42", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should only have EOF
	numCount := 0
	for _, tok := range result {
		if tok.Type == token.NUMBER {
			numCount++
		}
	}
	if numCount != 0 {
		t.Errorf("expected 0 NUMBER tokens, got %d", numCount)
	}
}

func TestDupMissingCount(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".dup", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	_, err := p.Process(tokens)
	if err == nil {
		t.Error("expected error for missing count")
	}
}

func TestDupMissingValue(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".dup", Row: 1},
		{Type: token.NUMBER, Value: "5", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	_, err := p.Process(tokens)
	if err == nil {
		t.Error("expected error for missing value")
	}
}

func TestPasteSimple(t *testing.T) {
	// x##0 -> x0
	tokens := []token.Token{
		{Type: token.IDENT, Value: "x", Row: 1},
		{Type: token.PASTE, Value: "##", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: IDENT(x0), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Type != token.IDENT {
		t.Errorf("expected IDENT, got %v", result[0].Type)
	}
	if result[0].Value != "x0" {
		t.Errorf("expected 'x0', got '%s'", result[0].Value)
	}
}

func TestPasteWithDefinedSymbol(t *testing.T) {
	// .define i 5
	// x##i -> x5
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "i", Row: 1},
		{Type: token.NUMBER, Value: "5", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "x", Row: 2},
		{Type: token.PASTE, Value: "##", Row: 2},
		{Type: token.IDENT, Value: "i", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: IDENT(x5), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Value != "x5" {
		t.Errorf("expected 'x5', got '%s'", result[0].Value)
	}
}

func TestPasteWithLabel(t *testing.T) {
	// handler##0: -> handler0:
	tokens := []token.Token{
		{Type: token.IDENT, Value: "handler", Row: 1},
		{Type: token.PASTE, Value: "##", Row: 1},
		{Type: token.LABEL, Value: "0:", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: LABEL(handler0:), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Type != token.LABEL {
		t.Errorf("expected LABEL, got %v", result[0].Type)
	}
	if result[0].Value != "handler0:" {
		t.Errorf("expected 'handler0:', got '%s'", result[0].Value)
	}
}

func TestPasteIdentWithIdent(t *testing.T) {
	// foo##bar -> foobar
	tokens := []token.Token{
		{Type: token.IDENT, Value: "foo", Row: 1},
		{Type: token.PASTE, Value: "##", Row: 1},
		{Type: token.IDENT, Value: "bar", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: IDENT(foobar), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Value != "foobar" {
		t.Errorf("expected 'foobar', got '%s'", result[0].Value)
	}
}

func TestPasteMissingLeft(t *testing.T) {
	tokens := []token.Token{
		{Type: token.PASTE, Value: "##", Row: 1},
		{Type: token.IDENT, Value: "x", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	_, err := p.Process(tokens)
	if err == nil {
		t.Error("expected error for missing left operand")
	}
}

func TestPasteMissingRight(t *testing.T) {
	tokens := []token.Token{
		{Type: token.IDENT, Value: "x", Row: 1},
		{Type: token.PASTE, Value: "##", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	_, err := p.Process(tokens)
	if err == nil {
		t.Error("expected error for missing right operand")
	}
}

func TestPasteMultiple(t *testing.T) {
	// a##b##c -> abc (left to right)
	tokens := []token.Token{
		{Type: token.IDENT, Value: "a", Row: 1},
		{Type: token.PASTE, Value: "##", Row: 1},
		{Type: token.IDENT, Value: "b", Row: 1},
		{Type: token.PASTE, Value: "##", Row: 1},
		{Type: token.IDENT, Value: "c", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: IDENT(abc), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Value != "abc" {
		t.Errorf("expected 'abc', got '%s'", result[0].Value)
	}
}

func TestForSimple(t *testing.T) {
	// .for i = 0, 3
	//     .word i
	// .endf
	// -> .word 0, .word 1, .word 2
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".for", Row: 1},
		{Type: token.IDENT, Value: "i", Row: 1},
		{Type: token.ASSIGN, Value: "=", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "3", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".word", Row: 2},
		{Type: token.IDENT, Value: "i", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.EOF, Row: 4},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Count .word directives and numbers
	wordCount := 0
	numbers := []string{}
	for i, tok := range result {
		if tok.Type == token.DIRECTIVE && tok.Value == ".word" {
			wordCount++
		}
		if tok.Type == token.NUMBER {
			numbers = append(numbers, tok.Value)
			_ = i
		}
	}

	if wordCount != 3 {
		t.Errorf("expected 3 .word directives, got %d", wordCount)
	}
	expectedNums := []string{"0", "1", "2"}
	if len(numbers) != len(expectedNums) {
		t.Errorf("expected %d numbers, got %d", len(expectedNums), len(numbers))
	} else {
		for i, n := range numbers {
			if n != expectedNums[i] {
				t.Errorf("number %d: expected '%s', got '%s'", i, expectedNums[i], n)
			}
		}
	}
}

func TestForCountingDown(t *testing.T) {
	// .for j = 5, 2
	//     .byte j
	// .endf
	// -> .byte 5, .byte 4, .byte 3 (auto step -1)
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".for", Row: 1},
		{Type: token.IDENT, Value: "j", Row: 1},
		{Type: token.ASSIGN, Value: "=", Row: 1},
		{Type: token.NUMBER, Value: "5", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "2", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".byte", Row: 2},
		{Type: token.IDENT, Value: "j", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 3},
		{Type: token.EOF, Row: 3},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	numbers := []string{}
	for _, tok := range result {
		if tok.Type == token.NUMBER {
			numbers = append(numbers, tok.Value)
		}
	}

	expectedNums := []string{"5", "4", "3"}
	if len(numbers) != len(expectedNums) {
		t.Errorf("expected %d numbers, got %d: %v", len(expectedNums), len(numbers), numbers)
	} else {
		for i, n := range numbers {
			if n != expectedNums[i] {
				t.Errorf("number %d: expected '%s', got '%s'", i, expectedNums[i], n)
			}
		}
	}
}

func TestForWithStep(t *testing.T) {
	// .for k = 0, 10, 2
	//     .byte k
	// .endf
	// -> .byte 0, .byte 2, .byte 4, .byte 6, .byte 8
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".for", Row: 1},
		{Type: token.IDENT, Value: "k", Row: 1},
		{Type: token.ASSIGN, Value: "=", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "10", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "2", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".byte", Row: 2},
		{Type: token.IDENT, Value: "k", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 3},
		{Type: token.EOF, Row: 3},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	numbers := []string{}
	for _, tok := range result {
		if tok.Type == token.NUMBER {
			numbers = append(numbers, tok.Value)
		}
	}

	expectedNums := []string{"0", "2", "4", "6", "8"}
	if len(numbers) != len(expectedNums) {
		t.Errorf("expected %d numbers, got %d: %v", len(expectedNums), len(numbers), numbers)
	} else {
		for i, n := range numbers {
			if n != expectedNums[i] {
				t.Errorf("number %d: expected '%s', got '%s'", i, expectedNums[i], n)
			}
		}
	}
}

func TestForWithNegativeStep(t *testing.T) {
	// .for m = 10, 0, -2
	//     .byte m
	// .endf
	// -> .byte 10, .byte 8, .byte 6, .byte 4, .byte 2
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".for", Row: 1},
		{Type: token.IDENT, Value: "m", Row: 1},
		{Type: token.ASSIGN, Value: "=", Row: 1},
		{Type: token.NUMBER, Value: "10", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.MINUS, Value: "-", Row: 1},
		{Type: token.NUMBER, Value: "2", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".byte", Row: 2},
		{Type: token.IDENT, Value: "m", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 3},
		{Type: token.EOF, Row: 3},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	numbers := []string{}
	for _, tok := range result {
		if tok.Type == token.NUMBER {
			numbers = append(numbers, tok.Value)
		}
	}

	expectedNums := []string{"10", "8", "6", "4", "2"}
	if len(numbers) != len(expectedNums) {
		t.Errorf("expected %d numbers, got %d: %v", len(expectedNums), len(numbers), numbers)
	} else {
		for i, n := range numbers {
			if n != expectedNums[i] {
				t.Errorf("number %d: expected '%s', got '%s'", i, expectedNums[i], n)
			}
		}
	}
}

func TestForWithTokenPaste(t *testing.T) {
	// .for r = 0, 3
	//     x##r
	// .endf
	// -> x0, x1, x2
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".for", Row: 1},
		{Type: token.IDENT, Value: "r", Row: 1},
		{Type: token.ASSIGN, Value: "=", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "3", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "x", Row: 2},
		{Type: token.PASTE, Value: "##", Row: 2},
		{Type: token.IDENT, Value: "r", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 3},
		{Type: token.EOF, Row: 3},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: x0, EOL, x1, EOL, x2, EOL, EOF
	idents := []string{}
	for _, tok := range result {
		if tok.Type == token.IDENT {
			idents = append(idents, tok.Value)
		}
	}

	expectedIdents := []string{"x0", "x1", "x2"}
	if len(idents) != len(expectedIdents) {
		t.Errorf("expected %d idents, got %d: %v", len(expectedIdents), len(idents), idents)
	} else {
		for i, id := range idents {
			if id != expectedIdents[i] {
				t.Errorf("ident %d: expected '%s', got '%s'", i, expectedIdents[i], id)
			}
		}
	}
}

func TestForWithDefinedBounds(t *testing.T) {
	// .define START 1
	// .define END 4
	// .for i = START, END
	//     .word i
	// .endf
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "START", Row: 1},
		{Type: token.NUMBER, Value: "1", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".define", Row: 2},
		{Type: token.IDENT, Value: "END", Row: 2},
		{Type: token.NUMBER, Value: "4", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".for", Row: 3},
		{Type: token.IDENT, Value: "i", Row: 3},
		{Type: token.ASSIGN, Value: "=", Row: 3},
		{Type: token.IDENT, Value: "START", Row: 3},
		{Type: token.COMMA, Value: ",", Row: 3},
		{Type: token.IDENT, Value: "END", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.DIRECTIVE, Value: ".word", Row: 4},
		{Type: token.IDENT, Value: "i", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 5},
		{Type: token.EOF, Row: 5},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	numbers := []string{}
	for _, tok := range result {
		if tok.Type == token.NUMBER {
			numbers = append(numbers, tok.Value)
		}
	}

	expectedNums := []string{"1", "2", "3"}
	if len(numbers) != len(expectedNums) {
		t.Errorf("expected %d numbers, got %d: %v", len(expectedNums), len(numbers), numbers)
	} else {
		for i, n := range numbers {
			if n != expectedNums[i] {
				t.Errorf("number %d: expected '%s', got '%s'", i, expectedNums[i], n)
			}
		}
	}
}

func TestForZeroIterations(t *testing.T) {
	// .for i = 5, 5
	//     .word i
	// .endf
	// -> nothing (start == end)
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".for", Row: 1},
		{Type: token.IDENT, Value: "i", Row: 1},
		{Type: token.ASSIGN, Value: "=", Row: 1},
		{Type: token.NUMBER, Value: "5", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "5", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".word", Row: 2},
		{Type: token.IDENT, Value: "i", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 3},
		{Type: token.EOF, Row: 3},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should only have EOF
	wordCount := 0
	for _, tok := range result {
		if tok.Type == token.DIRECTIVE && tok.Value == ".word" {
			wordCount++
		}
	}
	if wordCount != 0 {
		t.Errorf("expected 0 .word directives, got %d", wordCount)
	}
}

func TestForMissingEndf(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".for", Row: 1},
		{Type: token.IDENT, Value: "i", Row: 1},
		{Type: token.ASSIGN, Value: "=", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "5", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".word", Row: 2},
		{Type: token.IDENT, Value: "i", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	_, err := p.Process(tokens)
	if err == nil {
		t.Error("expected error for missing .endf")
	}
}

func TestForZeroStep(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".for", Row: 1},
		{Type: token.IDENT, Value: "i", Row: 1},
		{Type: token.ASSIGN, Value: "=", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "5", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".word", Row: 2},
		{Type: token.IDENT, Value: "i", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 3},
		{Type: token.EOF, Row: 3},
	}

	p := New("")
	_, err := p.Process(tokens)
	if err == nil {
		t.Error("expected error for zero step")
	}
}

func TestForNested(t *testing.T) {
	// .for i = 0, 2
	//     .for j = 0, 2
	//         .word i
	//     .endf
	// .endf
	// -> .word 0, .word 0, .word 1, .word 1
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".for", Row: 1},
		{Type: token.IDENT, Value: "i", Row: 1},
		{Type: token.ASSIGN, Value: "=", Row: 1},
		{Type: token.NUMBER, Value: "0", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.NUMBER, Value: "2", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.DIRECTIVE, Value: ".for", Row: 2},
		{Type: token.IDENT, Value: "j", Row: 2},
		{Type: token.ASSIGN, Value: "=", Row: 2},
		{Type: token.NUMBER, Value: "0", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.NUMBER, Value: "2", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".word", Row: 3},
		{Type: token.IDENT, Value: "i", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.DIRECTIVE, Value: ".endf", Row: 5},
		{Type: token.EOF, Row: 5},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	numbers := []string{}
	for _, tok := range result {
		if tok.Type == token.NUMBER {
			numbers = append(numbers, tok.Value)
		}
	}

	// Outer loop: i=0, i=1
	// Inner loop for each i: j=0, j=1
	// So we get: 0, 0, 1, 1
	expectedNums := []string{"0", "0", "1", "1"}
	if len(numbers) != len(expectedNums) {
		t.Errorf("expected %d numbers, got %d: %v", len(expectedNums), len(numbers), numbers)
	} else {
		for i, n := range numbers {
			if n != expectedNums[i] {
				t.Errorf("number %d: expected '%s', got '%s'", i, expectedNums[i], n)
			}
		}
	}
}

func TestMacroSimple(t *testing.T) {
	// .macro nop
	//     addi x0, x0, 0
	// .endmacro
	// nop
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".macro", Row: 1},
		{Type: token.IDENT, Value: "nop", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "addi", Row: 2},
		{Type: token.IDENT, Value: "x0", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.IDENT, Value: "x0", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.NUMBER, Value: "0", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endmacro", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.IDENT, Value: "nop", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.EOF, Row: 5},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: addi x0, x0, 0 expanded
	foundAddi := false
	for _, tok := range result {
		if tok.Type == token.IDENT && tok.Value == "addi" {
			foundAddi = true
			break
		}
	}
	if !foundAddi {
		t.Error("expected 'addi' in output from macro expansion")
	}
}

func TestMacroWithParam(t *testing.T) {
	// .macro inc reg
	//     addi \reg, \reg, 1
	// .endmacro
	// inc x5
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".macro", Row: 1},
		{Type: token.IDENT, Value: "inc", Row: 1},
		{Type: token.IDENT, Value: "reg", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "addi", Row: 2},
		{Type: token.MACRO_PARAM, Value: "reg", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.MACRO_PARAM, Value: "reg", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.NUMBER, Value: "1", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endmacro", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.IDENT, Value: "inc", Row: 4},
		{Type: token.IDENT, Value: "x5", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.EOF, Row: 5},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: addi x5, x5, 1
	x5Count := 0
	for _, tok := range result {
		if tok.Type == token.IDENT && tok.Value == "x5" {
			x5Count++
		}
	}
	if x5Count != 2 {
		t.Errorf("expected 2 x5 tokens, got %d", x5Count)
	}
}

func TestMacroMultipleParams(t *testing.T) {
	// .macro add3 dst, a, b
	//     add \dst, \a, \b
	// .endmacro
	// add3 x1, x2, x3
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".macro", Row: 1},
		{Type: token.IDENT, Value: "add3", Row: 1},
		{Type: token.IDENT, Value: "dst", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.IDENT, Value: "a", Row: 1},
		{Type: token.COMMA, Value: ",", Row: 1},
		{Type: token.IDENT, Value: "b", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "add", Row: 2},
		{Type: token.MACRO_PARAM, Value: "dst", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.MACRO_PARAM, Value: "a", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.MACRO_PARAM, Value: "b", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endmacro", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.IDENT, Value: "add3", Row: 4},
		{Type: token.IDENT, Value: "x1", Row: 4},
		{Type: token.COMMA, Value: ",", Row: 4},
		{Type: token.IDENT, Value: "x2", Row: 4},
		{Type: token.COMMA, Value: ",", Row: 4},
		{Type: token.IDENT, Value: "x3", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.EOF, Row: 5},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: add x1, x2, x3
	idents := []string{}
	for _, tok := range result {
		if tok.Type == token.IDENT {
			idents = append(idents, tok.Value)
		}
	}

	expectedIdents := []string{"add", "x1", "x2", "x3"}
	if len(idents) != len(expectedIdents) {
		t.Errorf("expected %d idents, got %d: %v", len(expectedIdents), len(idents), idents)
	} else {
		for i, id := range idents {
			if id != expectedIdents[i] {
				t.Errorf("ident %d: expected '%s', got '%s'", i, expectedIdents[i], id)
			}
		}
	}
}

func TestMacroMultipleInvocations(t *testing.T) {
	// .macro inc reg
	//     addi \reg, \reg, 1
	// .endmacro
	// inc x1
	// inc x2
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".macro", Row: 1},
		{Type: token.IDENT, Value: "inc", Row: 1},
		{Type: token.IDENT, Value: "reg", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "addi", Row: 2},
		{Type: token.MACRO_PARAM, Value: "reg", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.MACRO_PARAM, Value: "reg", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.NUMBER, Value: "1", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endmacro", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.IDENT, Value: "inc", Row: 4},
		{Type: token.IDENT, Value: "x1", Row: 4},
		{Type: token.EOL, Row: 4},
		{Type: token.IDENT, Value: "inc", Row: 5},
		{Type: token.IDENT, Value: "x2", Row: 5},
		{Type: token.EOL, Row: 5},
		{Type: token.EOF, Row: 6},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have two addi instructions
	addiCount := 0
	for _, tok := range result {
		if tok.Type == token.IDENT && tok.Value == "addi" {
			addiCount++
		}
	}
	if addiCount != 2 {
		t.Errorf("expected 2 addi tokens, got %d", addiCount)
	}
}

func TestMacroMissingEndmacro(t *testing.T) {
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".macro", Row: 1},
		{Type: token.IDENT, Value: "test", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "nop", Row: 2},
		{Type: token.EOF, Row: 3},
	}

	p := New("")
	_, err := p.Process(tokens)
	if err == nil {
		t.Error("expected error for missing .endmacro")
	}
}

func TestMacroEndm(t *testing.T) {
	// Test that .endm also works as .endmacro
	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".macro", Row: 1},
		{Type: token.IDENT, Value: "nop", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "addi", Row: 2},
		{Type: token.IDENT, Value: "x0", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.IDENT, Value: "x0", Row: 2},
		{Type: token.COMMA, Value: ",", Row: 2},
		{Type: token.NUMBER, Value: "0", Row: 2},
		{Type: token.EOL, Row: 2},
		{Type: token.DIRECTIVE, Value: ".endm", Row: 3},
		{Type: token.EOL, Row: 3},
		{Type: token.IDENT, Value: "nop", Row: 4},
		{Type: token.EOF, Row: 4},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	foundAddi := false
	for _, tok := range result {
		if tok.Type == token.IDENT && tok.Value == "addi" {
			foundAddi = true
			break
		}
	}
	if !foundAddi {
		t.Error("expected 'addi' in output from macro expansion")
	}
}

func TestEnvVarNumeric(t *testing.T) {
	// Set environment variable
	os.Setenv("GOASM_TEST_VALUE", "42")
	defer os.Unsetenv("GOASM_TEST_VALUE")

	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "VAL", Row: 1},
		{Type: token.ENVVAR, Value: "GOASM_TEST_VALUE", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "VAL", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: NUMBER(42), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Type != token.NUMBER {
		t.Errorf("expected NUMBER, got %v", result[0].Type)
	}
	if result[0].Value != "42" {
		t.Errorf("expected '42', got '%s'", result[0].Value)
	}
}

func TestEnvVarHex(t *testing.T) {
	// Set environment variable with hex value
	os.Setenv("GOASM_TEST_HEX", "0x1000")
	defer os.Unsetenv("GOASM_TEST_HEX")

	tokens := []token.Token{
		{Type: token.ENVVAR, Value: "GOASM_TEST_HEX", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: NUMBER(4096), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Type != token.NUMBER {
		t.Errorf("expected NUMBER, got %v", result[0].Type)
	}
	if result[0].Value != "4096" {
		t.Errorf("expected '4096', got '%s'", result[0].Value)
	}
}

func TestEnvVarString(t *testing.T) {
	// Set environment variable with non-numeric value
	os.Setenv("GOASM_TEST_PATH", "/usr/local/bin")
	defer os.Unsetenv("GOASM_TEST_PATH")

	tokens := []token.Token{
		{Type: token.ENVVAR, Value: "GOASM_TEST_PATH", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: STRING("/usr/local/bin"), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Type != token.STRING {
		t.Errorf("expected STRING, got %v", result[0].Type)
	}
	if result[0].Value != "\"/usr/local/bin\"" {
		t.Errorf("expected '\"/usr/local/bin\"', got '%s'", result[0].Value)
	}
}

func TestEnvVarUndefined(t *testing.T) {
	// Make sure the variable doesn't exist
	os.Unsetenv("GOASM_UNDEFINED_VAR")

	tokens := []token.Token{
		{Type: token.IDENT, Value: "test", Row: 1},
		{Type: token.ENVVAR, Value: "GOASM_UNDEFINED_VAR", Row: 1},
		{Type: token.EOF, Row: 1},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Undefined env var should be removed, so just IDENT(test), EOF
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(result), result)
	}
	if result[0].Type != token.IDENT {
		t.Errorf("expected IDENT, got %v", result[0].Type)
	}
	if result[0].Value != "test" {
		t.Errorf("expected 'test', got '%s'", result[0].Value)
	}
}

func TestEnvVarInExpression(t *testing.T) {
	// Set environment variable
	os.Setenv("GOASM_BASE", "0x1000")
	os.Setenv("GOASM_OFFSET", "256")
	defer os.Unsetenv("GOASM_BASE")
	defer os.Unsetenv("GOASM_OFFSET")

	tokens := []token.Token{
		{Type: token.DIRECTIVE, Value: ".define", Row: 1},
		{Type: token.IDENT, Value: "ADDR", Row: 1},
		{Type: token.ENVVAR, Value: "GOASM_BASE", Row: 1},
		{Type: token.PLUS, Value: "+", Row: 1},
		{Type: token.ENVVAR, Value: "GOASM_OFFSET", Row: 1},
		{Type: token.EOL, Row: 1},
		{Type: token.IDENT, Value: "ADDR", Row: 2},
		{Type: token.EOF, Row: 2},
	}

	p := New("")
	result, err := p.Process(tokens)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have: NUMBER(4352), EOF (0x1000 + 256 = 4352)
	if len(result) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(result))
	}
	if result[0].Type != token.NUMBER {
		t.Errorf("expected NUMBER, got %v", result[0].Type)
	}
	if result[0].Value != "4352" {
		t.Errorf("expected '4352', got '%s'", result[0].Value)
	}
}
