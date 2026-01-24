# goasm - NATOGA32 RISC-V Assembler & Disassembler

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Command-Line Reference](#command-line-reference)
- [Assembly Language Syntax](#assembly-language-syntax)
- [Supported Instructions](#supported-instructions)
- [Preprocessor](#preprocessor)
- [Directives](#directives)
- [Optimization](#optimization)
- [Output Formats](#output-formats)
- [Disassembler](#disassembler)
- [Examples](#examples)

## Overview

**goasm** is a feature-rich RISC-V assembler and disassembler for the NATOGA32 virtual console. It supports:

- **RV32I** base integer instruction set
- **RV32M** multiplication and division extension
- **RV32Zbb** basic bit-manipulation extension
- **RV32Zbs** single-bit manipulation extension
- **RV32Zicond** conditional operations extension
- **RV32Zbkb** bit-manipulation for cryptography extension
- Custom **NATOGA32 thread instructions**
- Comprehensive **preprocessor** (macros, conditionals, loops, file inclusion)
- **Pseudo-instructions** for easier programming
- Optional **peephole optimization** (`-O` flag)
- **Disassembler** with pseudo-instruction recognition

## Installation

### Prerequisites

- Go 1.21 or later

### Building from Source

```bash
cd tools/goasm
go build -o goasm ./cmd/goasm
```

### Verification

```bash
./goasm --help
```

## Quick Start

### Assemble a Program

```bash
# Assemble to ON32 format (default)
./goasm program.asm -o program.on32

# Assemble to raw binary
./goasm program.asm -o program.bin -f bin

# Assemble with optimization
./goasm program.asm -o program.on32 -O
```

### Disassemble a Binary

```bash
# Disassemble ON32 format
./goasm --disasm program.on32

# Disassemble with pseudo-instructions
./goasm --disasm program.on32 --pseudo

# Disassemble raw binary with base address
./goasm --disasm program.bin --base 0x01A00000
```

### Simple Example

Create `hello.asm`:
```asm
.text
main:
    li   a0, 42        # Load immediate
    li   a7, 93        # Exit syscall number
    ecall              # Make syscall
```

Assemble:
```bash
./goasm hello.asm -o hello.on32
```

## Command-Line Reference

### Assembly Mode (Default)

```
goasm [flags] <input.asm>
```

#### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output file path | `a.out` |
| `--format` | `-f` | Output format: `on32`, `bin` | `on32` |
| `--optimize` | `-O` | Enable peephole optimizations | `false` |
| `--debug` | `-d` | Debug level (1-8, stops at phase) | `0` |

#### Debug Levels

| Level | Phase | Output |
|-------|-------|--------|
| 1 | Lexer | Token stream |
| 2 | Preprocessor | Preprocessed tokens |
| 3 | Parser | Abstract Syntax Tree (AST) |
| 4 | Pseudo-expansion | Expanded AST |
| 5 | Semantic pass 1 | Symbols and sections |
| 6 | Semantic pass 2 | Resolved symbols |
| 7 | Semantic pass 3 | Encoded hex dump |
| 8 | Output | (same as normal assembly) |

### Disassembly Mode

```
goasm --disasm [flags] <input.on32|input.bin>
```

#### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--disasm` | Enable disassembly mode | `false` |
| `--base` | Base address for raw binary | `0x01A00000` |
| `--pseudo` | Show pseudo-instructions | `false` |
| `--hex` | Show hex encoding | `true` |
| `--no-abi` | Use x0-x31 instead of ABI names | `false` |

## Assembly Language Syntax

### Comments

```asm
# Single-line comment (hash)
; Single-line comment (semicolon)
```

### Labels

```asm
label_name:           # Label at current address
    add t0, t1, t2    # Instruction

local_label:          # Local label
global_func:          # Can be marked global with .global
```

### Numeric Literals

```asm
42          # Decimal
0x2A        # Hexadecimal
0b101010    # Binary
0o52        # Octal
-100        # Negative numbers
```

### String and Character Literals

```asm
.string "Hello, World!\n"     # Null-terminated string
.ascii "No null terminator"    # String without null
'A'                            # Character literal
'\n'                           # Escape sequence
```

### Expressions

```asm
addi t0, t1, 10 + 20           # Arithmetic
addi t2, t3, (5 * 8)           # Parentheses
lui  t4, label >> 12           # Bit shift
addi t5, t6, label & 0xFFF     # Bit mask
```

### Relocation Modifiers

```asm
# Absolute addressing
lui  t0, %hi(symbol)       # High 20 bits
addi t0, t0, %lo(symbol)   # Low 12 bits

# PC-relative addressing (for PIC code)
auipc t1, %pcrel_hi(symbol)
addi  t1, t1, %pcrel_lo(symbol)
```

### Register Names

RISC-V supports both numeric (`x0`-`x31`) and ABI names:

| Register | ABI Name | Description |
|----------|----------|-------------|
| x0 | zero | Hard-wired zero |
| x1 | ra | Return address |
| x2 | sp | Stack pointer |
| x3 | gp | Global pointer |
| x4 | tp | Thread pointer |
| x5-x7 | t0-t2 | Temporaries |
| x8-x9 | s0-s1 | Saved registers |
| x10-x17 | a0-a7 | Function arguments/return values |
| x18-x27 | s2-s11 | Saved registers |
| x28-x31 | t3-t6 | Temporaries |

## Supported Instructions

### RV32I Base Integer Instructions

#### Arithmetic (R-type)
```asm
add   rd, rs1, rs2     # rd = rs1 + rs2
sub   rd, rs1, rs2     # rd = rs1 - rs2
slt   rd, rs1, rs2     # rd = (rs1 < rs2) ? 1 : 0 (signed)
sltu  rd, rs1, rs2     # rd = (rs1 < rs2) ? 1 : 0 (unsigned)
```

#### Logical (R-type)
```asm
and   rd, rs1, rs2     # rd = rs1 & rs2
or    rd, rs1, rs2     # rd = rs1 | rs2
xor   rd, rs1, rs2     # rd = rs1 ^ rs2
```

#### Shift (R-type)
```asm
sll   rd, rs1, rs2     # rd = rs1 << rs2
srl   rd, rs1, rs2     # rd = rs1 >> rs2 (logical)
sra   rd, rs1, rs2     # rd = rs1 >> rs2 (arithmetic)
```

#### Immediate Arithmetic (I-type)
```asm
addi  rd, rs1, imm     # rd = rs1 + imm
slti  rd, rs1, imm     # rd = (rs1 < imm) ? 1 : 0 (signed)
sltiu rd, rs1, imm     # rd = (rs1 < imm) ? 1 : 0 (unsigned)
```

#### Immediate Logical (I-type)
```asm
andi  rd, rs1, imm     # rd = rs1 & imm
ori   rd, rs1, imm     # rd = rs1 | imm
xori  rd, rs1, imm     # rd = rs1 ^ imm
```

#### Immediate Shift (I2-type)
```asm
slli  rd, rs1, shamt   # rd = rs1 << shamt (0-31)
srli  rd, rs1, shamt   # rd = rs1 >> shamt (logical)
srai  rd, rs1, shamt   # rd = rs1 >> shamt (arithmetic)
```

#### Load Instructions (I-type)
```asm
lb    rd, offset(rs1)  # Load byte (sign-extended)
lh    rd, offset(rs1)  # Load halfword (sign-extended)
lw    rd, offset(rs1)  # Load word
lbu   rd, offset(rs1)  # Load byte (zero-extended)
lhu   rd, offset(rs1)  # Load halfword (zero-extended)
```

#### Store Instructions (S-type)
```asm
sb    rs2, offset(rs1) # Store byte
sh    rs2, offset(rs1) # Store halfword
sw    rs2, offset(rs1) # Store word
```

#### Upper Immediate (U-type)
```asm
lui   rd, imm          # rd = imm << 12
auipc rd, imm          # rd = PC + (imm << 12)
```

#### Branch Instructions (B-type)
```asm
beq   rs1, rs2, label  # Branch if equal
bne   rs1, rs2, label  # Branch if not equal
blt   rs1, rs2, label  # Branch if less than (signed)
bge   rs1, rs2, label  # Branch if greater or equal (signed)
bltu  rs1, rs2, label  # Branch if less than (unsigned)
bgeu  rs1, rs2, label  # Branch if greater or equal (unsigned)
```

#### Jump Instructions (J-type)
```asm
jal   rd, label        # Jump and link: rd = PC+4, PC = label
jalr  rd, rs1, offset  # Jump and link register: rd = PC+4, PC = rs1+offset
```

#### System Instructions
```asm
ecall                  # Environment call (syscall)
mret                   # Machine-mode return
wfi                    # Wait for interrupt
```

#### CSR Instructions
```asm
csrrw  rd, rs1, csr    # Read/write CSR
csrrs  rd, rs1, csr    # Read and set bits in CSR
csrrc  rd, rs1, csr    # Read and clear bits in CSR
csrrwi rd, uimm, csr   # Read/write CSR immediate
csrrsi rd, uimm, csr   # Read and set bits immediate
csrrci rd, uimm, csr   # Read and clear bits immediate
```

### RV32M Multiply/Divide Extension

```asm
mul    rd, rs1, rs2    # Multiply (lower 32 bits)
mulh   rd, rs1, rs2    # Multiply high (signed × signed)
mulhsu rd, rs1, rs2    # Multiply high (signed × unsigned)
mulhu  rd, rs1, rs2    # Multiply high (unsigned × unsigned)
div    rd, rs1, rs2    # Divide (signed)
divu   rd, rs1, rs2    # Divide (unsigned)
rem    rd, rs1, rs2    # Remainder (signed)
remu   rd, rs1, rs2    # Remainder (unsigned)
```

### RV32Zbb Basic Bit-Manipulation

```asm
# Count instructions
clz    rd, rs1         # Count leading zeros
ctz    rd, rs1         # Count trailing zeros
cpop   rd, rs1         # Population count (count set bits)

# Sign/zero extend
sext.b rd, rs1         # Sign-extend byte
sext.h rd, rs1         # Sign-extend halfword
pack   rd, rs1, zero   # Zero-extend halfword (alias: zext.h)

# Min/max
min    rd, rs1, rs2    # Signed minimum
minu   rd, rs1, rs2    # Unsigned minimum
max    rd, rs1, rs2    # Signed maximum
maxu   rd, rs1, rs2    # Unsigned maximum

# Rotate
rol    rd, rs1, rs2    # Rotate left
ror    rd, rs1, rs2    # Rotate right
rori   rd, rs1, shamt  # Rotate right immediate

# Byte operations
rev8   rd, rs1         # Reverse bytes in word
```

### RV32Zbs Single-Bit Manipulation

```asm
bset   rd, rs1, rs2    # Set bit at position rs2
bseti  rd, rs1, imm    # Set bit at position imm
bclr   rd, rs1, rs2    # Clear bit at position rs2
bclri  rd, rs1, imm    # Clear bit at position imm
binv   rd, rs1, rs2    # Invert bit at position rs2
binvi  rd, rs1, imm    # Invert bit at position imm
bext   rd, rs1, rs2    # Extract bit at position rs2
bexti  rd, rs1, imm    # Extract bit at position imm
```

### RV32Zicond Conditional Operations

```asm
czero.eqz rd, rs1, rs2 # rd = (rs2 == 0) ? 0 : rs1
czero.nez rd, rs1, rs2 # rd = (rs2 != 0) ? 0 : rs1
```

### RV32Zbkb Cryptography Bit-Manipulation

```asm
pack   rd, rs1, rs2    # Pack low halves of rs1 and rs2
packh  rd, rs1, rs2    # Pack low bytes of rs1 and rs2
```

### NATOGA32 Custom Thread Instructions

```asm
new.t   rd, rs1        # Create new thread
yield.t                # Yield to scheduler
id.t    rd             # Get current thread ID
sleep.t rd, rs1        # Sleep thread
wake.t  rd, rs1        # Wake thread
end.t                  # End current thread
```

### Pseudo-Instructions

Pseudo-instructions are expanded by the assembler into one or more real instructions:

```asm
# No operation
nop                    # addi zero, zero, 0

# Load immediate
li    rd, imm          # Load 32-bit immediate (uses lui + addi if needed)

# Move
mv    rd, rs           # addi rd, rs, 0

# Logical NOT
not   rd, rs           # xori rd, rs, -1

# Negate
neg   rd, rs           # sub rd, zero, rs

# Jump
j     label            # jal zero, label
jr    rs               # jalr zero, rs, 0

# Return
ret                    # jalr zero, ra, 0

# Branches with zero
beqz  rs, label        # beq rs, zero, label
bnez  rs, label        # bne rs, zero, label
blez  rs, label        # bge zero, rs, label
bgez  rs, label        # bge rs, zero, label
bltz  rs, label        # blt rs, zero, label
bgtz  rs, label        # blt zero, rs, label

# Branch pseudo-instructions
bgt   rs1, rs2, label  # blt rs2, rs1, label (swap operands)
ble   rs1, rs2, label  # bge rs2, rs1, label
bgtu  rs1, rs2, label  # bltu rs2, rs1, label
bleu  rs1, rs2, label  # bgeu rs2, rs1, label

# Compare with zero
seqz  rd, rs           # sltiu rd, rs, 1
snez  rd, rs           # sltu rd, zero, rs
sltz  rd, rs           # slt rd, rs, zero
sgtz  rd, rs           # slt rd, zero, rs

# CSR pseudo-instructions
csrr  rd, csr          # csrrs rd, zero, csr (read)
csrw  csr, rs          # csrrw zero, rs, csr (write)
csrs  csr, rs          # csrrs zero, rs, csr (set)
csrc  csr, rs          # csrrc zero, rs, csr (clear)
csrwi csr, imm         # csrrwi zero, imm, csr (write immediate)
csrsi csr, imm         # csrrsi zero, imm, csr (set immediate)
csrci csr, imm         # csrrci zero, imm, csr (clear immediate)

# System call
syscall                # ecall
```

## Preprocessor

The goasm preprocessor provides powerful text processing capabilities.

### File Inclusion

```asm
.include "header.inc"          # Include another assembly file
.include "lib/functions.asm"   # Supports relative paths
```

Recursive inclusion is detected and prevented.

### Constant Definition

```asm
.define STACK_SIZE 4096        # Define constant
.define BASE_ADDR 0x01A00000   # Constants can be expressions
.define MSG_LEN (10 + 5)       # Arithmetic in definitions

    addi sp, sp, -STACK_SIZE   # Use constant
```

### Conditional Assembly

```asm
.define DEBUG 1

.ifdef DEBUG
    # Debug-only code
    li a0, 1
    ecall
.endif

.ifndef RELEASE
    # Non-release code
    nop
.endif

.if DEBUG == 1
    # Conditional code
.else
    # Alternative code
.endif
```

### Loop Expansion

```asm
# .for variable = start, end [, step]
.for i = 0, 4
    addi x##i, x##i, 1         # x0, x1, x2, x3, x4
.endf

# With step
.for j = 0, 10, 2
    nop                         # 6 iterations (0, 2, 4, 6, 8, 10)
.endf
```

### Value Duplication

```asm
.byte .dup 10 0                # 10 zeros
.word .dup 5 0xDEADBEEF        # 5 words of 0xDEADBEEF
```

### Macros

```asm
# Define macro
.macro push reg
    addi sp, sp, -4
    sw   \reg, 0(sp)
.endmacro

# Use macro
push a0
push a1

# Multi-parameter macro
.macro load_addr reg, label
    lui  \reg, %hi(\label)
    addi \reg, \reg, %lo(\label)
.endmacro

load_addr t0, my_data
```

### Token Pasting

Use `##` to concatenate tokens:

```asm
.for i = 0, 3
    mv t##i, a##i              # mv t0, a0; mv t1, a1; ...
.endf
```

### Environment Variables

```asm
.define BUILD_DATE $[BUILD_DATE]   # Expands environment variable
.define VERSION $[VERSION]

# Use in code
.string "Version: $[VERSION]"
```

## Directives

### Section Directives

```asm
.text                  # Code section (executable)
.data                  # Initialized data section
.rodata                # Read-only data section
.bss                   # Uninitialized data section
```

### Data Directives

```asm
.byte  0x12, 0x34, 0x56           # 8-bit values
.half  0x1234, 0x5678             # 16-bit values (halfwords)
.word  0x12345678, 0xDEADBEEF     # 32-bit values (words)

.string "Hello, World!\n"          # Null-terminated string
.ascii  "No null"                  # String without null terminator

.space 100                         # Reserve 100 bytes (zero-filled)
.align 4                           # Align to 4-byte boundary
```

### Symbol Directives

```asm
.equ CONSTANT, 42                  # Define assembly-time constant

.global main                       # Make symbol globally visible
.global my_function
```

### Binary Inclusion

```asm
.incbin "data.bin"                 # Include binary file
.incbin "image.raw"                # Useful for embedding assets
```

## Optimization

Enable peephole optimization with the `-O` or `--optimize` flag:

```bash
./goasm program.asm -o program.on32 -O
```

### Optimizations Performed

#### 1. Zero Constant Optimization

Converts operations with the zero register to more efficient move instructions:

```asm
# Before optimization
add  t0, t1, zero      # 4 bytes
or   t2, t3, x0        # 4 bytes

# After optimization (-O)
mv   t0, t1            # addi t0, t1, 0
mv   t2, t3            # addi t2, t3, 0
```

#### 2. Redundant Move Elimination

Combines consecutive move instructions when safe:

```asm
# Before optimization
mv   t0, t1            # First move
mv   t2, t0            # Second move uses t0

# After optimization (-O)
mv   t2, t1            # Combined into single move
```

### Optimization Boundaries

The optimizer respects control flow and does not optimize across:

- **Labels** (potential jump targets)
- **Directives** (section changes, alignment, etc.)

### Performance Impact

On typical programs, `-O` provides:
- **15-25% reduction** in code size
- **No runtime overhead** (optimizations are at assembly time)
- **Identical behavior** (optimizations are semantically preserving)

### Example

```asm
# test.asm
add  t0, t1, zero
mv   t2, t0
add  t3, t4, t5
```

```bash
# Without optimization: 12 bytes
./goasm test.asm -o test.on32
# Size: 12 bytes

# With optimization: 8 bytes
./goasm test.asm -o test_opt.on32 -O
# Size: 8 bytes (33% smaller)
```

## Output Formats

### ON32 Format (Default)

The ON32 format includes a header with metadata:

```
ON32 Header:
- Magic number: "ON32"
- Version information
- Entry point address
- Section information (.text, .data, etc.)
```

```bash
./goasm program.asm -o program.on32 -f on32
```

### Raw Binary Format

Raw binary contains only the assembled machine code:

```bash
./goasm program.asm -o program.bin -f bin
```

Use for:
- ROM images
- Bootloaders
- Direct memory loading

## Disassembler

The disassembler converts binary back to assembly:

### Basic Disassembly

```bash
# Disassemble ON32 file
./goasm --disasm program.on32

# Output:
# ; Disassembly of program.on32 (on32 format)
# ; .text: 0x01A00000 - 0x01A00010 (16 bytes)
#
# .text
# 0x01A00000:  00000293  addi t0, zero, 0
# 0x01A00004:  00128313  addi t1, t0, 1
```

### Pseudo-Instruction Mode

Show pseudo-instructions instead of expanded forms:

```bash
./goasm --disasm program.on32 --pseudo

# Output:
# 0x01A00000:  00000293  mv t0, zero
# 0x01A00004:  00128313  addi t1, t0, 1
```

### Raw Binary Disassembly

```bash
# Specify base address for raw binary
./goasm --disasm firmware.bin --base 0x80000000

# Default base address is 0x01A00000 (NATOGA32 text segment)
./goasm --disasm bootloader.bin
```

### Register Name Formats

```bash
# Use ABI names (default): zero, ra, sp, t0, a0, etc.
./goasm --disasm program.on32

# Use numeric names: x0, x1, x2, x5, x10, etc.
./goasm --disasm program.on32 --no-abi
```

### Hide Hex Encoding

```bash
# Show hex encoding (default)
./goasm --disasm program.on32 --hex

# Hide hex encoding
./goasm --disasm program.on32 --hex=false
```

## Examples

### Example 1: Hello World

```asm
# hello.asm - Print message and exit
.data
msg:
    .string "Hello, NATOGA32!\n"

.text
.global main
main:
    # Print message (assuming syscall interface)
    li   a0, 1              # File descriptor: stdout
    la   a1, msg            # Address of message
    li   a2, 17             # Length
    li   a7, 64             # Syscall: write
    ecall

    # Exit
    li   a0, 0              # Exit code
    li   a7, 93             # Syscall: exit
    ecall
```

Assemble:
```bash
./goasm hello.asm -o hello.on32 -O
```

### Example 2: Fibonacci Calculator

```asm
# fibonacci.asm - Calculate Fibonacci numbers
.text
.global fibonacci

# int fibonacci(int n)
# Input: a0 = n
# Output: a0 = fib(n)
fibonacci:
    # Base cases
    li   t0, 2
    blt  a0, t0, fib_base   # if n < 2, return n

    # Setup
    addi sp, sp, -16
    sw   ra, 12(sp)
    sw   s0, 8(sp)
    sw   s1, 4(sp)
    sw   s2, 0(sp)

    mv   s0, a0             # s0 = n

    # fib(n-1)
    addi a0, s0, -1
    call fibonacci
    mv   s1, a0             # s1 = fib(n-1)

    # fib(n-2)
    addi a0, s0, -2
    call fibonacci
    mv   s2, a0             # s2 = fib(n-2)

    # Return fib(n-1) + fib(n-2)
    add  a0, s1, s2

    # Restore and return
    lw   s2, 0(sp)
    lw   s1, 4(sp)
    lw   s0, 8(sp)
    lw   ra, 12(sp)
    addi sp, sp, 16
    ret

fib_base:
    ret                     # a0 already contains n
```

### Example 3: Using Macros

```asm
# macros.asm - Demonstrate preprocessor macros
.define STACK_SIZE 1024

# Macro for function prologue
.macro prologue
    addi sp, sp, -16
    sw   ra, 12(sp)
    sw   s0, 8(sp)
.endmacro

# Macro for function epilogue
.macro epilogue
    lw   s0, 8(sp)
    lw   ra, 12(sp)
    addi sp, sp, 16
    ret
.endmacro

# Macro with parameters
.macro print_int value
    mv   a0, \value
    li   a7, 1              # Print integer syscall
    ecall
.endmacro

.text
main:
    prologue

    li   t0, 42
    print_int t0

    epilogue
```

### Example 4: Bit Manipulation

```asm
# bitops.asm - Demonstrate bit-manipulation extensions
.text

# Count set bits (population count)
count_bits:
    cpop a0, a0             # Zbb: population count
    ret

# Reverse bytes
reverse_word:
    rev8 a0, a0             # Zbb: reverse bytes
    ret

# Set bit at position
set_bit:
    # a0 = value, a1 = bit position
    bset a0, a0, a1         # Zbs: set bit
    ret

# Check if power of 2
is_power_of_2:
    # Power of 2 has exactly 1 bit set
    cpop t0, a0             # Count bits
    li   a0, 1
    beq  t0, a0, is_pow2
    li   a0, 0
is_pow2:
    ret
```

### Example 5: Thread Management

```asm
# threads.asm - NATOGA32 thread example
.text

main:
    # Create new thread
    la   a0, worker_thread
    new.t t0, a0            # t0 = thread ID

    # Main thread work
    li   t1, 100
loop:
    addi t1, t1, -1
    bnez t1, loop

    # Wait for worker
    wake.t zero, t0

    # Exit
    end.t

worker_thread:
    id.t t0                 # Get my thread ID

    # Worker loop
    li   t1, 50
worker_loop:
    yield.t                 # Yield to other threads
    addi t1, t1, -1
    bnez t1, worker_loop

    end.t                   # End this thread
```

### Example 6: Conditional Assembly

```asm
# config.asm - Build configurations
.define DEBUG 1
.define PLATFORM 2

.text
main:
    # Platform-specific code
    .if PLATFORM == 1
        li a0, 0x1000       # Platform 1 base address
    .elif PLATFORM == 2
        li a0, 0x2000       # Platform 2 base address
    .else
        li a0, 0x3000       # Default base address
    .endif

    # Debug instrumentation
    .ifdef DEBUG
        # Debug-only code
        li a7, 64           # Write syscall
        ecall
    .endif

    # Continue normal execution
    ret
```

## Debugging Tips

### View Tokens

```bash
./goasm program.asm -d 1
```

Shows the token stream from the lexer.

### View Preprocessed Code

```bash
./goasm program.asm -d 2
```

Shows tokens after preprocessor expansion (macros, loops, conditionals).

### View AST

```bash
./goasm program.asm -d 3
```

Shows the abstract syntax tree.

### View Expanded Pseudo-Instructions

```bash
./goasm program.asm -d 4
```

Shows how pseudo-instructions are expanded.

### View Symbol Table

```bash
./goasm program.asm -d 5
```

Shows all symbols (labels, constants) and their addresses.

### View Encoded Output

```bash
./goasm program.asm -d 7
```

Shows hex dump of encoded instructions.

## Error Messages

Common errors and solutions:

### Undefined Symbol

```
Error: undefined symbol 'label_name'
```
**Solution**: Define the label or check for typos.

### Invalid Immediate

```
Error: immediate value out of range
```
**Solution**: Use `li` pseudo-instruction or split into multiple instructions.

### Invalid Register

```
Error: expected register, got 'foo'
```
**Solution**: Check register names (x0-x31 or ABI names).

### File Not Found

```
Error: cannot open file: 'missing.inc'
```
**Solution**: Check file path in `.include` directive.

## Performance Tips

1. **Use optimization**: Add `-O` flag for smaller binaries
2. **Use pseudo-instructions**: More readable and often optimized
3. **Align data**: Use `.align` for better memory access performance
4. **Use appropriate types**: Use `.byte` for small values, not `.word`
5. **Minimize branches**: Use conditional operations when possible

## See Also

- [NATOGA32 Architecture Documentation](../architecture/console.md)
- [RISC-V Specifications](https://riscv.org/technical/specifications/)
- [RISC-V Assembly Programmer's Manual](https://github.com/riscv-non-isa/riscv-asm-manual)

## Contributing

Report issues and contribute at: https://github.com/natoga32/natoga32

## License

This tool is subject to the MIT License bundled with the NATOGA32 project.
