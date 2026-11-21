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
	case *ast.OffsetBase:
		return e.Location
	}
	return ast.SourceLocation{}
}

// evalImmediateWithError evaluates an expression as a 32-bit immediate value.
// It is evalExpressionWithError truncated to 32 bits; %pcrel_lo is refused
// here because only evalIImmediate knows the paired auipc.
func evalImmediateWithError(expr ast.Expression, symbols *symbol.Table, pc int64) (int32, error) {
	if lo, ok := expr.(*ast.PCRelLo); ok {
		return 0, fmt.Errorf("%s: error: %%pcrel_lo must be the whole immediate of an I-type or S-type instruction", lo.Location)
	}
	val, err := evalExpressionWithError(expr, symbols, pc)
	return int32(val), err
}

// evalExpressionWithError evaluates an expression in 64 bits. pc is the value of $.
// Every operator of the parser is handled; anything else is an error, never 0.
func evalExpressionWithError(expr ast.Expression, symbols *symbol.Table, pc int64) (int64, error) {
	switch e := expr.(type) {
	case *ast.Number:
		return e.Value, nil

	case *ast.CharLiteral:
		return int64(e.Char), nil

	case *ast.Identifier:
		// a name only listed in .global has no value yet: it is undefined
		if sym := symbols.Lookup(e.Name); sym != nil && sym.Defined {
			return sym.Value, nil
		}
		return 0, fmt.Errorf("%s: error: undefined symbol '%s'", e.Location, e.Name)

	case *ast.CurrentPC:
		return pc, nil

	case *ast.HiRel:
		val, err := evalExpressionWithError(e.Expr, symbols, pc)
		if err != nil {
			return 0, err
		}
		return (val + 0x800) >> 12, nil

	case *ast.LoRel:
		val, err := evalExpressionWithError(e.Expr, symbols, pc)
		if err != nil {
			return 0, err
		}
		return val & 0xFFF, nil

	case *ast.PCRelHi:
		val, err := evalExpressionWithError(e.Expr, symbols, pc)
		if err != nil {
			return 0, err
		}
		return (val - pc + 0x800) >> 12, nil

	case *ast.PCRelLo:
		return 0, fmt.Errorf("%s: error: %%pcrel_lo must be the whole immediate of an I-type or S-type instruction", e.Location)

	case *ast.UnaryOp:
		operand, err := evalExpressionWithError(e.Operand, symbols, pc)
		if err != nil {
			return 0, err
		}
		switch e.Op {
		case "-":
			return -operand, nil
		case "+":
			return operand, nil
		case "~":
			return ^operand, nil
		}
		return 0, fmt.Errorf("%s: error: unknown unary operator '%s'", e.Location, e.Op)

	case *ast.BinaryOp:
		left, err := evalExpressionWithError(e.Left, symbols, pc)
		if err != nil {
			return 0, err
		}
		right, err := evalExpressionWithError(e.Right, symbols, pc)
		if err != nil {
			return 0, err
		}
		return evalBinaryOp(e, left, right)
	}

	return 0, fmt.Errorf("%s: error: %s cannot be used in an expression", getExprLocation(expr), exprKind(expr))
}

// evalBinaryOp applies a binary operator to two 64-bit values
func evalBinaryOp(e *ast.BinaryOp, left, right int64) (int64, error) {
	switch e.Op {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, fmt.Errorf("%s: error: division by zero", e.Location)
		}
		return left / right, nil
	case "%":
		if right == 0 {
			return 0, fmt.Errorf("%s: error: modulo by zero", e.Location)
		}
		return left % right, nil
	case "&":
		return left & right, nil
	case "|":
		return left | right, nil
	case "^":
		return left ^ right, nil
	case "<<", ">>":
		if right < 0 || right > 63 {
			return 0, fmt.Errorf("%s: error: shift amount %d out of range (0..63)", e.Location, right)
		}
		if e.Op == "<<" {
			return left << uint(right), nil
		}
		return left >> uint(right), nil // arithmetic shift: -8 >> 1 is -4
	}
	return 0, fmt.Errorf("%s: error: unknown operator '%s'", e.Location, e.Op)
}

// exprKind names an expression that is not a value, for error messages
func exprKind(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return fmt.Sprintf("string %q", e.Text)
	case *ast.OffsetBase:
		return "offset(register)"
	}
	return fmt.Sprintf("%T", expr)
}

// evalRegister evaluates a register operand of instr: x0..x31 or an ABI name
// (case-insensitive). Anything else is an error, never a silent x0.
func evalRegister(instr *ast.Instruction, expr ast.Expression) (uint32, error) {
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		return 0, fmt.Errorf("%s: error: %s: expected a register", instr.Location, instr.Opcode)
	}
	if reg, ok := arch.LookupRegister(ident.Name); ok {
		return reg, nil
	}
	// operands made by a pseudo-instruction expansion have no location
	loc := ident.Location
	if loc.Line == 0 {
		loc = instr.Location
	}
	return 0, fmt.Errorf("%s: error: '%s' is not a register", loc, ident.Name)
}
