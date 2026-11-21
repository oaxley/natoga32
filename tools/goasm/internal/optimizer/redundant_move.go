// Redundant move elimination for the NATOGA32 assembler optimizer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package optimizer

import "github.com/natoga32/goasm/internal/ast"

// RedundantMoveOptimization removes a move that changes nothing because of the move before it.
// Patterns (mv is addi rd, rs, 0 after pseudo-expansion):
//   - mv x, y; mv y, x → mv x, y (y already holds x)
//   - mv x, y; mv x, x → mv x, y (copies x onto itself)
//
// mv x, y; mv z, x is NOT changed into mv z, y: the assembler cannot know
// whether a later instruction reads x, so x must still be written.
// Nothing is done when x is x0: x0 stays 0, so mv y, x0 writes 0 into y.
type RedundantMoveOptimization struct{}

func (r *RedundantMoveOptimization) Name() string {
	return "redundant-move"
}

// Optimize keeps the first move of a pair when the second one has no effect
func (r *RedundantMoveOptimization) Optimize(instructions []*ast.Instruction, index int) ([]*ast.Instruction, int) {
	if index+1 >= len(instructions) {
		return nil, 0
	}

	instr1 := instructions[index]
	instr2 := instructions[index+1]

	// Both must be move instructions (addi with immediate 0)
	if !isMoveInstruction(instr1) || !isMoveInstruction(instr2) {
		return nil, 0
	}

	rd1 := instr1.Operands[0] // x
	rs1 := instr1.Operands[1] // y
	rd2 := instr2.Operands[0]
	rs2 := instr2.Operands[1]

	// the second move must read x, and x must be a real register
	if !registersEqual(rd1, rs2) || isZeroRegister(rd1) {
		return nil, 0
	}

	// mv y, x (back to the source) or mv x, x: no effect
	if registersEqual(rd2, rs1) || registersEqual(rd2, rd1) {
		return []*ast.Instruction{instr1}, 2
	}

	return nil, 0
}

// isMoveInstruction checks if an instruction is a move (addi rd, rs, 0)
func isMoveInstruction(instr *ast.Instruction) bool {
	if instr.Opcode != "addi" || len(instr.Operands) != 3 {
		return false
	}
	// Check immediate is 0
	return isNumberValue(instr.Operands[2], 0)
}
