# goasm Quick Reference

## Command Line

```bash
# Assembly
goasm input.asm -o output.on32              # Assemble to ON32
goasm input.asm -o output.bin -f bin        # Assemble to raw binary
goasm input.asm -o output.on32 -O           # Assemble with optimization

# Disassembly
goasm --disasm program.on32                 # Disassemble ON32
goasm --disasm program.on32 --pseudo        # Show pseudo-instructions
goasm --disasm program.bin --base 0x1000    # Disassemble raw binary

# Debugging
goasm input.asm -d 1                        # Show tokens
goasm input.asm -d 3                        # Show AST
goasm input.asm -d 5                        # Show symbols
```

## Instruction Quick Reference

### Arithmetic

| Instruction | Operation | Example |
|-------------|-----------|---------|
| `add rd, rs1, rs2` | rd = rs1 + rs2 | `add t0, t1, t2` |
| `addi rd, rs1, imm` | rd = rs1 + imm | `addi sp, sp, -16` |
| `sub rd, rs1, rs2` | rd = rs1 - rs2 | `sub t0, t1, t2` |
| `lui rd, imm` | rd = imm << 12 | `lui t0, 0x12345` |
| `li rd, imm` | Load immediate | `li a0, 1000` |

### Logical

| Instruction | Operation | Example |
|-------------|-----------|---------|
| `and rd, rs1, rs2` | rd = rs1 & rs2 | `and t0, t1, t2` |
| `andi rd, rs1, imm` | rd = rs1 & imm | `andi t0, t1, 0xFF` |
| `or rd, rs1, rs2` | rd = rs1 \| rs2 | `or t0, t1, t2` |
| `ori rd, rs1, imm` | rd = rs1 \| imm | `ori t0, t1, 0x80` |
| `xor rd, rs1, rs2` | rd = rs1 ^ rs2 | `xor t0, t1, t2` |
| `xori rd, rs1, imm` | rd = rs1 ^ imm | `xori t0, t1, -1` |
| `not rd, rs` | Bitwise NOT | `not t0, t1` |

### Shifts

| Instruction | Operation | Example |
|-------------|-----------|---------|
| `sll rd, rs1, rs2` | Shift left logical | `sll t0, t1, t2` |
| `slli rd, rs1, shamt` | Shift left immediate | `slli t0, t1, 5` |
| `srl rd, rs1, rs2` | Shift right logical | `srl t0, t1, t2` |
| `srli rd, rs1, shamt` | Shift right immediate | `srli t0, t1, 3` |
| `sra rd, rs1, rs2` | Shift right arithmetic | `sra t0, t1, t2` |
| `srai rd, rs1, shamt` | Shift right arith imm | `srai t0, t1, 7` |

### Compare

| Instruction | Operation | Example |
|-------------|-----------|---------|
| `slt rd, rs1, rs2` | Set if less than (signed) | `slt t0, t1, t2` |
| `slti rd, rs1, imm` | Set if less than imm | `slti t0, t1, 100` |
| `sltu rd, rs1, rs2` | Set if less (unsigned) | `sltu t0, t1, t2` |
| `sltiu rd, rs1, imm` | Set if less imm (unsigned) | `sltiu t0, t1, 50` |
| `seqz rd, rs` | Set if equal zero | `seqz t0, t1` |
| `snez rd, rs` | Set if not equal zero | `snez t0, t1` |

### Memory

| Instruction | Operation | Example |
|-------------|-----------|---------|
| `lw rd, offset(rs1)` | Load word | `lw t0, 0(sp)` |
| `lh rd, offset(rs1)` | Load halfword (signed) | `lh t0, 4(a0)` |
| `lb rd, offset(rs1)` | Load byte (signed) | `lb t0, 8(a1)` |
| `lhu rd, offset(rs1)` | Load halfword (unsigned) | `lhu t0, 2(a2)` |
| `lbu rd, offset(rs1)` | Load byte (unsigned) | `lbu t0, 1(a3)` |
| `sw rs2, offset(rs1)` | Store word | `sw t0, 0(sp)` |
| `sh rs2, offset(rs1)` | Store halfword | `sh t0, 4(a0)` |
| `sb rs2, offset(rs1)` | Store byte | `sb t0, 8(a1)` |

### Branches

| Instruction | Condition | Example |
|-------------|-----------|---------|
| `beq rs1, rs2, label` | Branch if equal | `beq t0, t1, loop` |
| `bne rs1, rs2, label` | Branch if not equal | `bne t0, zero, skip` |
| `blt rs1, rs2, label` | Branch if less (signed) | `blt t0, t1, done` |
| `bge rs1, rs2, label` | Branch if >= (signed) | `bge t0, t1, end` |
| `bltu rs1, rs2, label` | Branch if less (unsigned) | `bltu a0, a1, loop` |
| `bgeu rs1, rs2, label` | Branch if >= (unsigned) | `bgeu a0, a1, exit` |
| `beqz rs, label` | Branch if zero | `beqz t0, skip` |
| `bnez rs, label` | Branch if not zero | `bnez t0, loop` |

### Jumps

| Instruction | Operation | Example |
|-------------|-----------|---------|
| `jal rd, label` | Jump and link | `jal ra, function` |
| `jalr rd, rs1, offset` | Jump and link register | `jalr ra, t0, 0` |
| `j label` | Jump | `j loop` |
| `jr rs` | Jump register | `jr t0` |
| `call label` | Call function | `call printf` |
| `ret` | Return from function | `ret` |

### Multiply/Divide (RV32M)

| Instruction | Operation | Example |
|-------------|-----------|---------|
| `mul rd, rs1, rs2` | Multiply (low 32 bits) | `mul t0, t1, t2` |
| `mulh rd, rs1, rs2` | Multiply high (s×s) | `mulh t0, t1, t2` |
| `div rd, rs1, rs2` | Divide (signed) | `div t0, t1, t2` |
| `divu rd, rs1, rs2` | Divide (unsigned) | `divu t0, t1, t2` |
| `rem rd, rs1, rs2` | Remainder (signed) | `rem t0, t1, t2` |
| `remu rd, rs1, rs2` | Remainder (unsigned) | `remu t0, t1, t2` |

### Bit Manipulation (Zbb/Zbs)

| Instruction | Operation | Example |
|-------------|-----------|---------|
| `clz rd, rs` | Count leading zeros | `clz t0, a0` |
| `ctz rd, rs` | Count trailing zeros | `ctz t0, a0` |
| `cpop rd, rs` | Count set bits | `cpop t0, a0` |
| `bset rd, rs1, rs2` | Set bit | `bset t0, t1, t2` |
| `bclr rd, rs1, rs2` | Clear bit | `bclr t0, t1, t2` |
| `binv rd, rs1, rs2` | Invert bit | `binv t0, t1, t2` |
| `bext rd, rs1, rs2` | Extract bit | `bext t0, t1, t2` |
| `rol rd, rs1, rs2` | Rotate left | `rol t0, t1, t2` |
| `ror rd, rs1, rs2` | Rotate right | `ror t0, t1, t2` |
| `rev8 rd, rs` | Reverse bytes | `rev8 t0, a0` |
| `min rd, rs1, rs2` | Minimum (signed) | `min t0, t1, t2` |
| `max rd, rs1, rs2` | Maximum (signed) | `max t0, t1, t2` |

## Registers

### Numeric Names
`x0` - `x31` (x0 is always zero)

### ABI Names

| Register | ABI Name | Usage | Saved |
|----------|----------|-------|-------|
| x0 | zero | Hard-wired zero | n/a |
| x1 | ra | Return address | No |
| x2 | sp | Stack pointer | Yes |
| x3 | gp | Global pointer | n/a |
| x4 | tp | Thread pointer | n/a |
| x5-x7 | t0-t2 | Temporaries | No |
| x8-x9 | s0-s1 | Saved registers | Yes |
| x10-x11 | a0-a1 | Args/return values | No |
| x12-x17 | a2-a7 | Arguments | No |
| x18-x27 | s2-s11 | Saved registers | Yes |
| x28-x31 | t3-t6 | Temporaries | No |

## Preprocessor

### Directives

```asm
.include "file.asm"                    # Include file
.define NAME value                     # Define constant
.equ NAME, value                       # Define constant (synonym)

.ifdef NAME ... .endif                 # If defined
.ifndef NAME ... .endif                # If not defined
.if expr ... .elif expr ... .else ... .endif   # If expression

.for var = start, end, step ... .endf  # Loop expansion
.macro name params ... .endmacro       # Define macro

.dup count value                       # Duplicate value
```

### Token Pasting

```asm
.for i = 0, 3
    mv t##i, a##i                      # Expands to: mv t0, a0; mv t1, a1; ...
.endf
```

### Macro Usage

```asm
# Define
.macro push reg
    addi sp, sp, -4
    sw   \reg, 0(sp)
.endmacro

# Use
push t0
push a0
```

## Assembler Directives

```asm
.text                   # Code section
.data                   # Data section
.rodata                 # Read-only data section
.bss                    # Uninitialized data section

.byte   val, ...        # 8-bit values
.half   val, ...        # 16-bit values
.word   val, ...        # 32-bit values

.string "text"          # Null-terminated string
.ascii  "text"          # String (no null)

.space  n               # Reserve n bytes
.align  n               # Align to n-byte boundary

.global symbol          # Make symbol global
.incbin "file"          # Include binary file
```

## Common Patterns

### Function Prologue/Epilogue

```asm
function:
    addi sp, sp, -16        # Allocate stack frame
    sw   ra, 12(sp)         # Save return address
    sw   s0, 8(sp)          # Save saved registers
    # ... function body ...
    lw   s0, 8(sp)          # Restore saved registers
    lw   ra, 12(sp)         # Restore return address
    addi sp, sp, 16         # Deallocate stack frame
    ret                     # Return
```

### Load 32-bit Address

```asm
    lui  t0, %hi(label)     # Load high 20 bits
    addi t0, t0, %lo(label) # Add low 12 bits
# Or use pseudo-instruction:
    la   t0, label          # Load address
```

### Load 32-bit Immediate

```asm
    li   t0, 0x12345678     # Pseudo-instruction (auto-expanded)
# Expands to:
#   lui  t0, 0x12345
#   addi t0, t0, 0x678
```

### Conditional Move

```asm
# if (a0 != 0) a1 = a2
    beqz a0, skip
    mv   a1, a2
skip:
# Or with Zicond:
    czero.eqz a1, a2, a0    # a1 = (a0 == 0) ? 0 : a2
```

### Array Access

```asm
# array[i] where array is in a0, i is in a1
    slli t0, a1, 2          # t0 = i * 4 (word size)
    add  t0, a0, t0         # t0 = &array[i]
    lw   t1, 0(t0)          # t1 = array[i]
```

### Loop Pattern

```asm
    li   t0, 10             # Counter
loop:
    # Loop body
    addi t0, t0, -1         # Decrement counter
    bnez t0, loop           # Continue if not zero
```

## Optimization Tips

- Use `-O` flag for smaller code
- Use `mv` instead of `add rd, rs, zero`
- Use `li` for loading immediates (auto-optimized)
- Prefer pseudo-instructions (more readable and optimized)
- Use `.align` for data structures
- Use appropriate data sizes (`.byte` vs `.word`)

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| Undefined symbol | Label not defined | Define label or check spelling |
| Immediate out of range | Value too large for instruction | Use `li` or split into multiple ops |
| Invalid register | Wrong register name | Use x0-x31 or ABI names |
| Unexpected token | Syntax error | Check instruction format |
| Circular include | `.include` loop | Remove circular dependencies |
| Macro recursion | Macro calls itself | Fix macro definition |

## Memory Map (NATOGA32)

| Section | Base Address | Usage |
|---------|--------------|-------|
| .text | 0x01A00000 | Code (executable) |
| .data | 0x02000000 | Initialized data |
| .bss | 0x02800000 | Uninitialized data |
| .rodata | 0x03000000 | Read-only data |

## Examples

### Minimal Program
```asm
.text
main:
    li a0, 0
    li a7, 93       # Exit syscall
    ecall
```

### Function Call
```asm
main:
    li   a0, 10
    call factorial
    # Result in a0
    ret

factorial:
    # ... implementation ...
    ret
```

### Data Section
```asm
.data
array:
    .word 1, 2, 3, 4, 5
count:
    .word 5
message:
    .string "Hello!\n"

.text
main:
    la   t0, array
    lw   t1, 0(t0)
    # ...
```
