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

// syscall with no operand is a bare ecall; syscall N and syscall rs set a7 first
func TestSyscallForms(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	ecall := []byte{0x73, 0x00, 0x00, 0x00}
	expectAssembles(t, goasm, "syscall", ecall)
	expectAssembles(t, goasm, "syscall 5", append([]byte{0x93, 0x08, 0x50, 0x00}, ecall...))  // addi a7, zero, 5
	expectAssembles(t, goasm, "syscall a0", append([]byte{0xb3, 0x08, 0xa0, 0x00}, ecall...)) // add a7, zero, a0
}

// -f bin is removed: it is refused before any file is written
func TestBinFormatRemoved(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	dir := t.TempDir()
	input := filepath.Join(dir, "in.s")
	output := filepath.Join(dir, "out.bin")
	if err := os.WriteFile(input, []byte(".text\nnop\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(goasm, input, "-f", "bin", "-o", output).CombinedOutput()
	if err == nil || !strings.Contains(string(out), "unknown output format: bin") {
		t.Errorf("expected 'unknown output format: bin', got (err=%v):\n%s", err, out)
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Errorf("%s should not have been created", output)
	}
	help, _ := exec.Command(goasm, "--help").CombinedOutput()
	if strings.Contains(string(help), "rscv, bin") {
		t.Errorf("--help still lists bin:\n%s", help)
	}
}

// a raw file (a rom image) is disassembled as "raw" at --base
func TestDisasmRawFormatName(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	data, out, err := assembleSource(t, goasm, ".rom\nnop\n", "rom")
	if err != nil {
		t.Fatalf("assembling failed: %v\n%s", err, out)
	}
	rom := filepath.Join(t.TempDir(), "x.rom")
	if err := os.WriteFile(rom, data, 0o644); err != nil {
		t.Fatal(err)
	}
	dis, err := exec.Command(goasm, "--disasm", rom, "--base", "0").CombinedOutput()
	if err != nil {
		t.Fatalf("disasm failed: %v\n%s", err, dis)
	}
	for _, want := range []string{"(raw format)", "0x00000000:"} {
		if !strings.Contains(string(dis), want) {
			t.Errorf("expected %q in output, got:\n%s", want, dis)
		}
	}
}

// jalr: standard short forms "jalr rs" (rd = ra) and "jalr rd, rs" (offset 0)
func TestJalrForms(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	jalrRaA0 := []byte{0xe7, 0x00, 0x05, 0x00} // jalr ra, a0, 0
	expectAssembles(t, goasm, "jalr a0", jalrRaA0)
	expectAssembles(t, goasm, "jalr ra, a0", jalrRaA0)
	expectAssembles(t, goasm, "jalr ra, a0, 0", jalrRaA0)
	expectAssembles(t, goasm, "jalr ra, 0(a0)", jalrRaA0)
	expectAssembles(t, goasm, "jalr t0, t1", []byte{0xe7, 0x02, 0x03, 0x00})     // jalr t0, t1, 0
	expectAssembles(t, goasm, "jalr ra, 4(a0)", []byte{0xe7, 0x00, 0x45, 0x00})  // jalr ra, a0, 4
	expectAssembles(t, goasm, "jalr ra, a0, -4", []byte{0xe7, 0x00, 0xc5, 0xff}) // jalr ra, a0, -4
	expectError(t, goasm, "jalr", "jalr expects 1 to 3 operands (rs | rd, rs | rd, rs, imm | rd, imm(rs)), got 0", "2")
	expectError(t, goasm, "jalr a0, a1, 4, 5", "jalr expects 1 to 3 operands (rs | rd, rs | rd, rs, imm | rd, imm(rs)), got 4", "2")
	expectError(t, goasm, "jalr 4", "jalr: expected a register", "2")
}

// rdcycle and friends read the counter CSRs; unimp is the standard trapping word;
// tail jumps through t1 without saving ra
func TestCounterUnimpTail(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	counters := []struct {
		src  string
		want []byte
	}{
		{"rdcycle a0", []byte{0x73, 0x25, 0x00, 0xc0}},    // csrrs a0, 0xC00, zero
		{"rdtime a0", []byte{0x73, 0x25, 0x10, 0xc0}},     // csrrs a0, 0xC01, zero
		{"rdinstret a0", []byte{0x73, 0x25, 0x20, 0xc0}},  // csrrs a0, 0xC02, zero
		{"rdcycleh t0", []byte{0xf3, 0x22, 0x00, 0xc8}},   // csrrs t0, 0xC80, zero
		{"rdtimeh t0", []byte{0xf3, 0x22, 0x10, 0xc8}},    // csrrs t0, 0xC81, zero
		{"rdinstreth t0", []byte{0xf3, 0x22, 0x20, 0xc8}}, // csrrs t0, 0xC82, zero
		{"unimp", []byte{0x73, 0x10, 0x00, 0xc0}},         // csrrw zero, 0xC00, zero
	}
	for _, tt := range counters {
		t.Run(tt.src, func(t *testing.T) {
			expectAssembles(t, goasm, tt.src, tt.want)
		})
	}

	// tail at 0: auipc t1, 0 ; jalr zero, t1, 12 (target at 12)
	expectAssembles(t, goasm, "tail target\nnop\ntarget: nop", []byte{
		0x17, 0x03, 0x00, 0x00, // auipc t1, 0
		0x67, 0x00, 0xc3, 0x00, // jalr zero, t1, 12
		0x13, 0x00, 0x00, 0x00,
		0x13, 0x00, 0x00, 0x00,
	})

	expectError(t, goasm, "rdcycle", "rdcycle expects 1 operand (rd), got 0", "2")
	expectError(t, goasm, "rdtime a0, a1", "rdtime expects 1 operand (rd), got 2", "2")
	expectError(t, goasm, "unimp a0", "unimp expects no operands, got 1", "2")
	expectError(t, goasm, "tail", "tail expects 1 operand (symbol), got 0", "2")
}

// --disasm --pseudo prints the one-word forms back
func TestDisasmCounterPseudo(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	src := "rdcycle a0\nrdtimeh t0\nrdinstret a1\nunimp\ncsrrs a0, 0xC00, a1\ncsrrs a0, 0x300, zero\n"
	data, out, err := assembleSource(t, goasm, ".rom\n"+src, "rom")
	if err != nil {
		t.Fatalf("assembling failed: %v\n%s", err, out)
	}
	rom := filepath.Join(t.TempDir(), "x.rom")
	if err := os.WriteFile(rom, data, 0o644); err != nil {
		t.Fatal(err)
	}
	dis, err := exec.Command(goasm, "--disasm", "--pseudo", "--hex=false", rom, "--base", "0").CombinedOutput()
	if err != nil {
		t.Fatalf("disasm failed: %v\n%s", err, dis)
	}
	for _, want := range []string{
		"rdcycle a0", "rdtimeh t0", "rdinstret a1", "unimp",
		"csrrs a0, 0xC00, a1", // sets bits: not rdcycle
		"csrr a0, 0x300",      // not a counter
	} {
		if !strings.Contains(string(dis), want) {
			t.Errorf("expected %q in output, got:\n%s", want, dis)
		}
	}
}

// a character the lexer does not know stops with its location, instead of
// being dropped
func TestLexerUnknownCharacter(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct{ src, want, loc string }{
		{".byte 1 @ 2", "unexpected character '@'", "2:9"},
		{".byte 5 ?", "unexpected character '?'", "2:9"},
		{"nop `", "unexpected character '`'", "2:5"},
		{".byte {1}", "unexpected character '{'", "2:7"},
		{".byte [1]", "unexpected character '['", "2:7"},
		{"li a0, 1 \\ 2", "unexpected character '\\'", "2:10"},
		{".byte 1 é", "unexpected character 'é'", "2:9"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectError(t, goasm, tt.src, tt.want, tt.loc)
		})
	}
	// in a comment or a string, any character is fine
	expectAssembles(t, goasm, ".byte 1 ; @ ? { é\n.byte '@', '?'\n.ascii \"`{}é\"", []byte{0x01, '@', '?', '`', '{', '}', 0xc3, 0xa9})
}

// .dup and .for errors give the file and line, like .if and .define
func TestDupForErrorLocation(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct{ src, want, line string }{
		{".byte .dup", ".dup: expected a count", "2"},
		{".byte .dup 2", ".dup: expected a value after the count", "2"},
		{".byte .dup (4/) 0", ".dup count:", "2"},
		{".byte .dup (2*2) (1/)", ".dup value:", "2"},
		{".byte .dup (-1) 0", ".dup count cannot be negative", "2"},
		{"nop\n.for 1 = 0, 2\n.endf", ".for: expected a variable name", "3"},
		{".for i 0, 2\n.endf", ".for i: expected '=' after the variable name", "2"},
		{".for i = 0 2\n.endf", ".for i start: unexpected '2' after the expression", "2"},
		{".for i = 0\n.endf", ".for i: expected ',' after the start value", "2"},
		{".for i = (1, 2\n.endf", ".for i start:", "2"},
		{".for i = 0, 2*\n.endf", ".for i end:", "2"},
		{".for i = 0, 2, 1+\n.endf", ".for i step:", "2"},
		{".for i = 0, 2, 0\n.endf", ".for step cannot be zero", "2"},
		{"nop\n.for i = 0, 2\nnop", "unterminated .for (missing .endf)", "3"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectError(t, goasm, tt.src, tt.want, tt.line)
		})
	}
}

// every preprocessor and parser error gives the file and the line
func TestPreprocessParserErrorLocation(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct{ src, want, line string }{
		{".ifdef\n.endif", ".ifdef: expected a name", "2"},
		{"nop\n.ifndef 5\n.endif", ".ifndef: expected a name", "3"},
		{"nop\n.ifdef X\nnop", "unterminated .ifdef (missing .endif)", "3"},
		{".if 1\nnop", "unterminated .if (missing .endif)", "2"},
		{".macro\n.endm", ".macro: expected a name", "2"},
		{"nop\n.macro M\nnop", "unterminated .macro M (missing .endm)", "3"},
		{".include 5", ".include: expected a file name in quotes", "2"},
		{"nop\n.include \"nope.s\"", "nope.s': no such file or directory", "3"},
		{".define X, 1\n.byte X ##", "## needs a token on its right", "3"},
		{"nop\n) nop", "unexpected ')' at the start of a line", "3"},
		{", nop", "unexpected ',' at the start of a line", "2"},
		{"1 nop", "unexpected number '1' at the start of a line", "2"},
		{".byte )", "unexpected ')' in an expression", "2"},
		{"nop\n.byte 1+", "expected a value at the end of the line", "3"},
		{"li a0, %hi 5", "expected '(' after %hi", "2"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectError(t, goasm, tt.src, tt.want, tt.line)
		})
	}

	// in an included file, the location is that file
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "inc.s"), []byte("nop\n.macro M\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.s"), []byte(".rom\n.include \"inc.s\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(goasm, filepath.Join(dir, "main.s"), "-f", "rom", "-o", filepath.Join(dir, "out.rom")).CombinedOutput()
	if err == nil || !strings.Contains(string(out), "inc.s:2: error: unterminated .macro M (missing .endm)") {
		t.Errorf("expected the error in inc.s:2, got (err=%v):\n%s", err, out)
	}

	// a recursive include names the file that includes itself and its line
	if err := os.WriteFile(filepath.Join(dir, "self.s"), []byte("nop\n.include \"self.s\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.s"), []byte(".rom\n.include \"self.s\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err = exec.Command(goasm, filepath.Join(dir, "main.s"), "-f", "rom", "-o", filepath.Join(dir, "out.rom")).CombinedOutput()
	if err == nil || !strings.Contains(string(out), "self.s:2: error: .include \"self.s\": the file includes itself") {
		t.Errorf("expected the recursive include at self.s:2, got (err=%v):\n%s", err, out)
	}
}

// operands are separated by exactly one comma: a missing, extra or trailing
// comma stops with its location
func TestOperandCommas(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct{ src, want, loc string }{
		{".word 1,", "expected a value after ','", "2:8"},
		{".byte 1,,2", "expected a value after ','", "2:8"},
		{".byte ,1", "expected a value before ','", "2:7"},
		{"add a0, a1, a2,", "expected a value after ','", "2:15"},
		{"nop,", "expected a value before ','", "2:4"},
		{".byte 1 2", "expected ',' before number '2'", "2:9"},
		{"add a0 a1, a2", "expected ',' before 'a1'", "2:8"},
		{".ascii \"a\",", "expected a value after ','", "2:11"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			expectError(t, goasm, tt.src, tt.want, tt.loc)
		})
	}
	expectAssembles(t, goasm, ".byte 1, 2 ,3\n.byte 4", []byte{1, 2, 3, 4})
}

// li with a name or an expression takes one instruction when its value
// allows it, like li with a number
func TestLiSymbolShort(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	nop := []byte{0x13, 0x00, 0x00, 0x00}
	addi8 := []byte{0x13, 0x05, 0x80, 0x00} // addi a0, zero, 8
	cat := func(parts ...[]byte) []byte { return bytes.Join(parts, nil) }
	tests := []struct {
		name, src string
		want      []byte
	}{
		{"constant", ".equ F, 8\nli a0, F", addi8},
		{"constant defined below", "li a0, F\n.equ F, 8", addi8},
		{"expression", "li a0, 4+4", addi8},
		{"negative", ".equ F, -1\nli a0, F", []byte{0x13, 0x05, 0xf0, 0xff}},
		{"low 12 bits zero", ".equ F, 0x12345000\nli a0, F", []byte{0x37, 0x55, 0x34, 0x12}},
		{"needs two", ".equ F, 0x12345678\nli a0, F", []byte{0x37, 0x55, 0x34, 0x12, 0x13, 0x05, 0x85, 0x67}},
		// x is at 12 with a two-instruction li, at 8 with one: it fits both ways
		{"label below", "nop\nli a0, x\nx: nop", cat(nop, addi8, nop)},
		// $ is the li address + 4, as documented: the two instructions stay
		{"dollar", "li a0, $", []byte{0x37, 0x05, 0x00, 0x00, 0x13, 0x05, 0x45, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectAssembles(t, goasm, tt.src, tt.want)
		})
	}
	// an undefined name is still reported at the li
	expectError(t, goasm, "nop\nli a0, nope", "undefined symbol 'nope'", "3:8")
	// a value that does not fit 32 bits is an error, not cut
	expectError(t, goasm, "nop\nli a0, F\n.equ F, 0x100000000", "li value 4294967296 (0x100000000) does not fit 32 bits", "3:1")
	expectError(t, goasm, "li a0, 0x100000000", "does not fit 32 bits", "2:1")
}

// .space, .align, .org and .equ can use names defined below them: goasm
// lays out the addresses again until they no longer change
func TestLayoutForwardNames(t *testing.T) {
	goasm := buildGoasm(t)
	defer os.Remove(goasm)

	tests := []struct {
		name, src string
		want      []byte
	}{
		{"equ of an equ below", ".equ A, B + 1\n.equ B, 2\n.byte A", []byte{3}},
		{"space of an equ below", ".space N\n.byte 9\n.equ N, 2", []byte{0, 0, 9}},
		{"align of an equ below", ".byte 1\n.align P\n.byte 2\n.equ P, 2", []byte{1, 0, 0, 0, 2}},
		{"org of an equ below", ".org START\n.byte 7\n.equ START, 3", []byte{0, 0, 0, 7}},
		// the size of the table is known only at its end label
		{"size of a table below", ".equ SIZE, end - table\n.byte SIZE\ntable: .byte 1, 2, 3\nend:", []byte{3, 1, 2, 3}},
		{"space of a size below", ".space end - tbl\ntbl: .byte 1, 2\nend:", []byte{0, 0, 1, 2}},
		// later = 0 + later: 0 is the only address that works
		{"space up to itself", ".space later - 0\nlater: nop", []byte{0x13, 0, 0, 0}},
		{"chain of equ", ".equ A, B\n.equ B, C\n.equ C, D\n.equ D, 5\n.byte A", []byte{5}},
		// a li with a name below and a forward .space
		{"li and space", "li a0, F\n.space G\n.equ F, 8\n.equ G, 1", []byte{0x13, 0x05, 0x80, 0x00, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectAssembles(t, goasm, tt.src, tt.want)
		})
	}

	errs := []struct{ name, src, want, loc string }{
		{"undefined", ".space N", "undefined symbol 'N'", "2:8"},
		{"self", ".equ A, A + 1", "the value of 'A' does not settle", "2"},
		// here = 1 + (4 - here): no address works
		{"label", "start: .byte 1\n.space 4 - (here - start)\nhere: .byte 2", "the value of 'here' does not settle", "4"},
		{"loop", ".equ A, B + 1\n.equ B, A", "does not settle", "2"},
		// the error is the one with the final values
		{"negative space", ".space N\n.equ N, 0 - 1", ".space size -1 is negative", "2:1"},
		{"duplicate", ".equ A, B\n.equ A, 1\n.equ B, 2", "symbol 'A' already defined at line 2", "3:1"},
	}
	for _, tt := range errs {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, goasm, tt.src, tt.want, tt.loc)
		})
	}
}
