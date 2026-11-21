# goasm

This page says what goasm is, how to build it, and what each command-line option does.

goasm is an assembler and disassembler for the NATOGA32 console, written in Go.
It reads RISC-V assembly (RV32I, M, some bit-manipulation extensions and the [thread instructions](../../console/cpu/instructions.md)) and writes a cartridge chunk, a raw binary or a boot ROM image.
It has a preprocessor with includes, constants, conditions, loops and macros.

## Build

```bash
cd tools/goasm
go build -o goasm ./cmd/goasm/
```

Run the tests with `go test ./...` in the same directory.

## Usage

```bash
# Cartridge code chunk (default format)
goasm game.s -o game.rscv

# Same, plus the symbol chunk
goasm game.s -o game.rscv --sym game.sym

# Boot ROM image
goasm boot.s -f rom -o rombios.sys

# Raw binary of the .text section
goasm game.s -f bin -o game.bin

# Disassemble a file
goasm --disasm game.rscv
```

Exactly one input file is required.
Options can come before or after it.

## Options

| Option | Default | Meaning |
|--------|---------|---------|
| `-o`, `--output FILE` | `a.out` | Output file. |
| `-f`, `--format FMT` | `rscv` | `rscv`, `bin` or `rom`. See [Output formats](output-formats.md). |
| `--sym FILE` | none | Also write the labels as a `dSYM` chunk. Needs `-f rscv`. |
| `-O`, `--optimize` | off | Peephole optimizer. See [Output formats](output-formats.md#optimizer). |
| `-d`, `--debug N` | `0` | Print the result of a phase. See below. |
| `--disasm` | off | Disassemble the input file instead of assembling. |
| `--base ADDR` | `27262976` (`0x01A0_0000`) | Load address of a raw binary. Used by `--disasm` only. |
| `--pseudo` | off | `--disasm` only: print pseudo-instructions (`mv`, `ret`, `j`, ...). |
| `--hex` | on | `--disasm` only: show the encoded word. `--hex=false` hides it. |
| `--no-abi` | off | `--disasm` only: print `x0`-`x31` instead of `zero`, `ra`, `a0`, ... |

An unknown format stops with `unknown output format: X`, and no file is written.
`--sym` with another format stops with `--sym needs -f rscv`.

## Phases and debug levels

goasm runs these phases in order.
Each phase prints a `Phase N` line on stdout.

| Level | Phase | What `-d` prints |
|-------|-------|------------------|
| 1 | Lexer | Tokens. |
| 2 | Preprocessor | Tokens after includes, macros, loops and conditions. |
| 3 | Parser | AST. |
| 4 | Pseudo expansion | AST with the pseudo-instructions expanded. |
| 5 | Pass 1 | Symbols and sections. |
| 6 | Pass 2 | Encoded sections as a hex dump. |
| 7 | Output | What was written: format, sections, file size, `dSYM` label count. |

With `-d 1` to `-d 6`, goasm stops after the printed phase and writes no file.
With `-d 7`, goasm writes the files as usual, then prints the summary.
`-d 0` (the default) prints only the `Phase N` lines.
Another value stops with `--debug must be 0 to 7, got 8` before anything is read or written.
The optimizer runs between phase 4 and phase 5, only with `-O`.

## Disassembly

`goasm --disasm FILE` reads a file and prints assembly on stdout.
Use `-o FILE` to write it to a file instead.

- If bytes 4 to 7 of the file are `RSCV`, the file is read as an [RSCV chunk](../../console/cartridge/chunks.md#rscv).
  A damaged chunk stops with `invalid RSCV chunk: ...`.
- Otherwise the file is a raw binary, placed at `--base`.
- An empty file stops with `file is empty`.

The output starts with `; Disassembly of FILE (rscv format)` (or `bin`), then one line per section, then the listing.
Text sections show address, word, mnemonic and operands.
Data sections show a hex dump with ASCII.
A `.bss` section shows as `.space N`.
A word that is not an instruction shows as `.word 0x...`.

## Pages

| Page | Content |
|------|---------|
| [Syntax](syntax.md) | Comments, labels, numbers, expressions, registers, sections, data directives. |
| [Preprocessor](preprocessor.md) | `.include`, `.define`, conditions, `.for`, `.dup`, macros, `##`, environment variables. |
| [Instructions](instructions.md) | Every instruction and pseudo-instruction. |
| [Output formats](output-formats.md) | `rscv`, `rom`, `bin`, `--sym` and `-O`. |
| [Quick reference](quick-reference.md) | One-page tables. |

## See also

- [Syntax](syntax.md)
- [Output formats](output-formats.md)
- [Cartridge chunks](../../console/cartridge/chunks.md)
- [n32pack](../n32pack/README.md)
