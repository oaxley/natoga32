# goasm - NATOGA32 RISC-V Assembler

A Go-based assembler for the NATOGA32 virtual console, targeting RISC-V RV32I with M extension and custom threading instructions.

## Building

```bash
cd tools/goasm
go build -o goasm ./cmd/goasm/
```

## Usage

```bash
# Basic assembly to ON32 format (default)
./goasm input.asm -o output.on32

# Assembly to raw binary
./goasm input.asm -o output.bin -f bin

# Debug modes (stop after specific phase)
./goasm input.asm -d 1  # Show tokens
./goasm input.asm -d 2  # Show preprocessed tokens
./goasm input.asm -d 3  # Show AST
./goasm input.asm -d 4  # Show expanded AST (after pseudo-expansion)
./goasm input.asm -d 5  # Show symbols and sections
./goasm input.asm -d 7  # Show encoded hex dump
```

### Command Line Options

| Option | Description |
|--------|-------------|
| `-o, --output` | Output file path (default: `a.out`) |
| `-f, --format` | Output format: `on32` or `bin` (default: `on32`) |
| `-d, --debug` | Debug level 1-8 to stop after specific phase |

## Assembly Syntax

### Basic Structure

```asm
; Comments start with semicolon or hash
# This is also a comment
// C++ style comments work too

.text                   ; Code section
main:                   ; Labels end with colon
    addi x1, x0, 10     ; Instructions

.data                   ; Data section
message:
    .string "Hello"     ; Directives start with dot
```

### Registers

Standard RISC-V register names and ABI aliases are supported:

| Register | ABI Name | Description |
|----------|----------|-------------|
| x0 | zero | Hardwired zero |
| x1 | ra | Return address |
| x2 | sp | Stack pointer |
| x3 | gp | Global pointer |
| x4 | tp | Thread pointer |
| x5-x7 | t0-t2 | Temporary registers |
| x8 | s0/fp | Saved register / Frame pointer |
| x9 | s1 | Saved register |
| x10-x11 | a0-a1 | Function arguments / Return values |
| x12-x17 | a2-a7 | Function arguments |
| x18-x27 | s2-s11 | Saved registers |
| x28-x31 | t3-t6 | Temporary registers |

### Number Formats

```asm
.word 42          ; Decimal
.word 0x2A        ; Hexadecimal
.word 0b101010    ; Binary
.word 0o52        ; Octal
.word 'A'         ; Character literal (ASCII value)
```

## Preprocessor Directives

### Constants and Symbols

```asm
.define BUFFER_SIZE 1024        ; Define a constant
.equ    MAX_VALUE, 255          ; Alternative syntax

.global main                    ; Export symbol globally
```

### File Inclusion

```asm
.include "header.inc"           ; Include another file
.incbin "data.bin"              ; Include binary file as raw bytes
```

### Conditional Assembly

```asm
.define DEBUG 1

.ifdef DEBUG
    ; Code compiled only if DEBUG is defined
.endif

.ifndef RELEASE
    ; Code compiled only if RELEASE is NOT defined
.endif

.if BUFFER_SIZE > 512
    ; Code compiled if expression is non-zero
.else
    ; Alternative code
.endif
```

### Loops

```asm
; Generate a series of values
.for i = 0, 10
    .word i                     ; Generates .word 0, .word 1, ... .word 9
.endf

; With custom step
.for j = 10, 0, -2
    .byte j                     ; Generates 10, 8, 6, 4, 2
.endf
```

### Macros

```asm
; Define a macro with parameters
.macro push reg
    addi sp, sp, -4
    sw \reg, 0(sp)
.endmacro

; Use the macro
push ra
push s0

; Multiple parameters
.macro add3 dst, a, b
    add \dst, \a, \b
.endmacro

add3 t0, a0, a1
```

### Token Pasting

```asm
.define COUNT 5

; Generate register names dynamically
.for i = 0, 4
    addi x##i, zero, i          ; Produces: addi x0, zero, 0 ... addi x3, zero, 3
.endf

; Generate labels
handler##COUNT:                  ; Produces: handler5:
```

### Value Duplication

```asm
.byte .dup 10 0                 ; 10 zero bytes
.word .dup 4 0xDEADBEEF         ; 4 copies of 0xDEADBEEF
.byte .dup (SIZE * 2) 0xFF      ; Expression for count
```

### Environment Variables

```asm
; Reference shell environment variables
.define ROM_BASE $[ROM_BASE]    ; Uses $ROM_BASE from environment
.define DEBUG $[DEBUG_LEVEL]    ; Numeric values become numbers

.string $[BUILD_VERSION]        ; Non-numeric values become strings
```

## Data Directives

| Directive | Description |
|-----------|-------------|
| `.text` | Switch to code section |
| `.data` | Switch to initialized data section |
| `.rodata` | Switch to read-only data section |
| `.bss` | Switch to uninitialized data section |
| `.byte val [, val...]` | Emit 8-bit values |
| `.half val [, val...]` | Emit 16-bit values |
| `.word val [, val...]` | Emit 32-bit values |
| `.string "text"` | Emit null-terminated string |
| `.ascii "text"` | Emit string without null terminator |
| `.space N` | Reserve N bytes (zero-filled) |
| `.align N` | Align to 2^N byte boundary |

## Relocation Modifiers

For accessing symbols that don't fit in immediate fields:

```asm
; Load full 32-bit address
lui  a0, %hi(symbol)
addi a0, a0, %lo(symbol)

; PC-relative addressing
auipc a0, %pcrel_hi(symbol)
addi  a0, a0, %pcrel_lo(symbol)
```

| Modifier | Description |
|----------|-------------|
| `%hi(sym)` | Upper 20 bits of absolute address |
| `%lo(sym)` | Lower 12 bits of absolute address |
| `%pcrel_hi(sym)` | Upper 20 bits of PC-relative offset |
| `%pcrel_lo(sym)` | Lower 12 bits of PC-relative offset |

## Supported Instructions

### RV32I Base Integer Instructions

**Arithmetic:**
- `add`, `sub`, `addi`
- `lui`, `auipc`

**Logical:**
- `and`, `or`, `xor`
- `andi`, `ori`, `xori`

**Shifts:**
- `sll`, `srl`, `sra`
- `slli`, `srli`, `srai`

**Compare:**
- `slt`, `sltu`
- `slti`, `sltiu`

**Branch:**
- `beq`, `bne`
- `blt`, `bge`, `bltu`, `bgeu`

**Jump:**
- `jal`, `jalr`

**Load:**
- `lb`, `lh`, `lw`
- `lbu`, `lhu`

**Store:**
- `sb`, `sh`, `sw`

### RV32M Multiply/Divide Extension

- `mul`, `mulh`, `mulhsu`, `mulhu`
- `div`, `divu`, `rem`, `remu`

### Zbb Bit Manipulation (partial)

- `clz`, `ctz`, `cpop`
- `min`, `max`, `minu`, `maxu`
- `sext.b`, `sext.h`, `zext.h`
- `rol`, `ror`, `rori`
- `rev8`

### System Instructions

- `ecall` - Environment call
- `mret` - Machine-mode return
- `wfi` - Wait for interrupt

**CSR Access:**
- `csrrw`, `csrrs`, `csrrc`
- `csrrwi`, `csrrsi`, `csrrci`

### Custom Thread Instructions (NATOGA32)

| Instruction | Description |
|-------------|-------------|
| `new.t rd, rs1` | Create new thread, store ID in rd |
| `yield.t` | Yield current thread |
| `id.t rd` | Get current thread ID |
| `sleep.t rd, rs1` | Put thread to sleep |
| `wake.t rd, rs1` | Wake sleeping thread |
| `end.t` | Terminate current thread |

## Pseudo-Instructions

The assembler automatically expands these into real instructions:

| Pseudo | Expansion | Description |
|--------|-----------|-------------|
| `nop` | `addi x0, x0, 0` | No operation |
| `li rd, imm` | `lui` + `addi` | Load immediate |
| `la rd, sym` | `auipc` + `addi` | Load address |
| `mv rd, rs` | `addi rd, rs, 0` | Move register |
| `not rd, rs` | `xori rd, rs, -1` | Bitwise NOT |
| `neg rd, rs` | `sub rd, x0, rs` | Negate |
| `j label` | `jal x0, label` | Unconditional jump |
| `jr rs` | `jalr x0, rs, 0` | Jump register |
| `ret` | `jalr x0, ra, 0` | Return from function |
| `call label` | `auipc` + `jalr` | Call function |

**Branch Aliases:**
| Pseudo | Expansion |
|--------|-----------|
| `beqz rs, label` | `beq rs, x0, label` |
| `bnez rs, label` | `bne rs, x0, label` |
| `blez rs, label` | `bge x0, rs, label` |
| `bgez rs, label` | `bge rs, x0, label` |
| `bltz rs, label` | `blt rs, x0, label` |
| `bgtz rs, label` | `blt x0, rs, label` |
| `bgt rs, rt, label` | `blt rt, rs, label` |
| `ble rs, rt, label` | `bge rt, rs, label` |
| `bgtu rs, rt, label` | `bltu rt, rs, label` |
| `bleu rs, rt, label` | `bgeu rt, rs, label` |

**Set Conditionals:**
| Pseudo | Expansion |
|--------|-----------|
| `seqz rd, rs` | `sltiu rd, rs, 1` |
| `snez rd, rs` | `sltu rd, x0, rs` |
| `sltz rd, rs` | `slt rd, rs, x0` |
| `sgtz rd, rs` | `slt rd, x0, rs` |

**CSR Pseudo-Instructions:**
| Pseudo | Expansion |
|--------|-----------|
| `csrr rd, csr` | `csrrs rd, csr, x0` |
| `csrw csr, rs` | `csrrw x0, csr, rs` |
| `csrs csr, rs` | `csrrs x0, csr, rs` |
| `csrc csr, rs` | `csrrc x0, csr, rs` |
| `csrwi csr, imm` | `csrrwi x0, csr, imm` |
| `csrsi csr, imm` | `csrrsi x0, csr, imm` |
| `csrci csr, imm` | `csrrci x0, csr, imm` |

## Example Program

```asm
; Simple program that sums numbers 1 to 10

.text
.global _start

_start:
    li   a0, 0              ; sum = 0
    li   a1, 1              ; i = 1
    li   a2, 10             ; limit = 10

loop:
    add  a0, a0, a1         ; sum += i
    addi a1, a1, 1          ; i++
    ble  a1, a2, loop       ; if i <= 10, continue

    ; Result in a0
    ecall                   ; Exit

.data
message:
    .string "Sum: "
```

## Output Formats

### Binary (`-f bin`)
Raw machine code output, suitable for direct ROM loading.

### ON32 (`-f on32`)
NATOGA32 object format with header containing:
- Magic number and version
- Entry point address
- Section information

## Testing

```bash
cd tools/goasm
go test ./...
```

## License

This project is subject to the MIT License. See LICENSE.txt for details.
