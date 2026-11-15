package pseudo

import (
	"testing"

	"github.com/natoga32/goasm/internal/ast"
)

func TestExpandLiSmall(t *testing.T) {
	e := New()

	// li x1, 10 -> addi x1, x0, 10
	i := &ast.Instruction{
		Opcode: "li",
		Operands: []ast.Expression{
			&ast.Identifier{Name: "x1"},
			&ast.Number{Value: 10},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}
	if result[0].Opcode != "addi" {
		t.Errorf("expected 'addi', got '%s'", result[0].Opcode)
	}
}

func TestExpandLiLarge(t *testing.T) {
	e := New()

	// li x1, 0x12345 -> lui x1, upper; addi x1, x1, lower
	i := &ast.Instruction{
		Opcode: "li",
		Operands: []ast.Expression{
			&ast.Identifier{Name: "x1"},
			&ast.Number{Value: 0x12345},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(result))
	}
	if result[0].Opcode != "lui" {
		t.Errorf("expected 'lui', got '%s'", result[0].Opcode)
	}
	if result[1].Opcode != "addi" {
		t.Errorf("expected 'addi', got '%s'", result[1].Opcode)
	}
}

func TestExpandLa(t *testing.T) {
	e := New()

	// la x1, symbol -> auipc x1, %pcrel_hi(symbol); addi x1, x1, %pcrel_lo(symbol)
	i := &ast.Instruction{
		Opcode: "la",
		Operands: []ast.Expression{
			&ast.Identifier{Name: "x1"},
			&ast.Identifier{Name: "symbol"},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(result))
	}
	if result[0].Opcode != "auipc" {
		t.Errorf("expected 'auipc', got '%s'", result[0].Opcode)
	}
	if result[1].Opcode != "addi" {
		t.Errorf("expected 'addi', got '%s'", result[1].Opcode)
	}

	// Check that pcrel_hi is used
	if _, ok := result[0].Operands[1].(*ast.PCRelHi); !ok {
		t.Error("expected pcrel_hi relocation")
	}
	if _, ok := result[1].Operands[2].(*ast.PCRelLo); !ok {
		t.Error("expected pcrel_lo relocation")
	}
}

func TestExpandMv(t *testing.T) {
	e := New()

	// mv x1, x2 -> addi x1, x2, 0
	i := &ast.Instruction{
		Opcode: "mv",
		Operands: []ast.Expression{
			&ast.Identifier{Name: "x1"},
			&ast.Identifier{Name: "x2"},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}
	if result[0].Opcode != "addi" {
		t.Errorf("expected 'addi', got '%s'", result[0].Opcode)
	}
}

func TestExpandNop(t *testing.T) {
	e := New()

	// nop -> addi x0, x0, 0
	i := &ast.Instruction{
		Opcode:   "nop",
		Operands: []ast.Expression{},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}
	if result[0].Opcode != "addi" {
		t.Errorf("expected 'addi', got '%s'", result[0].Opcode)
	}
}

func TestExpandRet(t *testing.T) {
	e := New()

	// ret -> jalr x0, x1, 0
	i := &ast.Instruction{
		Opcode:   "ret",
		Operands: []ast.Expression{},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}
	if result[0].Opcode != "jalr" {
		t.Errorf("expected 'jalr', got '%s'", result[0].Opcode)
	}
}

func TestExpandCall(t *testing.T) {
	e := New()

	// call func -> auipc ra, %pcrel_hi(func); jalr ra, ra, %pcrel_lo(func)
	i := &ast.Instruction{
		Opcode: "call",
		Operands: []ast.Expression{
			&ast.Identifier{Name: "func"},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(result))
	}
	if result[0].Opcode != "auipc" {
		t.Errorf("expected 'auipc', got '%s'", result[0].Opcode)
	}
	if result[1].Opcode != "jalr" {
		t.Errorf("expected 'jalr', got '%s'", result[1].Opcode)
	}
}

func TestExpandJ(t *testing.T) {
	e := New()

	// j label -> jal x0, label
	i := &ast.Instruction{
		Opcode: "j",
		Operands: []ast.Expression{
			&ast.Identifier{Name: "label"},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}
	if result[0].Opcode != "jal" {
		t.Errorf("expected 'jal', got '%s'", result[0].Opcode)
	}
}

func TestExpandBeqz(t *testing.T) {
	e := New()

	// beqz x1, label -> beq x1, x0, label
	i := &ast.Instruction{
		Opcode: "beqz",
		Operands: []ast.Expression{
			&ast.Identifier{Name: "x1"},
			&ast.Identifier{Name: "label"},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}
	if result[0].Opcode != "beq" {
		t.Errorf("expected 'beq', got '%s'", result[0].Opcode)
	}
	if len(result[0].Operands) != 3 {
		t.Errorf("expected 3 operands, got %d", len(result[0].Operands))
	}
}

func TestExpandBgt(t *testing.T) {
	e := New()

	// bgt x1, x2, label -> blt x2, x1, label (swap operands)
	i := &ast.Instruction{
		Opcode: "bgt",
		Operands: []ast.Expression{
			&ast.Identifier{Name: "x1"},
			&ast.Identifier{Name: "x2"},
			&ast.Identifier{Name: "label"},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}
	if result[0].Opcode != "blt" {
		t.Errorf("expected 'blt', got '%s'", result[0].Opcode)
	}
}

func TestExpandSyscall(t *testing.T) {
	e := New()

	// syscall 1 -> addi a7, x0, 1; ecall
	i := &ast.Instruction{
		Opcode: "syscall",
		Operands: []ast.Expression{
			&ast.Number{Value: 1},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(result))
	}
	if result[0].Opcode != "addi" {
		t.Errorf("expected 'addi', got '%s'", result[0].Opcode)
	}
	if result[1].Opcode != "ecall" {
		t.Errorf("expected 'ecall', got '%s'", result[1].Opcode)
	}
}

func TestExpandProgram(t *testing.T) {
	e := New()

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.Label{Name: "main"},
			&ast.Instruction{
				Opcode: "li",
				Operands: []ast.Expression{
					&ast.Identifier{Name: "x1"},
					&ast.Number{Value: 10},
				},
			},
			&ast.Instruction{
				Opcode: "ret",
			},
		},
	}

	result := e.ExpandProgram(program)

	// Should have: label, addi, jalr
	if len(result.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(result.Statements))
	}

	if _, ok := result.Statements[0].(*ast.Label); !ok {
		t.Error("expected first statement to be a label")
	}
	if instr, ok := result.Statements[1].(*ast.Instruction); !ok || instr.Opcode != "addi" {
		t.Error("expected second statement to be 'addi'")
	}
	if instr, ok := result.Statements[2].(*ast.Instruction); !ok || instr.Opcode != "jalr" {
		t.Error("expected third statement to be 'jalr'")
	}
}

func TestNonPseudoInstruction(t *testing.T) {
	e := New()

	// Regular instruction should pass through unchanged
	i := &ast.Instruction{
		Opcode: "add",
		Operands: []ast.Expression{
			&ast.Identifier{Name: "x1"},
			&ast.Identifier{Name: "x2"},
			&ast.Identifier{Name: "x3"},
		},
	}

	result := e.ExpandInstruction(i)
	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}
	if result[0] != i {
		t.Error("expected same instruction object")
	}
}
