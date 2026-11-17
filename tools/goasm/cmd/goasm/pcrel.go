package main

import (
	"fmt"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/symbol"
)

// encodeContext carries what %pcrel_lo needs to find its paired auipc
type encodeContext struct {
	// auipcs maps the address of every "auipc rd, %pcrel_hi(...)" to it
	auipcs map[int64]*ast.Instruction
	// prev is the previously encoded instruction (nil after a directive)
	prev     *ast.Instruction
	prevAddr int64
}

// newEncodeContext indexes the %pcrel_hi auipcs of a program whose
// instruction addresses were set by semanticPass1
func newEncodeContext(program *ast.Program) *encodeContext {
	ctx := &encodeContext{auipcs: make(map[int64]*ast.Instruction)}
	for _, stmt := range program.Statements {
		if in, ok := stmt.(*ast.Instruction); ok && pcrelHiOf(in) != nil {
			ctx.auipcs[int64(in.Address)] = in
		}
	}
	return ctx
}

// pcrelHiOf returns the %pcrel_hi operand of an auipc, or nil
func pcrelHiOf(in *ast.Instruction) *ast.PCRelHi {
	if in == nil || in.Opcode != "auipc" || len(in.Operands) < 2 {
		return nil
	}
	hi, _ := in.Operands[1].(*ast.PCRelHi)
	return hi
}

// evalPCRelLo resolves %pcrel_lo(expr) for the instruction at pc.
//
// The low part is relative to the paired auipc, found in this order:
//  1. the immediately preceding instruction, if it is an auipc with
//     %pcrel_hi of the same expression (la, call, hand-written adjacent pairs);
//  2. the auipc labelled expr (GNU style: "1: auipc t0, %pcrel_hi(sym)"
//     then "addi t0, t0, %pcrel_lo(1b)"), anywhere in the program.
func evalPCRelLo(lo *ast.PCRelLo, symbols *symbol.Table, pc int64, ctx *encodeContext) (int32, error) {
	if hi := pcrelHiOf(ctx.prev); hi != nil && ctx.prevAddr == pc-4 && debugExpr(hi.Expr) == debugExpr(lo.Expr) {
		return pcrelLoBits(hi, symbols, ctx.prevAddr)
	}

	if ident, ok := lo.Expr.(*ast.Identifier); ok {
		if sym := symbols.Lookup(ident.Name); sym != nil && sym.Defined {
			if auipc, ok := ctx.auipcs[sym.Value]; ok {
				return pcrelLoBits(pcrelHiOf(auipc), symbols, sym.Value)
			}
		}
	}

	return 0, fmt.Errorf("%s: error: %%pcrel_lo(%s) has no paired auipc: it must follow \"auipc rd, %%pcrel_hi(%s)\" or name the label of an auipc",
		lo.Location, debugExpr(lo.Expr), debugExpr(lo.Expr))
}

// pcrelLoBits returns the low 12 bits of the offset computed by the auipc at auipcAddr
func pcrelLoBits(hi *ast.PCRelHi, symbols *symbol.Table, auipcAddr int64) (int32, error) {
	target, err := evalExpressionWithError(hi.Expr, symbols, auipcAddr)
	if err != nil {
		return 0, err
	}
	return int32((target - auipcAddr) & 0xFFF), nil
}

// evalIImmediate evaluates an I/S-type immediate, resolving %pcrel_lo
func evalIImmediate(expr ast.Expression, symbols *symbol.Table, pc int64, ctx *encodeContext) (int32, error) {
	if lo, ok := expr.(*ast.PCRelLo); ok {
		return evalPCRelLo(lo, symbols, pc, ctx)
	}
	return evalImmediateWithError(expr, symbols, pc)
}
