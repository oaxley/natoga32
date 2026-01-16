// Pseudo-instruction expansion for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package pseudo

import (
	"github.com/natoga32/goasm/internal/arch"
	"github.com/natoga32/goasm/internal/ast"
)

// Expander expands pseudo-instructions into real RISC-V instructions
type Expander struct{}

// New creates a new pseudo-instruction expander
func New() *Expander {
	return &Expander{}
}

// ExpandProgram expands all pseudo-instructions in a program
func (e *Expander) ExpandProgram(program *ast.Program) *ast.Program {
	result := &ast.Program{
		Statements: make([]ast.Statement, 0, len(program.Statements)),
	}

	for _, stmt := range program.Statements {
		if instr, ok := stmt.(*ast.Instruction); ok {
			expanded := e.ExpandInstruction(instr)
			for _, exp := range expanded {
				result.AddStatement(exp)
			}
		} else {
			result.AddStatement(stmt)
		}
	}

	return result
}

// ExpandInstruction expands a single instruction if it's a pseudo-instruction
func (e *Expander) ExpandInstruction(instr *ast.Instruction) []*ast.Instruction {
	if !arch.IsPseudoInstruction(instr.Opcode) {
		return []*ast.Instruction{instr}
	}

	switch instr.Opcode {
	case "li":
		return e.expandLi(instr)
	case "la":
		return e.expandLa(instr)
	case "mv":
		return e.expandMv(instr)
	case "nop":
		return e.expandNop()
	case "ret":
		return e.expandRet()
	case "call":
		return e.expandCall(instr)
	case "j":
		return e.expandJ(instr)
	case "jr":
		return e.expandJr(instr)
	case "seqz":
		return e.expandSeqz(instr)
	case "snez":
		return e.expandSnez(instr)
	case "sltz":
		return e.expandSltz(instr)
	case "sgtz":
		return e.expandSgtz(instr)
	case "not":
		return e.expandNot(instr)
	case "neg":
		return e.expandNeg(instr)
	case "beqz":
		return e.expandBeqz(instr)
	case "bnez":
		return e.expandBnez(instr)
	case "blez":
		return e.expandBlez(instr)
	case "bgez":
		return e.expandBgez(instr)
	case "bltz":
		return e.expandBltz(instr)
	case "bgtz":
		return e.expandBgtz(instr)
	case "bgt":
		return e.expandBgt(instr)
	case "ble":
		return e.expandBle(instr)
	case "bgtu":
		return e.expandBgtu(instr)
	case "bleu":
		return e.expandBleu(instr)
	case "csrr":
		return e.expandCsrr(instr)
	case "csrw":
		return e.expandCsrw(instr)
	case "csrs":
		return e.expandCsrs(instr)
	case "csrc":
		return e.expandCsrc(instr)
	case "csrwi":
		return e.expandCsrwi(instr)
	case "csrsi":
		return e.expandCsrsi(instr)
	case "csrci":
		return e.expandCsrci(instr)
	case "syscall":
		return e.expandSyscall(instr)
	default:
		// Unknown pseudo-instruction, return as-is
		return []*ast.Instruction{instr}
	}
}

// Helper functions for creating common expressions
func x(n int) *ast.Identifier {
	return &ast.Identifier{Name: regName(n)}
}

func regName(n int) string {
	if n < 10 {
		return "x" + string('0'+byte(n))
	}
	return "x" + string('0'+byte(n/10)) + string('0'+byte(n%10))
}

func zero() *ast.Number {
	return &ast.Number{Value: 0}
}

func num(v int64) *ast.Number {
	return &ast.Number{Value: v}
}

func instr(opcode string, operands ...ast.Expression) *ast.Instruction {
	return &ast.Instruction{
		Opcode:   opcode,
		Operands: operands,
	}
}
