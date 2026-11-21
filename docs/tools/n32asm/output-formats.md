# n32asm output formats

This page describes the two output formats of n32asm (`rscv`, `rom`), the symbol file written by `--sym`, and the optimizer.

The format is chosen with `-f`.
The default is `rscv`.

| Format | Content | Typical use |
|--------|---------|-------------|
| `rscv` | A complete `RSCV` chunk: `.text`, `.data`, `.rodata`, `.bss`. | Game code, packed by [n32pack](../n32pack/README.md). |
| `rom` | The bytes of the `.rom` section, from address `0x0000_0000`. | The boot ROM. |

Another format, such as the old `bin`, stops with `unknown output format: bin (use rscv or rom)` before anything is read.
All checks run before the output file is created.
When a check fails, n32asm writes no file.

## rscv

`rscv` is the default.
The file is the whole chunk, ready for n32pack: length, type `RSCV`, sections, padding and CRC32.
The byte layout is in [Cartridge chunks](../../console/cartridge/chunks.md#rscv).

```bash
n32asm game.s -o game.rscv
```

n32asm writes a section only when it is not empty, in this order: `.text` (kind 1), `.data` (2), `.rodata` (3), `.bss` (4).
Each section has a 12-byte header with its load address and size.
A `.bss` section has a header and no bytes.
There is no relocation: the addresses are the fixed [section windows](syntax.md#sections).

| Error | Cause |
|-------|-------|
| `no .text section: a cartridge needs code` | The program has no code. |
| `.rom section can't go in a cartridge: use -f rom` | A `.rom` section holds data. |
| `.bss section holds non-zero data at 0xADDR: use .data` | `.bss` holds a non-zero value. `.space` is allowed. |
| `section .data is N bytes, exceeds its 6MiB window (...)` | A section is bigger than its window. |

The entry point is not in the chunk.
It is set in the manifest of n32pack, as an address or a label.

## rom

```bash
n32asm boot.s -f rom -o rombios.sys
```

The file holds the `.rom` section from address `0x0000_0000`.
Gaps made by `.org` are filled with zeros.
The boot ROM of the console is built this way (`rom/CMakeLists.txt`).

```asm
.rom
    j    boot
.org 0x100
boot:
    li   a0, 1
    ret
```

This program gives a file of 264 bytes: `j boot` at offset 0, zeros up to offset `0x100`, then the code of `boot`.

| Error | Cause |
|-------|-------|
| `no .rom section: -f rom needs code or data in a .rom section` | The program has no `.rom` bytes. |
| `ROM image exceeds 2 MiB (N bytes)` | The image is bigger than the ROM window. |

Other sections of the program are ignored by `-f rom`.

A ROM image has no header, so `n32asm --disasm` reads it as a raw file. Give `--base 0`, the ROM address:

```bash
n32asm --disasm rombios.sys --base 0
```

## Symbol file (--sym)

```bash
n32asm game.s -o game.rscv --sym game.sym
```

`--sym` also writes the labels of the program as a `dSYM` chunk.
It needs `-f rscv`.
The byte layout is in [Cartridge chunks](../../console/cartridge/chunks.md#dsym).

- Only the labels of `.text`, `.data`, `.rodata` and `.bss` are written. `.equ` constants and labels of `.rom` are not.
- Labels are sorted by address, then by name.
- The global bit is set by `.global NAME`, before or after the label.
- A name over 65535 bytes stops with `dSYM names are at most 65535 bytes`, and neither file is written.

For this program:

```asm
.text
main:
    j loop
loop:
    j loop
.global main
.data
value:
    .word 1
.bss
buffer:
    .space 64
```

the entries are:

| Address | Kind | Global | Name |
|---------|------|--------|------|
| `0x0100_0000` | 4 | no | `buffer` |
| `0x0140_0000` | 2 | no | `value` |
| `0x01A0_0000` | 1 | yes | `main` |
| `0x01A0_0004` | 1 | no | `loop` |

n32pack reads this file with the `dsym` field of its [manifest](../n32pack/manifest.md).

## Optimizer

`-O` turns on a peephole optimizer.
It runs after the pseudo-instructions are expanded, on runs of instructions between labels and directives.
It makes one sweep over each run and has two rules:

| Rule | Replaces | By |
|------|----------|----|
| Zero operand | `add rd, rs, zero`, `add rd, zero, rs`, `or rd, rs, zero`, `or rd, zero, rs` | `addi rd, rs, 0` |
| Redundant move | `addi x, y, 0` then `addi y, x, 0` (or `addi x, x, 0`) | `addi x, y, 0` |

Each rule changes only the instructions it replaces: the instructions around them stay.
The registers hold the same values after the optimized code as after the original code.

The redundant-move rule only removes a second move that has no effect.
A chain `mv t0, t1` / `mv t2, t0` is kept as it is: the assembler does not know whether a later instruction reads `t0`.
A move through `x0` is also kept (`mv x0, t1` / `mv t1, x0` writes 0 into `t1`).

```asm
.text
    add  t0, t1, zero       ; becomes addi t0, t1, 0
    li   a0, 1
    mv   t2, t0
    mv   t0, t2             ; removed: t0 already holds t2
```

## See also

- [n32asm overview](README.md)
- [Syntax](syntax.md)
- [Cartridge chunks](../../console/cartridge/chunks.md)
- [Memory map](../../console/memory-map.md)
- [n32pack](../n32pack/README.md)
