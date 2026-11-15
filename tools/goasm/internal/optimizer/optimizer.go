// Peephole optimizer for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package optimizer

import "github.com/natoga32/goasm/internal/ast"

// OptimizationPass represents a single optimization
type OptimizationPass interface {
	Name() string
	// Optimize attempts to optimize starting at the given index
	// Returns replacement instructions and whether optimization was applied
	// If applied, the optimizer will skip the next instruction
	Optimize(instructions []*ast.Instruction, index int) ([]*ast.Instruction, bool)
}

// Optimizer performs peephole optimizations on AST
type Optimizer struct {
	optimizations []OptimizationPass
}

// New creates a new optimizer with all enabled passes
func New() *Optimizer {
	return &Optimizer{
		optimizations: []OptimizationPass{
			&ZeroConstOptimization{},
			&RedundantMoveOptimization{},
		},
	}
}

// OptimizeProgram runs all optimization passes on the program
func (o *Optimizer) OptimizeProgram(program *ast.Program) *ast.Program {
	result := &ast.Program{
		Statements: make([]ast.Statement, 0, len(program.Statements)),
	}

	// Collect instructions and non-instructions separately
	var instructions []*ast.Instruction

	for _, stmt := range program.Statements {
		if instr, ok := stmt.(*ast.Instruction); ok {
			instructions = append(instructions, instr)
		} else {
			// Non-instruction (label, directive) - flush pending instructions
			result = flushInstructions(result, instructions, o)
			instructions = nil
			result.AddStatement(stmt)
		}
	}

	// Flush remaining instructions
	result = flushInstructions(result, instructions, o)
	return result
}

// flushInstructions optimizes a sequence of instructions and adds to result
func flushInstructions(result *ast.Program, instructions []*ast.Instruction, o *Optimizer) *ast.Program {
	optimized := o.optimizeSequence(instructions)
	for _, instr := range optimized {
		result.AddStatement(instr)
	}
	return result
}

// optimizeSequence applies all optimization passes to an instruction sequence
func (o *Optimizer) optimizeSequence(instructions []*ast.Instruction) []*ast.Instruction {
	result := make([]*ast.Instruction, 0, len(instructions))
	i := 0

	for i < len(instructions) {
		optimized := false

		// Try each optimization pass
		for _, pass := range o.optimizations {
			replacement, applied := pass.Optimize(instructions, i)
			if applied {
				result = append(result, replacement...)
				i += 2 // Most peephole optimizations look at 2 instructions
				optimized = true
				break
			}
		}

		// No optimization applied, keep original
		if !optimized {
			result = append(result, instructions[i])
			i++
		}
	}

	return result
}
