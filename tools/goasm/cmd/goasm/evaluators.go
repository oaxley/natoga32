package main

import (
	"github.com/natoga32/goasm/internal/arch"
	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/symbol"
)

// evalImmediate evaluates an expression as an immediate value
func evalImmediate(expr ast.Expression, symbols *symbol.Table, pc int64) int32 {
	switch e := expr.(type) {
	case *ast.Number:
		return int32(e.Value)

	case *ast.Identifier:
		if sym := symbols.Lookup(e.Name); sym != nil {
			return int32(sym.Value)
		}
		return 0

	case *ast.HiRel:
		val := evalExpression(e.Expr, symbols, pc)
		return int32((val + 0x800) >> 12)

	case *ast.LoRel:
		val := evalExpression(e.Expr, symbols, pc)
		return int32(val & 0xFFF)

	case *ast.PCRelHi:
		val := evalExpression(e.Expr, symbols, pc)
		offset := val - pc
		return int32((offset + 0x800) >> 12)

	case *ast.PCRelLo:
		val := evalExpression(e.Expr, symbols, pc)
		offset := val - pc
		return int32(offset & 0xFFF)

	case *ast.UnaryOp:
		operand := evalImmediate(e.Operand, symbols, pc)
		if e.Op == "-" {
			return -operand
		}
		return operand

	case *ast.BinaryOp:
		left := evalImmediate(e.Left, symbols, pc)
		right := evalImmediate(e.Right, symbols, pc)
		switch e.Op {
		case "+":
			return left + right
		case "-":
			return left - right
		case "*":
			return left * right
		case "/":
			if right != 0 {
				return left / right
			}
			return 0
		case "&":
			return left & right
		case "|":
			return left | right
		case "^":
			return left ^ right
		case "<<":
			return left << uint(right)
		case ">>":
			return left >> uint(right)
		}
	}

	return 0
}

// evalExpression evaluates an expression to get its value
func evalExpression(expr ast.Expression, symbols *symbol.Table, pc int64) int64 {
	switch e := expr.(type) {
	case *ast.Number:
		return e.Value

	case *ast.Identifier:
		if sym := symbols.Lookup(e.Name); sym != nil {
			return sym.Value
		}
		return 0

	case *ast.CurrentPC:
		return pc

	case *ast.UnaryOp:
		operand := evalExpression(e.Operand, symbols, pc)
		if e.Op == "-" {
			return -operand
		}
		return operand

	case *ast.BinaryOp:
		left := evalExpression(e.Left, symbols, pc)
		right := evalExpression(e.Right, symbols, pc)
		switch e.Op {
		case "+":
			return left + right
		case "-":
			return left - right
		case "*":
			return left * right
		case "/":
			if right != 0 {
				return left / right
			}
			return 0
		}
	}

	return 0
}

// evalRegister evaluates an expression as a register number
func evalRegister(expr ast.Expression) uint32 {
	if ident, ok := expr.(*ast.Identifier); ok {
		if reg, ok := arch.LookupRegister(ident.Name); ok {
			return reg
		}
	}
	return 0
}
