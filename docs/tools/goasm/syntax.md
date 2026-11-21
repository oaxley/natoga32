# goasm syntax

This page describes the source syntax of goasm: tokens, expressions, registers, sections and data directives.

A short program:

```asm
; Sum 1..10 and store the result in .data
.equ LIMIT, 10

.text
start:
    li   t0, 1              ; counter
    li   t1, LIMIT
    li   a0, 0              ; sum
loop:
    add  a0, a0, t0
    addi t0, t0, 1
    ble  t0, t1, loop
    la   t2, result
    sw   a0, 0(t2)
    end.t

.data
result:
    .word 0
message:
    .string "done"
```

## Lines, comments and labels

One statement per line.
A comment starts with `;`, `#` or `//` and runs to the end of the line.
(`##` is not a comment: it is the [paste operator](preprocessor.md#token-pasting).)

A label is a name followed by `:`.
A label can share a line with an instruction or a directive.
A name can contain letters, digits, `_` and `.`.
A label defined twice stops with `symbol 'l' already defined at line 0`. The line number is always 0.

Instruction and directive names are case-sensitive and lower case: `ADDI` gives `unknown instruction 'ADDI'`, and `.TEXT` gives `unknown directive`.
Register names are not case-sensitive.

## Numbers and characters

| Form | Example |
|------|---------|
| Decimal | `42`, `-8` |
| Hexadecimal | `0x2A` |
| Binary | `0b101010` |
| Octal | `0o52` |
| Character | `'A'`, `'\n'` |

A `_` inside a number is **not** supported.
Write `0x206800` in source, not `0x0020_6800`.

A character literal is a number.
The escapes are `\n`, `\t`, `\r`, `\0`, `\\` and `\'`.

Strings use double quotes and are used by `.string`, `.ascii`, `.asciz`, `.include` and `.incbin`.
In `.string` and `.ascii`, an escape such as `\n` is **not** translated: the bytes of `"a\nb"` are `61 5C 6E 62 00`.
To put a newline in a string, end the string and add `.byte 10`.

## Expressions

Instruction operands and data directives accept expressions with numbers, characters, labels and `.equ` names.
`$` is the address of the current statement. It works everywhere an expression is accepted: in `.byte`/`.half`/`.word`, in branch and `jal` targets (`j $+8`), and in immediates (`addi a1, a1, $`, `lw a0, $(sp)`). With `li`, `$` is the address of the second instruction of the expansion (the `li` address + 4).

| Operator | Precedence |
|----------|------------|
| `*` `/` `%` | 6 (highest) |
| `+` `-` | 5 |
| `<<` `>>` | 4 |
| `&` | 3 |
| `^` | 2 |
| `\|` | 1 (lowest) |

Unary `+` and `-` and parentheses also work.
Every operator works in instruction operands (immediates, branch and `jal` targets) and in `.byte`, `.half` and `.word`.
Values are computed in 64 bits, then cut to the size of the field. `>>` keeps the sign: `-8 >> 1` is `-4`.

In an instruction operand and in `.byte`, `.half` and `.word`, these stop with an error:

| Error | Cause |
|-------|-------|
| `division by zero` | `/` by 0. |
| `modulo by zero` | `%` by 0. |
| `shift amount N out of range (0..63)` | `<<` or `>>` by less than 0 or more than 63. |
| `undefined symbol 'X'` | A name that is not a label or an `.equ`, also one that is only listed in `.global`. |
| `string "x" cannot be used in an expression` | A string where a number is expected. |

A value that does not fit its field stops with an error; it is never cut:

| Field | Range | Error |
|-------|-------|-------|
| `addi`, logic, loads, stores, `jalr` | -2048..2047 | `addi immediate 5000 out of range (-2048..2047)` |
| `lui`, `auipc` | -524288..1048575 (signed or unsigned 20 bits) | `lui immediate N out of range (...)` |
| CSR number of `csrrw csrrs csrrc` | 0..0xFFF | `CSR number 0x1000 out of range (0..0xfff)` |
| Branch offset (`beq`..., `beqz`...) | -4096..4094, even | `beq offset 5000 out of range (-4096..4094)`, `beq offset 3 is odd` |
| `jal` / `j` offset | -1048576..1048574, even | `jal offset N out of range (-1048576..1048574)` |

A [relocation modifier](#relocation-modifiers) (`%lo`, `%hi`, `%pcrel_lo`, `%pcrel_hi`) is already the bits of its field and is not checked: `addi a0, a0, %lo(X)` is fine when `%lo(X)` is 4095.
An error in an instruction made by a pseudo-instruction (`beqz`, `j`, `li`...) is reported at the line of the pseudo-instruction, with the name of the real instruction (`beq`, `jal`, `lui`).

- The shift amount of `slli srli srai rori bseti bclri binvi bexti` must be 0..31; other values are an error.

## Relocation modifiers

| Modifier | Value |
|----------|-------|
| `%hi(x)` | `(x + 0x800) >> 12`, for `lui`. |
| `%lo(x)` | `x & 0xFFF`, for `addi`, loads and stores. |
| `%pcrel_hi(x)` | `((x - pc) + 0x800) >> 12`, for `auipc`. |
| `%pcrel_lo(x)` | The low 12 bits of the offset of the paired `auipc`. |

`%pcrel_lo(x)` must be the whole immediate of an I-type or S-type instruction.
Its `auipc` is either the previous instruction with the same `%pcrel_hi(x)`, or the `auipc` that carries the label `x`.
Otherwise goasm stops with `%pcrel_lo(X) has no paired auipc`.
`la` and `call` generate this pair for you.

## Registers

Each register has a number name and an ABI name.

| Registers | ABI names |
|-----------|-----------|
| `x0` | `zero` |
| `x1` | `ra` |
| `x2` | `sp` |
| `x3` | `gp` |
| `x4` | `tp` |
| `x5`-`x7` | `t0`-`t2` |
| `x8` | `s0`, `fp` |
| `x9` | `s1` |
| `x10`-`x17` | `a0`-`a7` |
| `x18`-`x27` | `s2`-`s11` |
| `x28`-`x31` | `t3`-`t6` |

A name that is not a register, where a register is expected, stops with `'a99' is not a register` and its line and column (`x32`, `s12` and labels too).
A number in place of a register stops with `expected a register`.

Memory operands are written `offset(base)`.
`(base)` means `0(base)`.

## Sections

Five sections exist.
The default section is `.text`.
Each section has a fixed window in the [memory map](../../console/memory-map.md).

| Directive | Base | Window size | Content |
|-----------|------|-------------|---------|
| `.rom` | `0x0000_0000` | 2 MiB | Boot ROM code and data. |
| `.bss` | `0x0100_0000` | 2 MiB | Zero-filled RAM. No bytes in the file. |
| `.rodata` | `0x0120_0000` | 2 MiB | Read-only data. |
| `.data` | `0x0140_0000` | 6 MiB | Initialised data. |
| `.text` | `0x01A0_0000` | 6 MiB | Code. |

A section can be opened several times.
The pieces are joined in order.

After pass 1, goasm checks the sizes.
A section that is too big stops the build before any output:

```text
section .bss is 3000000 bytes, exceeds its 2MiB window (0x01000000-0x011FFFFF)
```

A `.bss` section can hold `.space` only.
A non-zero value stops with `.bss section holds non-zero data at 0xADDR: use .data`.

## Data directives

| Directive | Effect |
|-----------|--------|
| `.byte a, b, ...` | 1 byte per value. |
| `.half a, ...` / `.short` | 2 bytes per value, little-endian. |
| `.word a, ...` | 4 bytes per value, little-endian. |
| `.string "s"` / `.asciz` | The bytes of `s`, then a 0 byte. |
| `.ascii "s"` | The bytes of `s`, no 0 byte. |
| `.space N` / `.skip N` | `N` zero bytes. `N` must not be negative. |
| `.align N` | Pad with zeros to a multiple of `2^N`. `N` is 0..31. |
| `.org ADDR` | Pad with zeros up to the **absolute** address `ADDR`. |
| `.equ NAME, VALUE` / `.set` | Define a constant. |
| `.global NAME` / `.globl` | Mark a label as global in the `dSYM` chunk. |
| `.incbin "file"` | Insert the bytes of a file. |

Notes:

- The operand of `.space`, `.align`, `.org` and the value of `.equ` are [expressions](#expressions): `.equ MASK, (1 << 4) - 1`, `.equ NEG, -2`, `.equ LETTER, 'A'`, `.space SIZE * 2`, `.org $ + 16`.
- These four are computed while goasm lays out the addresses, so they can only use labels and `.equ` names defined **above** them. A name defined later stops with `undefined symbol`. `.byte`, `.half`, `.word` and instructions can use names defined anywhere.
- An `.equ` name works everywhere an expression is accepted. Defining it twice stops with `symbol 'X' already defined`.
- A wrong operand count stops with `.space expects 1 operand` (or `.align`, `.org`), or `.equ expects 2 operands (name, value)`. `.space -1` stops with `size -1 is negative`, `.align 32` with `out of range (0..31)`.
- `.org` takes an absolute address, not an offset: `.org 0x100` in `.rom`, `.org 0x01A00010` in `.text`. An address before the current one stops with `.org address 0xX is before current address 0xY`.
- `.global` must come **after** the label it marks. Before the label, it has no effect.
- `.incbin` reads the file relative to the directory of the input file. The bytes count in the section size. A missing file stops with `cannot stat incbin file`.
- A name that is not a directive stops with `error: unknown directive '.wrod'`, with file, line and column.

An undefined label stops with `file:line:col: error: undefined symbol 'X'`.

## Operand forms

| Instruction group | Form |
|-------------------|------|
| Loads | `lw rd, off(rs1)` or `lw rd, (rs1)` |
| Stores | `sw rs2, off(rs1)` |
| Branches | `beq rs1, rs2, label` |
| `jal` | `jal rd, label` or `jal label` (`rd` = `ra`). Write `j label` for `rd` = `x0` |
| `jalr` | `jalr rd, rs1, off` or `jalr rd, off(rs1)` |
| `lui`, `auipc` | `lui rd, imm` (20 bits) |

`lw rd, label` is not accepted.
Use `la` and then `lw rd, 0(rd)`.

## See also

- [Preprocessor](preprocessor.md)
- [Instructions](instructions.md)
- [Output formats](output-formats.md)
- [Memory map](../../console/memory-map.md)
- [goasm overview](README.md)
