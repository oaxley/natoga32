// Redundant move elimination for the NATOGA32 assembler optimizer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package optimizer

import "github.com/natoga32/goasm/internal/ast"

// RedundantMoveOptimization eliminates redundant register moves
// Pattern: mv x, y; mv z, x → mv z, y (when x not used elsewhere)
// Note: After pseudo-expansion, mv is represented as addi rd, rs, 0
type RedundantMoveOptimization struct{}

func (r *RedundantMoveOptimization) Name() string {
	return "redundant-move"
}

// Optimize detects consecutive move instructions that can be combined
func (r *RedundantMoveOptimization) Optimize(instructions []*ast.Instruction, index int) ([]*ast.Instruction, bool) {
	if index+1 >= len(instructions) {
		return nil, false
	}

	instr1 := instructions[index]
	instr2 := instructions[index+1]

	// Both must be move instructions (addi with immediate 0)
	if !isMoveInstruction(instr1) || !isMoveInstruction(instr2) {
		return nil, false
	}

	// Pattern: addi x, y, 0; addi z, x, 0
	// Equivalent to: mv x, y; mv z, x
	rd1 := instr1.Operands[0]  // destination of first move (x)
	rs1 := instr1.Operands[1]  // source of first move (y)
	rd2 := instr2.Operands[0]  // destination of second move (z)
	rs2 := instr2.Operands[1]  // source of second move (x)

	// Check if first destination is second source (x == x)
	if registersEqual(rd1, rs2) {
		// Don't optimize if source and destination are the same in first move
		if registersEqual(rd1, rs1) {
			return nil, false
		}

		// Optimization: mv z, y (combine the moves)
		// Result: addi z, y, 0
		return []*ast.Instruction{
			{
				Opcode: "addi",
				Operands: []ast.Expression{
					rd2, // destination from second move (z)
					rs1, // source from first move (y)
					&ast.Number{Value: 0, Location: instr2.Location},
				},
				Location: instr2.Location,
			},
		}, true
	}

	return nil, false
}

// isMoveInstruction checks if an instruction is a move (addi rd, rs, 0)
func isMoveInstruction(instr *ast.Instruction) bool {
	if instr.Opcode != "addi" || len(instr.Operands) != 3 {
		return false
	}
	// Check immediate is 0
	return isNumberValue(instr.Operands[2], 0)
}
