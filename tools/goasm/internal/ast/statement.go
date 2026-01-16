// AST statement types for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package ast

// Label represents a label definition (e.g., "main:")
type Label struct {
	Name string
}

func (l *Label) node()      {}
func (l *Label) statement() {}

// Instruction represents an assembly instruction (e.g., "addi x1, x0, 10")
type Instruction struct {
	Opcode   string
	Operands []Expression
	Address  int // Set during semantic analysis
}

func (i *Instruction) node()      {}
func (i *Instruction) statement() {}

// Directive represents an assembler directive (e.g., ".text", ".byte 0x00")
type Directive struct {
	Name string
	Args []Expression
}

func (d *Directive) node()      {}
func (d *Directive) statement() {}
