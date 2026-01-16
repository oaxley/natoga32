// AST expression types for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package ast

// Number represents a numeric literal
type Number struct {
	Value int64
}

func (n *Number) node()       {}
func (n *Number) expression() {}

// Identifier represents a symbol reference
type Identifier struct {
	Name string
}

func (i *Identifier) node()       {}
func (i *Identifier) expression() {}

// StringLiteral represents a string literal
type StringLiteral struct {
	Text string
}

func (s *StringLiteral) node()       {}
func (s *StringLiteral) expression() {}

// CharLiteral represents a character literal
type CharLiteral struct {
	Char rune
}

func (c *CharLiteral) node()       {}
func (c *CharLiteral) expression() {}

// CurrentPC represents the current program counter ($)
type CurrentPC struct{}

func (c *CurrentPC) node()       {}
func (c *CurrentPC) expression() {}

// BinaryOp represents a binary operation (e.g., a + b)
type BinaryOp struct {
	Op    string // "+", "-", "*", "/", etc.
	Left  Expression
	Right Expression
}

func (b *BinaryOp) node()       {}
func (b *BinaryOp) expression() {}

// UnaryOp represents a unary operation (e.g., -x)
type UnaryOp struct {
	Op    string // "+" or "-"
	Operand Expression
}

func (u *UnaryOp) node()       {}
func (u *UnaryOp) expression() {}

// Relocation types for address calculations

// RelocExpr is the base interface for relocation expressions
type RelocExpr interface {
	Expression
	reloc()
}

// HiRel represents %hi(expr) - absolute high 20 bits
type HiRel struct {
	Expr Expression
}

func (h *HiRel) node()       {}
func (h *HiRel) expression() {}
func (h *HiRel) reloc()      {}

// LoRel represents %lo(expr) - absolute low 12 bits
type LoRel struct {
	Expr Expression
}

func (l *LoRel) node()       {}
func (l *LoRel) expression() {}
func (l *LoRel) reloc()      {}

// PCRelHi represents %pcrel_hi(expr) - PC-relative high 20 bits
type PCRelHi struct {
	Expr Expression
}

func (p *PCRelHi) node()       {}
func (p *PCRelHi) expression() {}
func (p *PCRelHi) reloc()      {}

// PCRelLo represents %pcrel_lo(expr) - PC-relative low 12 bits
type PCRelLo struct {
	Expr Expression
}

func (p *PCRelLo) node()       {}
func (p *PCRelLo) expression() {}
func (p *PCRelLo) reloc()      {}
