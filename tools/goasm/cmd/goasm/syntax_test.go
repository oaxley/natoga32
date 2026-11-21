package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// expectAssembles assembles a .rom source and compares the bytes
func expectAssembles(t *testing.T, goasm, src string, want []byte) {
	t.Helper()
	data, out, err := assembleSource(t, goasm, ".rom\n"+src+"\n", "rom")
	if err != nil {
		t.Fatalf("assembling failed: %v\n%s", err, out)
	}
	if !bytes.Equal(data, want) {
		t.Errorf("% x, want % x", data, want)
	}
}

// expectError assembles a .rom source and checks the error text and its line
func expectError(t *testing.T, goasm, src, want, line string) {
	t.Helper()
	_, out, err := assembleSource(t, goasm, ".rom\n"+src+"\n", "rom")
	if err == nil {
		t.Fatalf("expected an error, assembled fine")
	}
	if !strings.Contains(out, want) {
		t.Errorf("expected %q in output, got:\n%s", want, out)
	}
	if !strings.Contains(out, "in.s:"+line+":") {
		t.Errorf("expected the location in.s:%s in output, got:\n%s", line, out)
	}
}

// .if and .define evaluate with C precedence, parentheses, !, ~, && and ||
func TestPreprocessorExpressions(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	values := []struct {
		expr string
		want byte
	}{
		{"2+3*4", 14},
		{"(2+3)*4", 20},
		{"((1+1)*(2+1))", 6},
		{"10-4-3", 3},
		{"100/10/5", 2},
		{"7 % 3", 1},
		{"-3 * -2", 6},
		{"~0 & 0xFF", 0xFF},
		{"2 | 1 << 2", 6},
		{"6 & 3 == 3", 0}, // C: == before &
		{"1 + 2 == 3", 1},
		{"5 > 3 && 2 > 1", 1},
		{"1 || 0 && 0", 1}, // && before ||
		{"!0", 1},
		{"!5", 0},
		{"!!7", 1},
		{"!UNDEFINED", 1},
		{"0 && 1/0", 0}, // short-circuit: no division by zero
		{"1 || 1/0", 1},
		{"3 >= 3 && 3 <= 3 && 3 != 4", 1},
	}
	for _, tt := range values {
		t.Run(".define "+tt.expr, func(t *testing.T) {
			expectAssembles(t, goasm, ".define X "+tt.expr+"\n.byte X", []byte{tt.want})
		})
	}

	conds := []struct {
		cond string
		want byte
	}{
		{"!X", 0},
		{"!Y", 1},
		{"!(X == 4)", 0},
		{"X == 4 && Y == 0", 1},
		{"X == 1 || Y == 1", 0},
		{"(X > 3) && !(Y)", 1},
	}
	for _, tt := range conds {
		t.Run(".if "+tt.cond, func(t *testing.T) {
			src := ".define X 4\n.define Y 0\n.if " + tt.cond + "\n.byte 1\n.else\n.byte 0\n.endif"
			expectAssembles(t, goasm, src, []byte{tt.want})
		})
	}
}

// A malformed .if or .define expression is an error at its line, never a silent value
func TestPreprocessorExpressionErrors(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want string
	}{
		{".if (1\n.endif", "missing ')'"},
		{".if 1)\n.endif", "unexpected ')'"},
		{".if 1 +\n.endif", "expected a value"},
		{".if 1 2\n.endif", "unexpected '2'"},
		{".if && 1\n.endif", "expected a value, got '&&'"},
		{".if\n.endif", ".if expects an expression"},
		{".if 1/0\n.endif", "division by zero"},
		{".if 1%0\n.endif", "modulo by zero"},
		{".if 1 << 64\n.endif", "shift amount 64 out of range (0..63)"},
		{".define X 2 *", "expected a value"},
		{".define X (3", "missing ')'"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectError(t, goasm, tt.src, tt.want, "2")
		})
	}
}

// A missing ')' in an operand or a data value is an error, not ignored
func TestOperandMissingParen(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	for _, src := range []string{".byte (1+2", "addi a0, zero, (3", "addi a0, zero, %lo(5", "lw a0, ((4)(sp)"} {
		t.Run(src, func(t *testing.T) {
			expectError(t, goasm, src, "missing ')'", "2")
		})
	}
	expectAssembles(t, goasm, ".byte ((1+2)*2)", []byte{6})
	expectAssembles(t, goasm, "lw a0, (4)(sp)", []byte{0x03, 0x25, 0x41, 0x00})
}

// ! ~ && || were dropped by the lexer: in an instruction, ~ works and the others are errors
func TestLogicalOperatorsInInstructions(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	expectAssembles(t, goasm, "addi a0, zero, ~0", []byte{0x13, 0x05, 0xf0, 0xff}) // addi a0, zero, -1
	expectAssembles(t, goasm, ".byte ~0x0F", []byte{0xf0})
	for _, src := range []string{"addi a0, zero, !1", ".byte 1 && 1", ".byte 1 || 0"} {
		t.Run(src, func(t *testing.T) {
			_, out, err := assembleSource(t, goasm, ".rom\n"+src+"\n", "rom")
			if err == nil {
				t.Fatalf("expected an error, assembled fine")
			}
			if !strings.Contains(out, "only in .if and .define") {
				t.Errorf("expected 'only in .if and .define' in output, got:\n%s", out)
			}
		})
	}
}

// .string, .ascii and .asciz translate the escapes
func TestStringEscapes(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want []byte
	}{
		{`.ascii "a\nb"`, []byte{0x61, 0x0a, 0x62}},
		{`.string "\t\"q\"\\"`, []byte{0x09, 0x22, 0x71, 0x22, 0x5c, 0x00}},
		{`.asciz "\r\0\'"`, []byte{0x0d, 0x00, 0x27, 0x00}},
		{`.ascii "\x41\x7f\xFF"`, []byte{0x41, 0x7f, 0xff}},
		{`.byte '\x41', '\\', '\''`, []byte{0x41, 0x5c, 0x27}},
		// the size in pass 1 is the size after the escapes
		{".ascii \"a\\nb\"\nend:\n.byte end", []byte{0x61, 0x0a, 0x62, 0x03}},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectAssembles(t, goasm, tt.src, tt.want)
		})
	}

	errors := []struct {
		src  string
		want string
	}{
		{`.ascii "\q"`, `unknown escape '\q'`},
		{`.ascii "\x4"`, `escape '\x' needs 2 hex digits`},
		{`.ascii "\xG1"`, `escape '\x' needs 2 hex digits`},
		{`.byte '\q'`, `unknown escape '\q'`},
	}
	for _, tt := range errors {
		t.Run(tt.src, func(t *testing.T) {
			expectError(t, goasm, tt.src, tt.want, "2")
		})
	}
}

// _ separates digits in a number
func TestDigitSeparators(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		src  string
		want []byte
	}{
		{".word 0x0020_6800", []byte{0x00, 0x68, 0x20, 0x00}},
		{".word 1_000_000", []byte{0x40, 0x42, 0x0f, 0x00}},
		{".byte 0b1010_0101", []byte{0xa5}},
		{".byte 0o7_7", []byte{0x3f}},
		{"addi a0, zero, 1_0", []byte{0x13, 0x05, 0xa0, 0x00}},
		{".define X 0x1_0\n.byte X", []byte{0x10}},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectAssembles(t, goasm, tt.src, tt.want)
		})
	}

	for _, src := range []string{".word 1__0", ".word 1_", ".word 0x_1", ".word 0b_1"} {
		t.Run(src, func(t *testing.T) {
			expectError(t, goasm, src, "invalid number", "2")
		})
	}
}

// .define accepts "NAME VALUE" and "NAME, VALUE"
func TestDefineCommaForm(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	expectAssembles(t, goasm, ".define A, 5\n.byte A", []byte{5})
	expectAssembles(t, goasm, ".define A, 2+3*2\n.byte A", []byte{8})
	expectAssembles(t, goasm, ".define A 5\n.byte A", []byte{5})
	expectAssembles(t, goasm, ".define FLAG\n.ifdef FLAG\n.byte 1\n.endif", []byte{1})
	expectError(t, goasm, ".define A,", ".define A: expected a value after ','", "2")
	expectError(t, goasm, ".define , 5", "expected a name after .define", "2")
}

// A name defined twice reports both lines: the second definition and the first one
func TestDuplicateLabelLine(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	expectError(t, goasm, "main:\nnop\nmain:\nnop", "symbol 'main' already defined at line 2", "4")
	expectError(t, goasm, "nop\nfoo: nop\n.equ foo, 3", "symbol 'foo' already defined at line 3", "4")
	expectError(t, goasm, ".equ K, 1\nK:", "symbol 'K' already defined at line 2", "3")
	// the first definition in an included file gives its file name
	t.Run("include", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "inc.s"), []byte("main:\nnop\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		input := filepath.Join(dir, "in.s")
		if err := os.WriteFile(input, []byte(".rom\n.include \"inc.s\"\nmain:\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command(goasm, input, "-f", "rom", "-o", filepath.Join(dir, "out.rom")).CombinedOutput()
		want := "in.s:3:1: error: symbol 'main' already defined at " + filepath.Join(dir, "inc.s") + ":1"
		if err == nil || !strings.Contains(string(out), want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	})
	// a .global before the label is not a definition
	expectAssembles(t, goasm, ".global g\ng: nop", []byte{0x13, 0x00, 0x00, 0x00})
}
