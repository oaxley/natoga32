# goasm quick reference

This page is a one-screen summary of goasm: command line, directives, instructions and common patterns.
Details are in the other goasm pages.

## Command line

```bash
goasm in.s -o out.rscv               # RSCV chunk (default)
goasm in.s -o out.rscv --sym out.sym # plus the dSYM chunk
goasm boot.s -f rom -o rombios.sys   # boot ROM image
goasm --disasm out.rscv --pseudo     # disassemble
goasm --disasm rombios.sys --base 0  # disassemble a ROM image
goasm in.s -d 3                      # stop after phase 3 (AST)
```

Options: `-o`, `-f rscv|rom`, `--sym`, `-O`, `-d 0..7`, `--disasm`, `--base`, `--pseudo`, `--hex`, `--no-abi`.
See [the overview](README.md).

## Sections

| Directive | Base | Size |
|-----------|------|------|
| `.rom` | `0x0000_0000` | 2 MiB |
| `.bss` | `0x0100_0000` | 2 MiB |
| `.rodata` | `0x0120_0000` | 2 MiB |
| `.data` | `0x0140_0000` | 6 MiB |
| `.text` (default) | `0x01A0_0000` | 6 MiB |

## Directives

| Directive | Meaning |
|-----------|---------|
| `.byte` `.half` `.word` | Values of 1, 2, 4 bytes. |
| `.string` `.asciz` / `.ascii` | Text with / without a 0 byte. Escapes `\n \t \r \0 \\ \' \" \xHH`. |
| `.space N` / `.align N` / `.org ADDR` | Zeros; to a multiple of `2^N`; up to an absolute address. |
| `.equ NAME, VALUE` | Constant (any expression, also of names defined below). |
| `.global NAME` | Before or after the label. |
| `.incbin "f"` / `.include "f"` | Bytes / source, relative to the input file. |
| `.define NAME V` | Constant (`.define NAME, V` also works). C precedence, `( ) ! ~ && \|\|`. |
| `.if` `.ifdef` `.ifndef` `.else` `.endif` | Conditions. |
| `.for i = a, b[, s]` ... `.endf` | Loop, `b` excluded. |
| `.macro n p` ... `.endm` | Macro, `\p` in the body. |
| `.dup COUNT VALUE` | Repeat a value. |
| `a##b` / `$[VAR]` | Paste / environment variable. |

## Registers

| Number | ABI |
|--------|-----|
| `x0` `x1` `x2` `x3` `x4` | `zero` `ra` `sp` `gp` `tp` |
| `x5`-`x7` | `t0`-`t2` |
| `x8` `x9` | `s0` (`fp`) `s1` |
| `x10`-`x17` | `a0`-`a7` |
| `x18`-`x27` | `s2`-`s11` |
| `x28`-`x31` | `t3`-`t6` |

## Instructions

| Group | Instructions |
|-------|--------------|
| Arithmetic | `add sub addi` |
| Logic | `and or xor andi ori xori` |
| Shifts | `sll srl sra slli srli srai` |
| Compare | `slt sltu slti sltiu` |
| Loads / stores | `lb lh lw lbu lhu` / `sb sh sw` |
| Branches | `beq bne blt bge bltu bgeu` |
| Jumps | `jal jalr` (`jalr rs` is `jalr ra, rs, 0`; `jalr rd, rs` is `jalr rd, rs, 0`) |
| Upper | `lui auipc` |
| M | `mul mulh mulhsu mulhu div divu rem remu` |
| Zba | `sh1add sh2add sh3add` |
| Zbb | `clz ctz cpop sext.b sext.h zext.h orc.b min minu max maxu rol ror rori rev8 andn orn xnor` |
| Zbs | `bset bclr binv bext bseti bclri binvi bexti` |
| Zbkb | `pack packh` |
| Zicond | `czero.eqz czero.nez` |
| CSR | `csrrw csrrs csrrc csrrwi csrrsi csrrci` (`csrrw rd, csr, rs1`, standard order) |
| Memory ordering | `fence fence.i` (`fence` alone is `fence iorw, iorw`; both do nothing on the CPU) |
| System | `ecall ebreak mret wfi` (`ebreak` and `wfi` do nothing on the CPU) |
| Threads | `new.t yield.t id.t sleep.t wake.t end.t` |

## Pseudo-instructions

`nop mv li la call tail ret j jr not neg seqz snez sltz sgtz beqz bnez blez bgez bltz bgtz bgt ble bgtu bleu csrr csrw csrs csrc csrwi csrsi csrci rdcycle[h] rdtime[h] rdinstret[h] unimp syscall [N|rs]`.
Expansions are in [Instructions](instructions.md#pseudo-instructions).

## Patterns

```asm
.text
func:                       ; prologue and epilogue
    addi sp, sp, -16
    sw   ra, 12(sp)
    sw   s0, 8(sp)
    lw   s0, 8(sp)
    lw   ra, 12(sp)
    addi sp, sp, 16
    ret
cmov:                       ; if (a0 != 0) a1 = a2, with Zicond
    czero.nez t0, a1, a0
    czero.eqz t1, a2, a0
    or   a1, t0, t1
loop:                       ; count to 10
    li   t0, 0
    li   t1, 10
again:
    addi t0, t0, 1
    blt  t0, t1, again
    slli t2, a1, 2          ; a3 = array[i]: a0 = array, a1 = i
    add  t2, a0, t2
    lw   a3, 0(t2)
    lui  t3, 0x12345        ; load 0x1234_5678
    addi t3, t3, 0x678
    call func
```

## See also

- [goasm overview](README.md)
- [Syntax](syntax.md)
- [Preprocessor](preprocessor.md)
- [Instructions](instructions.md)
- [Output formats](output-formats.md)
