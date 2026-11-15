// Pattern matching utilities for the NATOGA32 assembler optimizer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package optimizer

import "github.com/natoga32/goasm/internal/ast"

// isRegister checks if an expression is a register identifier with a specific name
func isRegister(expr ast.Expression, name string) bool {
	if ident, ok := expr.(*ast.Identifier); ok {
		return ident.Name == name
	}
	return false
}

// getRegisterName returns the register name if expression is an identifier
func getRegisterName(expr ast.Expression) (string, bool) {
	if ident, ok := expr.(*ast.Identifier); ok {
		return ident.Name, true
	}
	return "", false
}

// isNumberValue checks if an expression is a specific number
func isNumberValue(expr ast.Expression, value int64) bool {
	if num, ok := expr.(*ast.Number); ok {
		return num.Value == value
	}
	return false
}

// getNumberValue returns the number value if expression is a number
func getNumberValue(expr ast.Expression) (int64, bool) {
	if num, ok := expr.(*ast.Number); ok {
		return num.Value, true
	}
	return 0, false
}

// registersEqual checks if two expressions refer to the same register
func registersEqual(a, b ast.Expression) bool {
	regA, okA := getRegisterName(a)
	regB, okB := getRegisterName(b)
	return okA && okB && regA == regB
}

// isZeroRegister checks if an expression is the zero register (x0 or zero)
func isZeroRegister(expr ast.Expression) bool {
	return isRegister(expr, "x0") || isRegister(expr, "zero")
}
