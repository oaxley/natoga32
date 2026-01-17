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

	case ".byte":
		sections.Current().Size += int64(len(dir.Args))

	case ".half", ".short":
		sections.Current().Size += int64(len(dir.Args) * 2)

	case ".word":
		sections.Current().Size += int64(len(dir.Args) * 4)

	case ".align":
		if len(dir.Args) > 0 {
			if num, ok := dir.Args[0].(*ast.Number); ok {
				alignment := 1 << num.Value // .align N means align to 2^N
				sections.Align(alignment)
			}
		}

	case ".space", ".skip":
		if len(dir.Args) > 0 {
			if num, ok := dir.Args[0].(*ast.Number); ok {
				sections.Current().Size += num.Value
			}
		}

	case ".equ", ".set":
		if len(dir.Args) >= 2 {
			if ident, ok := dir.Args[0].(*ast.Identifier); ok {
				if num, ok := dir.Args[1].(*ast.Number); ok {
					return symbols.DefineConstant(ident.Name, num.Value, 0)
				}
			}
		}

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
					return fmt.Errorf("cannot stat incbin file '%s': %v", str.Text, err)
				}
				sections.Current().Size += info.Size()
			}
		}

	case ".org":
		if len(dir.Args) > 0 {
			if num, ok := dir.Args[0].(*ast.Number); ok {
				targetAddr := num.Value
				currentAddr := sections.CurrentAddress()
				gap := targetAddr - currentAddr
				if gap < 0 {
					return fmt.Errorf(".org address 0x%X is before current address 0x%X", targetAddr, currentAddr)
				}
				sections.Current().Size += gap
			}
		}
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

	case ".byte":
		for _, arg := range dir.Args {
			val := evalExpression(arg, symbols, sections.CurrentAddress())
			sections.EmitByte(byte(val))
		}

	case ".half", ".short":
		for _, arg := range dir.Args {
			val := evalExpression(arg, symbols, sections.CurrentAddress())
			sections.EmitHalf(uint16(val))
		}

	case ".word":
		for _, arg := range dir.Args {
			val := evalExpression(arg, symbols, sections.CurrentAddress())
			sections.EmitWord(uint32(val))
		}

	case ".align":
		if len(dir.Args) > 0 {
			if num, ok := dir.Args[0].(*ast.Number); ok {
				alignment := 1 << num.Value
				sections.Align(alignment)
			}
		}

	case ".space", ".skip":
		if len(dir.Args) > 0 {
			if num, ok := dir.Args[0].(*ast.Number); ok {
				for i := int64(0); i < num.Value; i++ {
					sections.EmitByte(0)
				}
			}
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
					return fmt.Errorf("cannot read incbin file '%s': %v", str.Text, err)
				}
				sections.Emit(data)
			}
		}

	case ".org":
		if len(dir.Args) > 0 {
			if num, ok := dir.Args[0].(*ast.Number); ok {
				targetAddr := num.Value
				currentAddr := sections.CurrentAddress()
				gap := targetAddr - currentAddr
				if gap < 0 {
					return fmt.Errorf(".org address 0x%X is before current address 0x%X", targetAddr, currentAddr)
				}
				for i := int64(0); i < gap; i++ {
					sections.EmitByte(0)
				}
			}
		}
	}

	return nil
}
