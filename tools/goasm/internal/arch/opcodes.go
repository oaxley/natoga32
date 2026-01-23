// RISC-V opcode definitions for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package arch

// Opcodes maps instruction mnemonics to their encoding information
var Opcodes = map[string]InstrOpcode{
	// ========================================
	// Opcode 0x03 (3) - Load instructions
	// ========================================
	"lb":  {TypeI, 0b000_0011, 0, 0b000, 0, 0, 0},
	"lh":  {TypeI, 0b000_0011, 0, 0b001, 0, 0, 0},
	"lw":  {TypeI, 0b000_0011, 0, 0b010, 0, 0, 0},
	"lbu": {TypeI, 0b000_0011, 0, 0b100, 0, 0, 0},
	"lhu": {TypeI, 0b000_0011, 0, 0b101, 0, 0, 0},

	// ========================================
	// Opcode 0x0B (11) - Custom thread instructions
	// ========================================
	"new.t":   {TypeR, 0b000_1011, 0, 0b000, 0, 0, 0},
	"yield.t": {TypeI, 0b000_1011, 0, 0b001, 0, 0, 0},
	"id.t":    {TypeI, 0b000_1011, 0, 0b010, 0, 0, 0},
	"sleep.t": {TypeR2, 0b000_1011, 0, 0b100, 0, 0, 0},
	"wake.t":  {TypeR2, 0b000_1011, 0, 0b101, 0, 0, 0},
	"end.t":   {TypeI, 0b000_1011, 0, 0b111, 0, 0, 0},

	// ========================================
	// Opcode 0x13 (19) - Integer register-immediate
	// ========================================
	"addi":   {TypeI, 0b001_0011, 0, 0b000, 0, 0, 0},
	"slli":   {TypeI2, 0b001_0011, 0, 0b001, 0, 0, 0b000_0000},
	"bseti":  {TypeR, 0b001_0011, 0, 0b001, 0, 0, 0b001_0100},
	"bclri":  {TypeR, 0b001_0011, 0, 0b001, 0, 0, 0b010_0100},
	"sext.h": {TypeRU, 0b001_0011, 0, 0b001, 0, 0b00101, 0b011_0000},
	"clz":    {TypeRU, 0b001_0011, 0, 0b001, 0, 0b00000, 0b011_0000},
	"ctz":    {TypeRU, 0b001_0011, 0, 0b001, 0, 0b00001, 0b011_0000},
	"cpop":   {TypeRU, 0b001_0011, 0, 0b001, 0, 0b00010, 0b011_0000},
	"sext.b": {TypeRU, 0b001_0011, 0, 0b001, 0, 0b00100, 0b011_0000},
	"binvi":  {TypeR, 0b001_0011, 0, 0b001, 0, 0, 0b011_0000},
	"slti":   {TypeI, 0b001_0011, 0, 0b010, 0, 0, 0},
	"sltiu":  {TypeI, 0b001_0011, 0, 0b011, 0, 0, 0},
	"xori":   {TypeI, 0b001_0011, 0, 0b100, 0, 0, 0},
	"srli":   {TypeI2, 0b001_0011, 0, 0b101, 0, 0, 0b000_0000},
	"srai":   {TypeI2, 0b001_0011, 0, 0b101, 0, 0, 0b010_0000},
	"bexti":  {TypeR, 0b001_0011, 0, 0b101, 0, 0, 0b010_0100},
	"rori":   {TypeR, 0b001_0011, 0, 0b101, 0, 0, 0b011_0000},
	"rev8":   {TypeRU, 0b001_0011, 0, 0b101, 0, 0b11000, 0b011_0100},
	"ori":    {TypeI, 0b001_0011, 0, 0b110, 0, 0, 0},
	"andi":   {TypeI, 0b001_0011, 0, 0b111, 0, 0, 0},

	// ========================================
	// Opcode 0x17 (23) - AUIPC
	// ========================================
	"auipc": {TypeU, 0b001_0111, 0, 0, 0, 0, 0},

	// ========================================
	// Opcode 0x23 (35) - Store instructions
	// ========================================
	"sb": {TypeS, 0b010_0011, 0, 0b000, 0, 0, 0},
	"sh": {TypeS, 0b010_0011, 0, 0b001, 0, 0, 0},
	"sw": {TypeS, 0b010_0011, 0, 0b010, 0, 0, 0},

	// ========================================
	// Opcode 0x33 (51) - Integer register-register
	// ========================================
	"add":       {TypeR, 0b011_0011, 0, 0b000, 0, 0, 0b000_0000},
	"mul":       {TypeR, 0b011_0011, 0, 0b000, 0, 0, 0b000_0001},
	"sub":       {TypeR, 0b011_0011, 0, 0b000, 0, 0, 0b010_0000},
	"sll":       {TypeR, 0b011_0011, 0, 0b001, 0, 0, 0b000_0000},
	"mulh":      {TypeR, 0b011_0011, 0, 0b001, 0, 0, 0b000_0001},
	"bset":      {TypeR, 0b011_0011, 0, 0b001, 0, 0, 0b001_0100},
	"bclr":      {TypeR, 0b011_0011, 0, 0b001, 0, 0, 0b010_0100},
	"rol":       {TypeR, 0b011_0011, 0, 0b001, 0, 0, 0b011_0000},
	"binv":      {TypeR, 0b011_0011, 0, 0b001, 0, 0, 0b011_0100},
	"slt":       {TypeR, 0b011_0011, 0, 0b010, 0, 0, 0b000_0000},
	"mulhsu":    {TypeR, 0b011_0011, 0, 0b010, 0, 0, 0b000_0001},
	"sltu":      {TypeR, 0b011_0011, 0, 0b011, 0, 0, 0b000_0000},
	"mulhu":     {TypeR, 0b011_0011, 0, 0b011, 0, 0, 0b000_0001},
	"xor":       {TypeR, 0b011_0011, 0, 0b100, 0, 0, 0b000_0000},
	"div":       {TypeR, 0b011_0011, 0, 0b100, 0, 0, 0b000_0001},
	"zext.h":    {TypeR, 0b011_0011, 0, 0b100, 0, 0, 0b000_0100},
	"pack":      {TypeR, 0b011_0011, 0, 0b100, 0, 0, 0b000_0100},
	"min":       {TypeR, 0b011_0011, 0, 0b100, 0, 0, 0b000_0101},
	"srl":       {TypeR, 0b011_0011, 0, 0b101, 0, 0, 0b000_0000},
	"divu":      {TypeR, 0b011_0011, 0, 0b101, 0, 0, 0b000_0001},
	"minu":      {TypeR, 0b011_0011, 0, 0b101, 0, 0, 0b000_0101},
	"czero.eqz": {TypeR, 0b011_0011, 0, 0b101, 0, 0, 0b000_0111},
	"sra":       {TypeR, 0b011_0011, 0, 0b101, 0, 0, 0b010_0000},
	"bext":      {TypeR, 0b011_0011, 0, 0b101, 0, 0, 0b010_0100},
	"ror":       {TypeR, 0b011_0011, 0, 0b101, 0, 0, 0b011_0000},
	"or":        {TypeR, 0b011_0011, 0, 0b110, 0, 0, 0b000_0000},
	"rem":       {TypeR, 0b011_0011, 0, 0b110, 0, 0, 0b000_0001},
	"max":       {TypeR, 0b011_0011, 0, 0b110, 0, 0, 0b000_0101},
	"and":       {TypeR, 0b011_0011, 0, 0b111, 0, 0, 0b000_0000},
	"remu":      {TypeR, 0b011_0011, 0, 0b111, 0, 0, 0b000_0001},
	"packh":     {TypeR, 0b011_0011, 0, 0b111, 0, 0, 0b000_0100},
	"maxu":      {TypeR, 0b011_0011, 0, 0b111, 0, 0, 0b000_0101},
	"czero.nez": {TypeR, 0b011_0011, 0, 0b111, 0, 0, 0b000_0111},

	// ========================================
	// Opcode 0x37 (55) - LUI
	// ========================================
	"lui": {TypeU, 0b011_0111, 0, 0, 0, 0, 0},

	// ========================================
	// Opcode 0x63 (99) - Branch instructions
	// ========================================
	"beq":  {TypeB, 0b110_0011, 0, 0b000, 0, 0, 0},
	"bne":  {TypeB, 0b110_0011, 0, 0b001, 0, 0, 0},
	"blt":  {TypeB, 0b110_0011, 0, 0b100, 0, 0, 0},
	"bge":  {TypeB, 0b110_0011, 0, 0b101, 0, 0, 0},
	"bltu": {TypeB, 0b110_0011, 0, 0b110, 0, 0, 0},
	"bgeu": {TypeB, 0b110_0011, 0, 0b111, 0, 0, 0},

	// ========================================
	// Opcode 0x67 (103) - JALR
	// ========================================
	"jalr": {TypeI, 0b110_0111, 0, 0b000, 0, 0, 0},

	// ========================================
	// Opcode 0x6F (111) - JAL
	// ========================================
	"jal": {TypeJ, 0b110_1111, 0, 0, 0, 0, 0},

	// ========================================
	// Opcode 0x73 (115) - System instructions
	// ========================================
	"ecall":  {TypeI, 0b111_0011, 0, 0b000, 0, 0, 0b000_0000},
	"mret":   {TypeI, 0b111_0011, 0, 0b000, 0, 0b00010, 0b001_1000},
	"wfi":    {TypeI, 0b111_0011, 0, 0b000, 0, 0b00101, 0b000_1000},
	"csrrw":  {TypeI, 0b111_0011, 0, 0b001, 0, 0, 0},
	"csrrs":  {TypeI, 0b111_0011, 0, 0b010, 0, 0, 0},
	"csrrc":  {TypeI, 0b111_0011, 0, 0b011, 0, 0, 0},
	"csrrwi": {TypeI, 0b111_0011, 0, 0b101, 0, 0, 0},
	"csrrsi": {TypeI, 0b111_0011, 0, 0b110, 0, 0, 0},
	"csrrci": {TypeI, 0b111_0011, 0, 0b111, 0, 0, 0},
}

// LookupOpcode returns the opcode information for a given mnemonic.
// Returns (opcode, true) if found, (zero, false) if not a valid opcode.
func LookupOpcode(mnemonic string) (InstrOpcode, bool) {
	op, ok := Opcodes[mnemonic]
	return op, ok
}

// IsInstruction checks if the given mnemonic is a valid instruction
func IsInstruction(mnemonic string) bool {
	_, ok := LookupOpcode(mnemonic)
	return ok
}
