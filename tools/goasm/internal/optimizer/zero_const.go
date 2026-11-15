// Zero constant optimization for the NATOGA32 assembler optimizer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package optimizer

import "github.com/natoga32/goasm/internal/ast"

// ZeroConstOptimization optimizes operations with zero register
// Patterns:
//   - add rd, rs, zero → mv rd, rs (addi rd, rs, 0)
//   - add rd, zero, rs → mv rd, rs (addi rd, rs, 0)
//   - or rd, rs, zero → mv rd, rs (addi rd, rs, 0)
//   - or rd, zero, rs → mv rd, rs (addi rd, rs, 0)
type ZeroConstOptimization struct{}

func (z *ZeroConstOptimization) Name() string {
	return "zero-constant"
}

// Optimize detects patterns that can use zero register more efficiently
func (z *ZeroConstOptimization) Optimize(instructions []*ast.Instruction, index int) ([]*ast.Instruction, bool) {
	if index >= len(instructions) {
		return nil, false
	}

	instr := instructions[index]

	// Pattern 1: add rd, rs, zero → mv rd, rs (addi rd, rs, 0)
	if instr.Opcode == "add" && len(instr.Operands) == 3 {
		rd := instr.Operands[0]
		rs1 := instr.Operands[1]
		rs2 := instr.Operands[2]

		// Check if right operand is zero
		if isZeroRegister(rs2) {
			// add rd, rs, zero → addi rd, rs, 0 (mv rd, rs)
			return []*ast.Instruction{
				{
					Opcode: "addi",
					Operands: []ast.Expression{
						rd,
						rs1,
						&ast.Number{Value: 0, Location: instr.Location},
					},
					Location: instr.Location,
				},
			}, true
		}

		// Check if left operand is zero
		if isZeroRegister(rs1) {
			// add rd, zero, rs → addi rd, rs, 0 (mv rd, rs)
			return []*ast.Instruction{
				{
					Opcode: "addi",
					Operands: []ast.Expression{
						rd,
						rs2,
						&ast.Number{Value: 0, Location: instr.Location},
					},
					Location: instr.Location,
				},
			}, true
		}
	}

	// Pattern 2: or rd, rs, zero → mv rd, rs (addi rd, rs, 0)
	if instr.Opcode == "or" && len(instr.Operands) == 3 {
		rd := instr.Operands[0]
		rs1 := instr.Operands[1]
		rs2 := instr.Operands[2]

		// Check if right operand is zero
		if isZeroRegister(rs2) {
			return []*ast.Instruction{
				{
					Opcode: "addi",
					Operands: []ast.Expression{
						rd,
						rs1,
						&ast.Number{Value: 0, Location: instr.Location},
					},
					Location: instr.Location,
				},
			}, true
		}

		// Check if left operand is zero
		if isZeroRegister(rs1) {
			return []*ast.Instruction{
				{
					Opcode: "addi",
					Operands: []ast.Expression{
						rd,
						rs2,
						&ast.Number{Value: 0, Location: instr.Location},
					},
					Location: instr.Location,
				},
			}, true
		}
	}

	return nil, false
}
