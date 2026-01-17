package main

import (
	"testing"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/section"
	"github.com/natoga32/goasm/internal/symbol"
)

func TestOrgDirectivePass1(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000)

	// Start in .text at 0x1000
	// Emit 4 bytes first
	sections.Current().Size += 4 // Simulate 4 bytes emitted

	// .org 0x1010 should add 12 bytes of padding (0x1010 - 0x1004 = 12)
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1010}},
	}

	err := handleDirective(dir, symbols, sections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSize := int64(4 + 12) // original 4 + 12 padding
	if sections.Current().Size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, sections.Current().Size)
	}
}

func TestOrgDirectivePass1AtCurrentAddress(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000)

	// .org at current address should be a no-op
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1000}},
	}

	err := handleDirective(dir, symbols, sections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sections.Current().Size != 0 {
		t.Errorf("expected size 0, got %d", sections.Current().Size)
	}
}

func TestOrgDirectivePass1BackwardError(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000)

	// Emit 16 bytes first
	sections.Current().Size += 16 // Current address is now 0x1010

	// .org 0x1008 should fail (trying to go backward)
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1008}},
	}

	err := handleDirective(dir, symbols, sections)
	if err == nil {
		t.Error("expected error for backward .org, got nil")
	}
}

func TestOrgDirectivePass3(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000)

	// Emit 4 bytes first
	sections.EmitWord(0xDEADBEEF)

	// .org 0x1010 should emit 12 zero bytes
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1010}},
	}

	err := emitDirective(dir, symbols, sections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSize := int64(4 + 12)
	if sections.Current().Size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, sections.Current().Size)
	}

	// Verify the padding bytes are all zeros
	data := sections.Current().Data
	for i := 4; i < 16; i++ {
		if data[i] != 0 {
			t.Errorf("expected zero at offset %d, got 0x%02X", i, data[i])
		}
	}

	// Verify current address
	if sections.CurrentAddress() != 0x1010 {
		t.Errorf("expected current address 0x1010, got 0x%X", sections.CurrentAddress())
	}
}

func TestOrgDirectivePass3BackwardError(t *testing.T) {
	symbols := symbol.New()
	sections := section.NewManager()
	sections.SetBaseAddresses(0x1000, 0x2000, 0x3000, 0x4000)

	// Emit 16 bytes first
	for i := 0; i < 4; i++ {
		sections.EmitWord(0x11111111)
	}

	// .org 0x1008 should fail (trying to go backward)
	dir := &ast.Directive{
		Name: ".org",
		Args: []ast.Expression{&ast.Number{Value: 0x1008}},
	}

	err := emitDirective(dir, symbols, sections)
	if err == nil {
		t.Error("expected error for backward .org, got nil")
	}
}
