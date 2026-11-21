// AST node interfaces for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package ast

// Node is the base interface for all AST nodes
type Node interface {
	node()
}

// Statement represents a statement in the program
type Statement interface {
	Node
	statement()
}

// Expression represents an expression (operand)
type Expression interface {
	Node
	expression()
}

// Program is the root AST node containing all statements
type Program struct {
	Statements []Statement
}

func (p *Program) node() {}

// AddStatement adds a statement to the program
func (p *Program) AddStatement(stmt Statement) {
	p.Statements = append(p.Statements, stmt)
}
