// Unit tests for the NATOGA32 assembler optimizer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package optimizer

import (
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

	result, applied := opt.Optimize(input, 0)

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

	result, applied := opt.Optimize(input, 0)

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

	result, applied := opt.Optimize(input, 0)

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

	_, applied := opt.Optimize(input, 0)

	if applied {
		t.Error("optimization should not be applied for add without zero")
	}
}

func TestRedundantMoveElimination_ConsecutiveMoves(t *testing.T) {
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
				&ast.Identifier{Name: "t0"},
				&ast.Number{Value: 0},
			},
		},
	}

	result, applied := opt.Optimize(input, 0)

	if !applied {
		t.Fatal("expected optimization to be applied")
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(result))
	}

	if result[0].Opcode != "addi" {
		t.Errorf("expected opcode 'addi', got '%s'", result[0].Opcode)
	}

	rd, _ := getRegisterName(result[0].Operands[0])
	rs, _ := getRegisterName(result[0].Operands[1])
	imm, _ := getNumberValue(result[0].Operands[2])

	if rd != "t2" || rs != "t1" || imm != 0 {
		t.Errorf("unexpected operands: rd=%s, rs=%s, imm=%d", rd, rs, imm)
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

	_, applied := opt.Optimize(input, 0)

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

	_, applied := opt.Optimize(input, 0)

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
	opt := New()

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

	if len(result.Statements) != 1 {
		t.Fatalf("expected 1 statement after optimization, got %d", len(result.Statements))
	}

	instr, ok := result.Statements[0].(*ast.Instruction)
	if !ok {
		t.Fatal("expected instruction")
	}

	rd, _ := getRegisterName(instr.Operands[0])
	rs, _ := getRegisterName(instr.Operands[1])

	if rd != "t2" || rs != "t1" {
		t.Errorf("expected mv t2, t1 (addi t2, t1, 0), got rd=%s, rs=%s", rd, rs)
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
