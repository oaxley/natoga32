package main

import (
	"fmt"

	"github.com/natoga32/goasm/internal/arch"
	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/symbol"
)

// getExprLocation returns the location of an expression
func getExprLocation(expr ast.Expression) ast.SourceLocation {
	switch e := expr.(type) {
	case *ast.Number:
		return e.Location
	case *ast.Identifier:
		return e.Location
	case *ast.StringLiteral:
		return e.Location
	case *ast.CharLiteral:
		return e.Location
	case *ast.CurrentPC:
		return e.Location
	case *ast.BinaryOp:
		return e.Location
	case *ast.UnaryOp:
		return e.Location
	case *ast.HiRel:
		return e.Location
	case *ast.LoRel:
		return e.Location
	case *ast.PCRelHi:
		return e.Location
	case *ast.PCRelLo:
		return e.Location
	}
	return ast.SourceLocation{}
}

// evalImmediateWithError evaluates an expression as an immediate value with error reporting
func evalImmediateWithError(expr ast.Expression, symbols *symbol.Table, pc int64) (int32, error) {
	switch e := expr.(type) {
	case *ast.Number:
		return int32(e.Value), nil

	case *ast.Identifier:
		if sym := symbols.Lookup(e.Name); sym != nil {
			return int32(sym.Value), nil
		}
		return 0, fmt.Errorf("%s: error: undefined symbol '%s'", e.Location, e.Name)

	case *ast.HiRel:
		val, err := evalExpressionWithError(e.Expr, symbols, pc)
		if err != nil {
			return 0, err
		}
		return int32((val + 0x800) >> 12), nil

	case *ast.LoRel:
		val, err := evalExpressionWithError(e.Expr, symbols, pc)
		if err != nil {
			return 0, err
		}
		return int32(val & 0xFFF), nil

	case *ast.PCRelHi:
		val, err := evalExpressionWithError(e.Expr, symbols, pc)
		if err != nil {
			return 0, err
		}
		offset := val - pc
		return int32((offset + 0x800) >> 12), nil

	case *ast.PCRelLo:
		val, err := evalExpressionWithError(e.Expr, symbols, pc)
		if err != nil {
			return 0, err
		}
		offset := val - pc
		return int32(offset & 0xFFF), nil

	case *ast.UnaryOp:
		operand, err := evalImmediateWithError(e.Operand, symbols, pc)
		if err != nil {
			return 0, err
		}
		if e.Op == "-" {
			return -operand, nil
		}
		return operand, nil

	case *ast.BinaryOp:
		left, err := evalImmediateWithError(e.Left, symbols, pc)
		if err != nil {
			return 0, err
		}
		right, err := evalImmediateWithError(e.Right, symbols, pc)
		if err != nil {
			return 0, err
		}
		switch e.Op {
		case "+":
			return left + right, nil
		case "-":
			return left - right, nil
		case "*":
			return left * right, nil
		case "/":
			if right != 0 {
				return left / right, nil
			}
			return 0, fmt.Errorf("%s: error: division by zero", e.Location)
		case "&":
			return left & right, nil
		case "|":
			return left | right, nil
		case "^":
			return left ^ right, nil
		case "<<":
			return left << uint(right), nil
		case ">>":
			return left >> uint(right), nil
		}
	}

	return 0, nil
}

// evalExpressionWithError evaluates an expression to get its value with error reporting
func evalExpressionWithError(expr ast.Expression, symbols *symbol.Table, pc int64) (int64, error) {
	switch e := expr.(type) {
	case *ast.Number:
		return e.Value, nil

	case *ast.Identifier:
		if sym := symbols.Lookup(e.Name); sym != nil {
			return sym.Value, nil
		}
		return 0, fmt.Errorf("%s: error: undefined symbol '%s'", e.Location, e.Name)

	case *ast.CurrentPC:
		return pc, nil

	case *ast.UnaryOp:
		operand, err := evalExpressionWithError(e.Operand, symbols, pc)
		if err != nil {
			return 0, err
		}
		if e.Op == "-" {
			return -operand, nil
		}
		return operand, nil

	case *ast.BinaryOp:
		left, err := evalExpressionWithError(e.Left, symbols, pc)
		if err != nil {
			return 0, err
		}
		right, err := evalExpressionWithError(e.Right, symbols, pc)
		if err != nil {
			return 0, err
		}
		switch e.Op {
		case "+":
			return left + right, nil
		case "-":
			return left - right, nil
		case "*":
			return left * right, nil
		case "/":
			if right != 0 {
				return left / right, nil
			}
			return 0, fmt.Errorf("%s: error: division by zero", e.Location)
		}
	}

	return 0, nil
}

// evalImmediate evaluates an expression as an immediate value (legacy, no error reporting)
func evalImmediate(expr ast.Expression, symbols *symbol.Table, pc int64) int32 {
	val, _ := evalImmediateWithError(expr, symbols, pc)
	return val
}

// evalExpression evaluates an expression to get its value (legacy, no error reporting)
func evalExpression(expr ast.Expression, symbols *symbol.Table, pc int64) int64 {
	val, _ := evalExpressionWithError(expr, symbols, pc)
	return val
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
