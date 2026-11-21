package main

import (
	"fmt"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

// li with a name or an expression is expanded to lui %hi + addi %lo, since
// the value is not known before pass 1. layout then makes it one
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

// layout runs pass 1 until the addresses and the names no longer change,
// and returns its symbols and sections.
//
// .space, .align, .org and .equ can use names defined below them. In the
// first pass such a name is 0, in the next ones it has its value of the
// previous pass, so a pass may be wrong and its errors are kept only when
// the layout is stable: then every name has the value it was computed
// with. Each li with a name also takes its shortest form (see above).
func layout(program *ast.Program, newSections func() *section.Manager) (*symbol.Table, *section.Manager, error) {
	base := program.Statements
	records := findLiPairs(base)
	program.Statements = rebuildStatements(base, records)
	maxPasses := layoutPassLimit(base, records)
	var prev *symbol.Table
	for pass := 1; ; pass++ {
		symbols, sections := symbol.New(), newSections()
		symbols.StartLayout(prev)
		err := semanticPass1(program, symbols, sections)
		changed, tooBig := chooseLiForms(records, symbols)
		// rebuild every time: with the same forms the layout is the same,
		// but the short instructions get the values of this pass
		program.Statements = rebuildStatements(base, records)
		if !changed && symbol.SameValues(symbols, prev) {
			symbols.EndLayout()
			if err != nil {
				return nil, nil, fmt.Errorf("semantic pass 1 error: %v", err)
			}
			if tooBig != nil {
				return nil, nil, fmt.Errorf("semantic pass 1 error: %s: error: li value %d (0x%X) does not fit 32 bits",
					tooBig.lui.Location, tooBig.value, tooBig.value)
			}
			return symbols, sections, nil
		}
		if pass >= maxPasses {
			return nil, nil, unsettledError(symbols, prev, err)
		}
		prev = symbols
	}
}

// layoutPassLimit bounds the passes of layout. A chain of n names, each
// defined with the next one, needs n+1 passes, and each li can grow once.
// Labels move only when one of these changes.
func layoutPassLimit(stmts []ast.Statement, records map[int]*liRecord) int {
	n := 4 + len(records)
	for _, stmt := range stmts {
		if s, ok := stmt.(*ast.Directive); ok {
			switch s.Name {
			case ".equ", ".set", ".space", ".skip", ".align", ".org":
				n++
			}
		}
	}
	return n
}

// unsettledError reports the layout that does not settle: a name that
// depends on itself (.equ A, A + 1), directly or not. It names the first one, by
// line, whose value still changes.
func unsettledError(last, prev *symbol.Table, err error) error {
	var worst *symbol.Symbol
	for _, sym := range last.DefinedSymbols() {
		p := prev.Lookup(sym.Name)
		if p != nil && p.Defined && p.Value == sym.Value {
			continue
		}
		if worst == nil || sym.File < worst.File || (sym.File == worst.File && sym.Line < worst.Line) {
			worst = sym
		}
	}
	if worst == nil {
		// only a li keeps changing, or a pass error: report what pass 1 said
		if err != nil {
			return fmt.Errorf("semantic pass 1 error: %v", err)
		}
		return fmt.Errorf("semantic pass 1 error: the layout does not settle")
	}
	where := fmt.Sprintf("line %d", worst.Line)
	if worst.File != "" {
		where = fmt.Sprintf("%s:%d", worst.File, worst.Line)
	}
	return fmt.Errorf("semantic pass 1 error: %s: error: the value of '%s' does not settle (it depends on itself, directly or through other names)",
		where, worst.Name)
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
