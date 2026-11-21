package main

import (
	"fmt"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

// li with a name or an expression is expanded to lui %hi + addi %lo, since
// the value is not known before pass 1. relaxLi then makes it one
// instruction when the value allows it, like li with a number: addi alone
// for -2048..2047, lui alone when the low 12 bits are 0.
//
// Each such li starts as one instruction. When its value does not fit, it
// becomes two instructions for good, which moves the labels below it, so
// pass 1 runs again. A li only grows, so the loop ends, and in the final
// layout every short li has a value that fits it.

type liForm int

const (
	liLong liForm = iota // lui + addi
	liAddi               // addi rd, zero, value
	liLui                // lui rd, value >> 12
)

// liRecord is one li expanded with %hi/%lo
type liRecord struct {
	lui, addi *ast.Instruction
	short     *ast.Instruction // the one instruction of the short forms
	form      liForm
	value     int64 // the value at the last pass 1
}

// relaxLi runs pass 1 until the li sizes are stable, and returns its
// symbols and sections
func relaxLi(program *ast.Program, newSections func() *section.Manager) (*symbol.Table, *section.Manager, error) {
	base := program.Statements
	records := findLiPairs(base)
	program.Statements = rebuildStatements(base, records)
	for {
		symbols, sections := symbol.New(), newSections()
		if err := semanticPass1(program, symbols, sections); err != nil {
			return nil, nil, fmt.Errorf("semantic pass 1 error: %v", err)
		}
		changed, tooBig := chooseLiForms(records, symbols)
		// rebuild every time: with the same forms the layout is the same,
		// but the short instructions get the values of this pass
		program.Statements = rebuildStatements(base, records)
		if !changed {
			if tooBig != nil {
				return nil, nil, fmt.Errorf("semantic pass 1 error: %s: error: li value %d (0x%X) does not fit 32 bits",
					tooBig.lui.Location, tooBig.value, tooBig.value)
			}
			return symbols, sections, nil
		}
	}
}

// findLiPairs finds the lui %hi + addi %lo pairs made by li, by index of the lui
func findLiPairs(stmts []ast.Statement) map[int]*liRecord {
	records := make(map[int]*liRecord)
	for i := 0; i+1 < len(stmts); i++ {
		lui, ok1 := stmts[i].(*ast.Instruction)
		addi, ok2 := stmts[i+1].(*ast.Instruction)
		if !ok1 || !ok2 || lui.Pseudo != "li" || addi.Pseudo != "li" ||
			lui.Opcode != "lui" || addi.Opcode != "addi" || len(lui.Operands) != 2 || len(addi.Operands) != 3 {
			continue
		}
		hi, ok := lui.Operands[1].(*ast.HiRel)
		if !ok || usesCurrentPC(hi.Expr) {
			continue // li rd, $...: $ is the address of the addi
		}
		r := &liRecord{lui: lui, addi: addi, form: liAddi}
		r.setShort()
		records[i] = r
		i++
	}
	return records
}

// chooseLiForms picks the form of each li from its value. It reports
// whether a li grew, so pass 1 must run again, and a li whose value does
// not fit 32 bits (an error once the layout is final).
func chooseLiForms(records map[int]*liRecord, symbols *symbol.Table) (changed bool, tooBig *liRecord) {
	for _, r := range records {
		if r.form == liLong {
			continue // grown for good
		}
		expr := r.lui.Operands[1].(*ast.HiRel).Expr
		v, err := evalExpressionWithError(expr, symbols, int64(r.short.Address)+4)
		if err != nil {
			// keep the li as written: pass 2 reports the error
			r.form, changed = liLong, true
			r.setShort()
			continue
		}
		r.value = v
		if v < -0x80000000 || v > 0xFFFFFFFF {
			if tooBig == nil || r.lui.Location.Line < tooBig.lui.Location.Line {
				tooBig = r
			}
			continue
		}
		v = int64(int32(uint32(v))) // 0x80000000..0xFFFFFFFF is a 32-bit pattern
		r.value = v
		switch {
		case v >= -2048 && v <= 2047:
			r.form = liAddi
		case v&0xFFF == 0:
			r.form = liLui
		default:
			r.form, changed = liLong, true
		}
		r.setShort()
	}
	return changed, tooBig
}

// setShort writes the instruction of a short form with the value
func (r *liRecord) setShort() {
	rd, loc := r.lui.Operands[0], r.lui.Location
	addr := r.lui.Address
	if r.short != nil {
		addr = r.short.Address
	}
	switch r.form {
	case liAddi:
		r.short = &ast.Instruction{Opcode: "addi", Location: loc, Pseudo: "li", Address: addr,
			Operands: []ast.Expression{rd, &ast.Identifier{Name: "zero", Location: loc}, &ast.Number{Value: r.value, Location: loc}}}
	case liLui:
		r.short = &ast.Instruction{Opcode: "lui", Location: loc, Pseudo: "li", Address: addr,
			Operands: []ast.Expression{rd, &ast.Number{Value: (r.value >> 12) & 0xFFFFF, Location: loc}}}
	default:
		r.short = nil
	}
}

// rebuildStatements writes the program with each li in its form
func rebuildStatements(base []ast.Statement, records map[int]*liRecord) []ast.Statement {
	out := make([]ast.Statement, 0, len(base))
	for i := 0; i < len(base); i++ {
		if r, ok := records[i]; ok && r.form != liLong {
			out = append(out, r.short)
			i++ // skip the addi
			continue
		}
		out = append(out, base[i])
	}
	return out
}

// usesCurrentPC reports whether $ appears in expr
func usesCurrentPC(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.CurrentPC:
		return true
	case *ast.BinaryOp:
		return usesCurrentPC(e.Left) || usesCurrentPC(e.Right)
	case *ast.UnaryOp:
		return usesCurrentPC(e.Operand)
	case *ast.HiRel:
		return usesCurrentPC(e.Expr)
	case *ast.LoRel:
		return usesCurrentPC(e.Expr)
	}
	return false
}

// newSectionManager returns a section manager with the NATOGA32 memory map
func newSectionManager() *section.Manager {
	sections := section.NewManager()
	sections.SetBaseAddresses(
		MMAP_TEXT_SEGMENT,
		MMAP_DATA_SEGMENT,
		MMAP_BSS_SEGMENT,
		MMAP_ROM_SEGMENT,
		MMAP_RODATA_SEGMENT,
	)
	return sections
}
