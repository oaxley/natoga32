// Unit tests for the NATOGA32 assembler optimizer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package optimizer

import (
	"fmt"
	"testing"

	"github.com/natoga32/goasm/internal/ast"
)

func TestZeroConstOptimization_AddWithZeroRight(t *testing.T) {
	opt := &ZeroConstOptimization{}

	input := []*ast.Instruction{
		{
			Opcode: "add",
			Operands: []ast.Expression{
				&ast.Identifier{Name: "t0"},
				&ast.Identifier{Name: "t1"},
				&ast.Identifier{Name: "zero"},
			},
		},
	}

	result, consumed := opt.Optimize(input, 0)
	applied := consumed > 0

	if !applied {
		t.Fatal("expected optimization to be applied")
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}

	if result[0].Opcode != "addi" {
		t.Errorf("expected opcode 'addi', got '%s'", result[0].Opcode)
	}

	if len(result[0].Operands) != 3 {
		t.Fatalf("expected 3 operands, got %d", len(result[0].Operands))
	}

	rd, _ := getRegisterName(result[0].Operands[0])
	rs, _ := getRegisterName(result[0].Operands[1])
	imm, _ := getNumberValue(result[0].Operands[2])

	if rd != "t0" || rs != "t1" || imm != 0 {
		t.Errorf("unexpected operands: rd=%s, rs=%s, imm=%d", rd, rs, imm)
	}
}

func TestZeroConstOptimization_AddWithZeroLeft(t *testing.T) {
	opt := &ZeroConstOptimization{}

	input := []*ast.Instruction{
		{
			Opcode: "add",
			Operands: []ast.Expression{
				&ast.Identifier{Name: "t0"},
				&ast.Identifier{Name: "x0"},
				&ast.Identifier{Name: "t1"},
			},
		},
	}

	result, consumed := opt.Optimize(input, 0)
	applied := consumed > 0

	if !applied {
		t.Fatal("expected optimization to be applied")
	}

	if result[0].Opcode != "addi" {
		t.Errorf("expected opcode 'addi', got '%s'", result[0].Opcode)
	}

	rd, _ := getRegisterName(result[0].Operands[0])
	rs, _ := getRegisterName(result[0].Operands[1])
	imm, _ := getNumberValue(result[0].Operands[2])

	if rd != "t0" || rs != "t1" || imm != 0 {
		t.Errorf("unexpected operands: rd=%s, rs=%s, imm=%d", rd, rs, imm)
	}
}

func TestZeroConstOptimization_OrWithZero(t *testing.T) {
	opt := &ZeroConstOptimization{}

	input := []*ast.Instruction{
		{
			Opcode: "or",
			Operands: []ast.Expression{
				&ast.Identifier{Name: "t2"},
				&ast.Identifier{Name: "t3"},
				&ast.Identifier{Name: "zero"},
			},
		},
	}

	result, consumed := opt.Optimize(input, 0)
	applied := consumed > 0

	if !applied {
		t.Fatal("expected optimization to be applied")
	}

	if result[0].Opcode != "addi" {
		t.Errorf("expected opcode 'addi', got '%s'", result[0].Opcode)
	}

	rd, _ := getRegisterName(result[0].Operands[0])
	rs, _ := getRegisterName(result[0].Operands[1])

	if rd != "t2" || rs != "t3" {
		t.Errorf("unexpected operands: rd=%s, rs=%s", rd, rs)
	}
}

func TestZeroConstOptimization_NoOptimization(t *testing.T) {
	opt := &ZeroConstOptimization{}

	input := []*ast.Instruction{
		{
			Opcode: "add",
			Operands: []ast.Expression{
				&ast.Identifier{Name: "t0"},
				&ast.Identifier{Name: "t1"},
				&ast.Identifier{Name: "t2"},
			},
		},
	}

	_, consumed := opt.Optimize(input, 0)
	applied := consumed > 0

	if applied {
		t.Error("optimization should not be applied for add without zero")
	}
}

func TestRedundantMoveElimination_ConsecutiveMoves(t *testing.T) {
	opt := &RedundantMoveOptimization{}

	// mv t0, t1; mv t1, t0: the second move has no effect
	input := []*ast.Instruction{
		ins("addi", reg("t0"), reg("t1"), num(0)),
		ins("addi", reg("t1"), reg("t0"), num(0)),
	}

	result, consumed := opt.Optimize(input, 0)
	if consumed != 2 {
		t.Fatalf("expected the pair to be replaced (consumed 2), got %d", consumed)
	}
	if len(result) != 1 || instrText(result[0]) != "addi t0, t1, 0" {
		t.Errorf("got %v, want [addi t0, t1, 0]", result)
	}
}

// mv t0, t1; mv t2, t0 is kept: t0 may be read later
func TestRedundantMoveElimination_NoOptimization_Chain(t *testing.T) {
	opt := &RedundantMoveOptimization{}

	input := []*ast.Instruction{
		ins("addi", reg("t0"), reg("t1"), num(0)),
		ins("addi", reg("t2"), reg("t0"), num(0)),
	}
	if _, consumed := opt.Optimize(input, 0); consumed != 0 {
		t.Errorf("a move chain must not be merged, consumed %d", consumed)
	}
}

// mv x0, t1; mv t1, x0 writes 0 into t1: it is not a no-op
func TestRedundantMoveElimination_NoOptimization_ZeroRegister(t *testing.T) {
	opt := &RedundantMoveOptimization{}

	input := []*ast.Instruction{
		ins("addi", reg("x0"), reg("t1"), num(0)),
		ins("addi", reg("t1"), reg("x0"), num(0)),
	}
	if _, consumed := opt.Optimize(input, 0); consumed != 0 {
		t.Errorf("a move through x0 must be kept, consumed %d", consumed)
	}
}

func TestRedundantMoveElimination_NoOptimization_DifferentRegisters(t *testing.T) {
	opt := &RedundantMoveOptimization{}

	input := []*ast.Instruction{
		{
			Opcode: "addi",
			Operands: []ast.Expression{
				&ast.Identifier{Name: "t0"},
				&ast.Identifier{Name: "t1"},
				&ast.Number{Value: 0},
			},
		},
		{
			Opcode: "addi",
			Operands: []ast.Expression{
				&ast.Identifier{Name: "t2"},
				&ast.Identifier{Name: "t3"}, // Different source register
				&ast.Number{Value: 0},
			},
		},
	}

	_, consumed := opt.Optimize(input, 0)
	applied := consumed > 0

	if applied {
		t.Error("optimization should not be applied when registers don't match")
	}
}

func TestRedundantMoveElimination_NoOptimization_NotMoveInstruction(t *testing.T) {
	opt := &RedundantMoveOptimization{}

	input := []*ast.Instruction{
		{
			Opcode: "addi",
			Operands: []ast.Expression{
				&ast.Identifier{Name: "t0"},
				&ast.Identifier{Name: "t1"},
				&ast.Number{Value: 5}, // Not a move (imm != 0)
			},
		},
		{
			Opcode: "addi",
			Operands: []ast.Expression{
				&ast.Identifier{Name: "t2"},
				&ast.Identifier{Name: "t0"},
				&ast.Number{Value: 0},
			},
		},
	}

	_, consumed := opt.Optimize(input, 0)
	applied := consumed > 0

	if applied {
		t.Error("optimization should not be applied when first instruction is not a move")
	}
}

func TestOptimizerIntegration_ZeroConstant(t *testing.T) {
	opt := New()

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.Instruction{
				Opcode: "add",
				Operands: []ast.Expression{
					&ast.Identifier{Name: "t0"},
					&ast.Identifier{Name: "t1"},
					&ast.Identifier{Name: "zero"},
				},
			},
		},
	}

	result := opt.OptimizeProgram(program)

	if len(result.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(result.Statements))
	}

	instr, ok := result.Statements[0].(*ast.Instruction)
	if !ok {
		t.Fatal("expected instruction")
	}

	if instr.Opcode != "addi" {
		t.Errorf("expected opcode 'addi', got '%s'", instr.Opcode)
	}
}

func TestOptimizerIntegration_RedundantMove(t *testing.T) {
	got := optimizeTexts(t, []*ast.Instruction{
		ins("addi", reg("t0"), reg("t1"), num(0)),
		ins("addi", reg("t1"), reg("t0"), num(0)),
	})
	want := []string{"addi t0, t1, 0"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOptimizerIntegration_LabelBoundary(t *testing.T) {
	opt := New()

	// Optimizer should not combine instructions across labels
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.Instruction{
				Opcode: "addi",
				Operands: []ast.Expression{
					&ast.Identifier{Name: "t0"},
					&ast.Identifier{Name: "t1"},
					&ast.Number{Value: 0},
				},
			},
			&ast.Label{Name: "loop"},
			&ast.Instruction{
				Opcode: "addi",
				Operands: []ast.Expression{
					&ast.Identifier{Name: "t2"},
					&ast.Identifier{Name: "t0"},
					&ast.Number{Value: 0},
				},
			},
		},
	}

	result := opt.OptimizeProgram(program)

	if len(result.Statements) != 3 {
		t.Fatalf("expected 3 statements (instr, label, instr), got %d", len(result.Statements))
	}

	// First and third should still be separate instructions
	_, ok1 := result.Statements[0].(*ast.Instruction)
	_, ok2 := result.Statements[1].(*ast.Label)
	_, ok3 := result.Statements[2].(*ast.Instruction)

	if !ok1 || !ok2 || !ok3 {
		t.Error("expected instruction, label, instruction sequence")
	}
}

func TestOptimizerIntegration_DirectiveBoundary(t *testing.T) {
	opt := New()

	// Optimizer should not combine instructions across directives
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.Instruction{
				Opcode: "addi",
				Operands: []ast.Expression{
					&ast.Identifier{Name: "t0"},
					&ast.Identifier{Name: "t1"},
					&ast.Number{Value: 0},
				},
			},
			&ast.Directive{Name: ".align", Args: []ast.Expression{&ast.Number{Value: 4}}},
			&ast.Instruction{
				Opcode: "addi",
				Operands: []ast.Expression{
					&ast.Identifier{Name: "t2"},
					&ast.Identifier{Name: "t0"},
					&ast.Number{Value: 0},
				},
			},
		},
	}

	result := opt.OptimizeProgram(program)

	if len(result.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(result.Statements))
	}
}

func TestPatterns_IsZeroRegister(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected bool
	}{
		{
			name:     "x0 register",
			expr:     &ast.Identifier{Name: "x0"},
			expected: true,
		},
		{
			name:     "zero register",
			expr:     &ast.Identifier{Name: "zero"},
			expected: true,
		},
		{
			name:     "other register",
			expr:     &ast.Identifier{Name: "t0"},
			expected: false,
		},
		{
			name:     "number",
			expr:     &ast.Number{Value: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isZeroRegister(tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPatterns_RegistersEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        ast.Expression
		b        ast.Expression
		expected bool
	}{
		{
			name:     "same register",
			a:        &ast.Identifier{Name: "t0"},
			b:        &ast.Identifier{Name: "t0"},
			expected: true,
		},
		{
			name:     "different registers",
			a:        &ast.Identifier{Name: "t0"},
			b:        &ast.Identifier{Name: "t1"},
			expected: false,
		},
		{
			name:     "one not a register",
			a:        &ast.Identifier{Name: "t0"},
			b:        &ast.Number{Value: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registersEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// helpers for the sequence tests
func reg(name string) ast.Expression { return &ast.Identifier{Name: name} }
func num(v int64) ast.Expression     { return &ast.Number{Value: v} }
func ins(op string, operands ...ast.Expression) *ast.Instruction {
	return &ast.Instruction{Opcode: op, Operands: operands}
}

// instrText prints an instruction as "op a, b, c" for the comparisons
func instrText(instr *ast.Instruction) string {
	s := instr.Opcode
	for i, op := range instr.Operands {
		if i == 0 {
			s += " "
		} else {
			s += ", "
		}
		switch e := op.(type) {
		case *ast.Identifier:
			s += e.Name
		case *ast.Number:
			s += fmt.Sprint(e.Value)
		default:
			s += "?"
		}
	}
	return s
}

func optimizeTexts(t *testing.T, input []*ast.Instruction) []string {
	t.Helper()
	program := &ast.Program{}
	for _, in := range input {
		program.AddStatement(in)
	}
	var got []string
	for _, stmt := range New().OptimizeProgram(program).Statements {
		got = append(got, instrText(stmt.(*ast.Instruction)))
	}
	return got
}

// The instruction after a replacement must survive (the loop used to skip it)
func TestOptimizerKeepsInstructionAfterReplacement(t *testing.T) {
	tests := []struct {
		name  string
		input []*ast.Instruction
		want  []string
	}{
		{
			"zero operand then li, li",
			[]*ast.Instruction{
				ins("add", reg("t0"), reg("t1"), reg("zero")),
				ins("addi", reg("a0"), reg("x0"), num(1)),
				ins("addi", reg("a1"), reg("x0"), num(2)),
			},
			[]string{"addi t0, t1, 0", "addi a0, x0, 1", "addi a1, x0, 2"},
		},
		{
			"zero operand then move reading its result",
			[]*ast.Instruction{
				ins("add", reg("t0"), reg("t1"), reg("zero")),
				ins("addi", reg("t2"), reg("t0"), num(0)),
			},
			// t0 is still written: the move pair rule must not drop it
			[]string{"addi t0, t1, 0", "addi t2, t0, 0"},
		},
		{
			"two zero operand rules in a row",
			[]*ast.Instruction{
				ins("or", reg("a0"), reg("zero"), reg("a1")),
				ins("add", reg("a2"), reg("a3"), reg("x0")),
				ins("sub", reg("a4"), reg("a5"), reg("a6")),
			},
			[]string{"addi a0, a1, 0", "addi a2, a3, 0", "sub a4, a5, a6"},
		},
		{
			"move pair then next instruction",
			[]*ast.Instruction{
				ins("addi", reg("t0"), reg("t1"), num(0)),
				ins("addi", reg("t0"), reg("t0"), num(0)),
				ins("sub", reg("a4"), reg("a5"), reg("a6")),
			},
			[]string{"addi t0, t1, 0", "sub a4, a5, a6"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optimizeTexts(t, tt.input)
			if fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// mv x, y; mv z, x cannot become mv z, y: x would never be written
func TestRedundantMoveKeepsIntermediateRegister(t *testing.T) {
	got := optimizeTexts(t, []*ast.Instruction{
		ins("addi", reg("t0"), reg("t1"), num(0)),
		ins("addi", reg("t2"), reg("t0"), num(0)),
		ins("add", reg("a0"), reg("t0"), reg("t2")),
	})
	want := []string{"addi t0, t1, 0", "addi t2, t0, 0", "add a0, t0, t2"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
