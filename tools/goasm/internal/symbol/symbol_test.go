package symbol

import "testing"

func TestDefineLabel(t *testing.T) {
	table := New()

	err := table.DefineLabel("main", 0x100, ".text", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sym := table.Lookup("main")
	if sym == nil {
		t.Fatal("symbol not found")
	}
	if sym.Name != "main" {
		t.Errorf("expected name 'main', got '%s'", sym.Name)
	}
	if sym.Type != SymLabel {
		t.Errorf("expected type LABEL, got %v", sym.Type)
	}
	if sym.Value != 0x100 {
		t.Errorf("expected value 0x100, got 0x%X", sym.Value)
	}
	if sym.Section != ".text" {
		t.Errorf("expected section '.text', got '%s'", sym.Section)
	}
}

func TestDefineConstant(t *testing.T) {
	table := New()

	err := table.DefineConstant("BUFFER_SIZE", 1024, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sym := table.Lookup("BUFFER_SIZE")
	if sym == nil {
		t.Fatal("symbol not found")
	}
	if sym.Type != SymConstant {
		t.Errorf("expected type CONST, got %v", sym.Type)
	}
	if sym.Value != 1024 {
		t.Errorf("expected value 1024, got %d", sym.Value)
	}
}

func TestDuplicateDefinition(t *testing.T) {
	table := New()

	err := table.DefineLabel("main", 0x100, ".text", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = table.DefineLabel("main", 0x200, ".text", 10)
	if err == nil {
		t.Fatal("expected error for duplicate definition")
	}
}

func TestReference(t *testing.T) {
	table := New()

	// Reference an undefined symbol
	table.Reference("undefined_func")

	sym := table.Lookup("undefined_func")
	if sym == nil {
		t.Fatal("referenced symbol not found")
	}
	if sym.Defined {
		t.Error("referenced symbol should not be defined")
	}

	// Later define it
	err := table.DefineLabel("undefined_func", 0x300, ".text", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sym = table.Lookup("undefined_func")
	if !sym.Defined {
		t.Error("symbol should now be defined")
	}
}

func TestGlobalSymbol(t *testing.T) {
	table := New()

	err := table.DefineLabel("public_func", 0x100, ".text", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = table.SetGlobal("public_func")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sym := table.Lookup("public_func")
	if !sym.Global {
		t.Error("symbol should be marked as global")
	}

	globals := table.GlobalSymbols()
	if len(globals) != 1 {
		t.Errorf("expected 1 global symbol, got %d", len(globals))
	}
}

func TestUpdateValue(t *testing.T) {
	table := New()

	err := table.DefineLabel("loop", 0, ".text", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = table.UpdateValue("loop", 0x400)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sym := table.Lookup("loop")
	if sym.Value != 0x400 {
		t.Errorf("expected value 0x400, got 0x%X", sym.Value)
	}
}

func TestUndefinedSymbols(t *testing.T) {
	table := New()

	table.Reference("func1")
	table.Reference("func2")
	table.DefineLabel("func1", 0x100, ".text", 1)

	undefined := table.UndefinedSymbols()
	if len(undefined) != 1 {
		t.Errorf("expected 1 undefined symbol, got %d", len(undefined))
	}
	if undefined[0].Name != "func2" {
		t.Errorf("expected undefined symbol 'func2', got '%s'", undefined[0].Name)
	}
}

func TestIsDefined(t *testing.T) {
	table := New()

	if table.IsDefined("nonexistent") {
		t.Error("nonexistent symbol should not be defined")
	}

	table.Reference("forward_ref")
	if table.IsDefined("forward_ref") {
		t.Error("referenced symbol should not be defined")
	}

	table.DefineLabel("defined", 0x100, ".text", 1)
	if !table.IsDefined("defined") {
		t.Error("defined symbol should be defined")
	}
}

func TestSize(t *testing.T) {
	table := New()

	if table.Size() != 0 {
		t.Errorf("expected size 0, got %d", table.Size())
	}

	table.DefineLabel("a", 0, ".text", 1)
	table.DefineLabel("b", 4, ".text", 2)
	table.DefineConstant("c", 100, 3)

	if table.Size() != 3 {
		t.Errorf("expected size 3, got %d", table.Size())
	}
}
