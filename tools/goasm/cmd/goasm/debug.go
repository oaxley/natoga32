package main

import (
	"fmt"

	"github.com/natoga32/goasm/internal/ast"
)

// debugProgram prints the AST for debugging
func debugProgram(program *ast.Program) {
	for i, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.Label:
			fmt.Printf("%4d: LABEL %s\n", i, s.Name)
		case *ast.Instruction:
			fmt.Printf("%4d: INSTR %s", i, s.Opcode)
			for j, op := range s.Operands {
				if j > 0 {
					fmt.Print(",")
				}
				fmt.Printf(" %s", debugExpr(op))
			}
			fmt.Println()
		case *ast.Directive:
			fmt.Printf("%4d: DIR   %s", i, s.Name)
			for j, arg := range s.Args {
				if j > 0 {
					fmt.Print(",")
				}
				fmt.Printf(" %s", debugExpr(arg))
			}
			fmt.Println()
		}
	}
}

// debugExpr returns a string representation of an expression
func debugExpr(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Number:
		return fmt.Sprintf("%d", e.Value)
	case *ast.Identifier:
		return e.Name
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Text)
	case *ast.CharLiteral:
		return fmt.Sprintf("'%c'", e.Char)
	case *ast.CurrentPC:
		return "$"
	case *ast.BinaryOp:
		return fmt.Sprintf("(%s %s %s)", debugExpr(e.Left), e.Op, debugExpr(e.Right))
	case *ast.UnaryOp:
		return fmt.Sprintf("(%s%s)", e.Op, debugExpr(e.Operand))
	case *ast.HiRel:
		return fmt.Sprintf("%%hi(%s)", debugExpr(e.Expr))
	case *ast.LoRel:
		return fmt.Sprintf("%%lo(%s)", debugExpr(e.Expr))
	case *ast.PCRelHi:
		return fmt.Sprintf("%%pcrel_hi(%s)", debugExpr(e.Expr))
	case *ast.PCRelLo:
		return fmt.Sprintf("%%pcrel_lo(%s)", debugExpr(e.Expr))
	default:
		return "<?>"
	}
}
