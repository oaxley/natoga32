package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

// handleDirective processes assembler directives
func handleDirective(dir *ast.Directive, symbols *symbol.Table, sections *section.Manager) error {
	switch dir.Name {
	case ".text":
		return sections.SetSection(section.SectionText.String())
	case ".data":
		return sections.SetSection(section.SectionData.String())
	case ".bss":
		return sections.SetSection(section.SectionBss.String())
	case ".rodata":
		return sections.SetSection(section.SectionRodata.String())
	case ".rom":
		return sections.SetSection(section.SectionRom.String())

	case ".byte":
		sections.Current().Size += int64(len(dir.Args))

	case ".half", ".short":
		sections.Current().Size += int64(len(dir.Args) * 2)

	case ".word":
		sections.Current().Size += int64(len(dir.Args) * 4)

	case ".align":
		alignment, err := alignOperand(dir, symbols, sections)
		if err != nil {
			return err
		}
		sections.Align(alignment)

	case ".space", ".skip":
		size, err := spaceOperand(dir, symbols, sections)
		if err != nil {
			return err
		}
		sections.Current().Size += size

	case ".equ", ".set":
		return defineEqu(dir, symbols, sections)

	case ".global", ".globl":
		for _, arg := range dir.Args {
			if ident, ok := arg.(*ast.Identifier); ok {
				symbols.SetGlobal(ident.Name)
			}
		}

	case ".string", ".asciz":
		for _, arg := range dir.Args {
			if str, ok := arg.(*ast.StringLiteral); ok {
				sections.Current().Size += int64(len(str.Text) + 1) // +1 for null terminator
			}
		}

	case ".ascii":
		for _, arg := range dir.Args {
			if str, ok := arg.(*ast.StringLiteral); ok {
				sections.Current().Size += int64(len(str.Text))
			}
		}

	case ".incbin":
		if len(dir.Args) > 0 {
			if str, ok := dir.Args[0].(*ast.StringLiteral); ok {
				filename := filepath.Join(baseDir, str.Text)
				info, err := os.Stat(filename)
				if err != nil {
					return fmt.Errorf("%s: error: cannot stat incbin file '%s': %v", dir.Location, str.Text, err)
				}
				sections.Current().Size += info.Size()
			}
		}

	case ".org":
		gap, err := orgGap(dir, symbols, sections)
		if err != nil {
			return err
		}
		sections.Current().Size += gap

	default:
		// never silently ignore a directive: a typo would put data in the wrong section
		return fmt.Errorf("%s: error: unknown directive '%s'", dir.Location, dir.Name)
	}

	return nil
}

// emitDirective emits data for a directive
func emitDirective(dir *ast.Directive, symbols *symbol.Table, sections *section.Manager) error {
	switch dir.Name {
	case ".text":
		return sections.SetSection(section.SectionText.String())
	case ".data":
		return sections.SetSection(section.SectionData.String())
	case ".bss":
		return sections.SetSection(section.SectionBss.String())
	case ".rodata":
		return sections.SetSection(section.SectionRodata.String())
	case ".rom":
		return sections.SetSection(section.SectionRom.String())

	case ".byte":
		for _, arg := range dir.Args {
			val, err := evalExpressionWithError(arg, symbols, sections.CurrentAddress())
			if err != nil {
				return err
			}
			sections.EmitByte(byte(val))
		}

	case ".half", ".short":
		for _, arg := range dir.Args {
			val, err := evalExpressionWithError(arg, symbols, sections.CurrentAddress())
			if err != nil {
				return err
			}
			sections.EmitHalf(uint16(val))
		}

	case ".word":
		for _, arg := range dir.Args {
			val, err := evalExpressionWithError(arg, symbols, sections.CurrentAddress())
			if err != nil {
				return err
			}
			sections.EmitWord(uint32(val))
		}

	case ".align":
		alignment, err := alignOperand(dir, symbols, sections)
		if err != nil {
			return err
		}
		sections.Align(alignment)

	case ".space", ".skip":
		size, err := spaceOperand(dir, symbols, sections)
		if err != nil {
			return err
		}
		for i := int64(0); i < size; i++ {
			sections.EmitByte(0)
		}

	case ".string", ".asciz":
		for _, arg := range dir.Args {
			if str, ok := arg.(*ast.StringLiteral); ok {
				for _, ch := range str.Text {
					sections.EmitByte(byte(ch))
				}
				sections.EmitByte(0) // Null terminator
			}
		}

	case ".ascii":
		for _, arg := range dir.Args {
			if str, ok := arg.(*ast.StringLiteral); ok {
				for _, ch := range str.Text {
					sections.EmitByte(byte(ch))
				}
			}
		}

	case ".incbin":
		if len(dir.Args) > 0 {
			if str, ok := dir.Args[0].(*ast.StringLiteral); ok {
				filename := filepath.Join(baseDir, str.Text)
				data, err := os.ReadFile(filename)
				if err != nil {
					return fmt.Errorf("%s: error: cannot read incbin file '%s': %v", dir.Location, str.Text, err)
				}
				sections.Emit(data)
			}
		}

	case ".org":
		gap, err := orgGap(dir, symbols, sections)
		if err != nil {
			return err
		}
		for i := int64(0); i < gap; i++ {
			sections.EmitByte(0)
		}
	}

	return nil
}

// The operands of .space, .align, .org and .equ change addresses or define a
// name, so they are computed in pass 1: they can only use names defined above
// them. Both passes call the same helpers, so they always agree.

// singleOperand evaluates the only operand of a directive; $ is the current address
func singleOperand(dir *ast.Directive, symbols *symbol.Table, sections *section.Manager) (int64, error) {
	if len(dir.Args) != 1 {
		return 0, fmt.Errorf("%s: error: %s expects 1 operand, got %d", dir.Location, dir.Name, len(dir.Args))
	}
	return evalExpressionWithError(dir.Args[0], symbols, sections.CurrentAddress())
}

// spaceOperand returns the number of zero bytes of .space N
func spaceOperand(dir *ast.Directive, symbols *symbol.Table, sections *section.Manager) (int64, error) {
	size, err := singleOperand(dir, symbols, sections)
	if err != nil {
		return 0, err
	}
	if size < 0 {
		return 0, fmt.Errorf("%s: error: %s size %d is negative", dir.Location, dir.Name, size)
	}
	return size, nil
}

// alignOperand returns the alignment in bytes of .align N (2^N)
func alignOperand(dir *ast.Directive, symbols *symbol.Table, sections *section.Manager) (int, error) {
	n, err := singleOperand(dir, symbols, sections)
	if err != nil {
		return 0, err
	}
	if n < 0 || n > 31 {
		return 0, fmt.Errorf("%s: error: .align %d out of range (0..31)", dir.Location, n)
	}
	return 1 << n, nil
}

// orgGap returns the number of zero bytes up to the absolute address of .org ADDR
func orgGap(dir *ast.Directive, symbols *symbol.Table, sections *section.Manager) (int64, error) {
	targetAddr, err := singleOperand(dir, symbols, sections)
	if err != nil {
		return 0, err
	}
	currentAddr := sections.CurrentAddress()
	if targetAddr < currentAddr {
		return 0, fmt.Errorf("%s: error: .org address 0x%X is before current address 0x%X", dir.Location, targetAddr, currentAddr)
	}
	return targetAddr - currentAddr, nil
}

// defineEqu defines the constant of .equ NAME, VALUE (pass 1 only)
func defineEqu(dir *ast.Directive, symbols *symbol.Table, sections *section.Manager) error {
	if len(dir.Args) != 2 {
		return fmt.Errorf("%s: error: %s expects 2 operands (name, value), got %d", dir.Location, dir.Name, len(dir.Args))
	}
	ident, ok := dir.Args[0].(*ast.Identifier)
	if !ok {
		return fmt.Errorf("%s: error: %s expects a name as first operand", dir.Location, dir.Name)
	}
	value, err := evalExpressionWithError(dir.Args[1], symbols, sections.CurrentAddress())
	if err != nil {
		return err
	}
	if err := symbols.DefineConstant(ident.Name, value, dir.Location.Line); err != nil {
		return fmt.Errorf("%s: error: %v", dir.Location, err)
	}
	return nil
}
