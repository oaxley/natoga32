# goasm instructions

This page lists every instruction and pseudo-instruction that goasm assembles, and the ones whose goasm behaviour differs from the CPU.

All names are lower case.
The [CPU instruction list](../../console/cpu/README.md) is the list of what the console runs.
The goasm list is the same, plus one instruction (`wfi`) that the CPU does not run.

```asm
.text
    add  a0, a1, a2
    addi a0, a1, -1
    slli a0, a1, 3
    mul  a3, a0, a1
    clz  a4, a0
    min  a5, a0, a1
    bset a6, a0, a1
    czero.eqz t0, a0, a1
    lui  t2, 0x12345
    sb   a0, 3(sp)
    beq  a0, a1, done
    jal  ra, done
done:
    ecall
```

## Base integer (RV32I)

| Group | Instructions |
|-------|--------------|
| Register-register | `add sub sll slt sltu xor srl sra or and` |
| Register-immediate | `addi slti sltiu xori ori andi slli srli srai` |
| Loads | `lb lh lw lbu lhu` |
| Stores | `sb sh sw` |
| Branches | `beq bne blt bge bltu bgeu` |
| Jumps | `jal jalr` |
| Upper immediate | `lui auipc` |
| System | `ecall` |

`fence`, `fence.i` and `ebreak` are not accepted.
Operand forms are in [Syntax](syntax.md#operand-forms).
`lui` and `auipc` take the 20-bit value that goes in the upper bits: `lui t2, 0x12345` loads `0x1234_5000`.

## M extension

`mul mulh mulhsu mulhu div divu rem remu`.

The CPU has a known bug in `mulhsu` (see [CPU](../../console/cpu/README.md)).

## Bit-manipulation extensions

| Extension | Instructions |
|-----------|--------------|
| Zbb | `clz ctz cpop sext.b sext.h zext.h min minu max maxu rol ror rori rev8` |
| Zbs | `bset bclr binv bext bseti bclri binvi bexti` |
| Zbkb | `pack packh` |
| Zicond | `czero.eqz czero.nez` |

Operand forms:

- `clz rd, rs1`, same for `ctz cpop sext.b sext.h zext.h rev8`.
- `min rd, rs1, rs2`, same for the other three-register instructions.
- `zext.h` shares its encoding with `pack`. `pack rd, rs1, x0` is `zext.h rd, rs1`.
- `rori rd, rs1, shamt`, same for `bseti bclri binvi bexti`. `shamt` is a number from 0 to 31, like for `slli srli srai`.
  A value outside 0..31 stops with `shift amount N out of range (0..31)`.
  A register in its place stops with `shift amount must be a number (0..31), not register R`.

## System and CSR

| Instruction | Form | Note |
|-------------|------|------|
| `ecall` | `ecall` | Environment call. |
| `mret` | `mret` | Return from a trap handler (`0x3020_0073`). |
| `wfi` | `wfi` | Assembles to `0x1050_0073`. **The CPU does not run it**: the word does nothing. |
| `csrrw` `csrrs` `csrrc` | `csrrw rd, rs1, csr` | Operand order is `rd, rs1, csr`. |
| `csrrwi` `csrrsi` `csrrci` | `csrrwi rd, uimm, csr` | `uimm` is 0 to 31. |

- The standard RISC-V order is `rd, csr, rs1`. goasm puts the CSR **last**. The pseudo-instructions below hide this.
- `csr` is a number from 0 to `0xFFF`. CSR names such as `mstatus` are not known: use a number or an `.equ` constant. The numbers are in [CSR](../../console/cpu/csr.md).
- An immediate out of range stops with `csrrwi immediate 40 out of range (0..31)`.
- `ecall`, `mret` and `wfi` take no operand. `mret a0` stops with `mret takes no operands`.

The disassembler shows the `uimm` of `csrrwi` as a register name (`x5`).

```asm
.equ MSTATUS, 0x300
.text
    csrr   a0, MSTATUS      ; csrrs a0, x0, MSTATUS
    csrw   MSTATUS, a0      ; csrrw x0, a0, MSTATUS
    csrs   MSTATUS, a1
    csrc   MSTATUS, a1
    csrrw  a1, a0, 0x305
    csrwi  MSTATUS, 8
    csrrwi a2, 5, 0x340
    mret
    ecall
```

## Thread instructions

`new.t yield.t id.t sleep.t wake.t end.t`

They use the custom opcode `0x0B`.
Syntax, encoding and behaviour are in [Thread instructions](../../console/cpu/instructions.md).

```asm
.text
main:
    la      a0, worker
    new.t   a1, a0
    id.t    a2
    yield.t
    sleep.t a0, a1
    wake.t  a1
    end.t
worker:
    end.t
```

## Pseudo-instructions

goasm expands these in phase 4, before the labels are resolved.
`--disasm --pseudo` folds some of them back.

| Pseudo | Expansion |
|--------|-----------|
| `nop` | `addi x0, x0, 0` |
| `mv rd, rs` | `addi rd, rs, 0` |
| `li rd, imm` | `addi rd, x0, imm` for -2048 to 2047, else `lui` + `addi` |
| `la rd, sym` | `auipc rd, %pcrel_hi(sym)` + `addi rd, rd, %pcrel_lo(sym)` |
| `call sym` | `auipc x1, %pcrel_hi(sym)` + `jalr x1, x1, %pcrel_lo(sym)` |
| `ret` | `jalr x0, x1, 0` |
| `j label` | `jal x0, label` |
| `jr rs` | `jalr x0, rs, 0` |
| `not rd, rs` | `xori rd, rs, -1` |
| `neg rd, rs` | `sub rd, x0, rs` |
| `seqz rd, rs` | `sltiu rd, rs, 1` |
| `snez rd, rs` | `sltu rd, x0, rs` |
| `sltz rd, rs` | `slt rd, rs, x0` |
| `sgtz rd, rs` | `slt rd, x0, rs` |
| `beqz rs, l` | `beq rs, x0, l` |
| `bnez rs, l` | `bne rs, x0, l` |
| `blez rs, l` | `bge x0, rs, l` |
| `bgez rs, l` | `bge rs, x0, l` |
| `bltz rs, l` | `blt rs, x0, l` |
| `bgtz rs, l` | `blt x0, rs, l` |
| `bgt rs, rt, l` | `blt rt, rs, l` |
| `ble rs, rt, l` | `bge rt, rs, l` |
| `bgtu rs, rt, l` | `bltu rt, rs, l` |
| `bleu rs, rt, l` | `bgeu rt, rs, l` |
| `csrr rd, csr` | `csrrs rd, x0, csr` |
| `csrw csr, rs` | `csrrw x0, rs, csr` |
| `csrs csr, rs` | `csrrs x0, rs, csr` |
| `csrc csr, rs` | `csrrc x0, rs, csr` |
| `csrwi csr, imm` | `csrrwi x0, imm, csr` |
| `csrsi csr, imm` | `csrrsi x0, imm, csr` |
| `csrci csr, imm` | `csrrci x0, imm, csr` |
| `syscall N` | `addi x17, x0, N` + `ecall` (`N` a number) |
| `syscall rs` | `add x17, x0, rs` + `ecall` |

Notes:

- `li` with a number outside -2048 to 2047, or with an expression, uses `lui` with `%hi` and `addi` with `%lo`. `li a0, 100000` gives `lui a0, 0x18` and `addi a0, a0, 1696` (100000 is `0x186A0`; `%lo` is `0x6A0`).
- `li rd, label` loads the absolute address of the label (`lui` + `addi`). `la` is the PC-relative form.
- `la` and `call` are relative to the PC.
- `syscall` always needs one operand. The console defines no syscall numbers.
- `jal label` with a single operand does not jump to `label`. Write `j label` or `jal ra, label`.

```asm
.text
start:
    li   a0, 100000
    la   a1, data
    mv   a2, a0
    not  a3, a2
    neg  a4, a2
    seqz a5, a0
    beqz a0, done
    bgt  a0, a1, done
    call func
    j    done
func:
    ret
done:
    nop
.data
data:
    .word 0
```

## goasm and the CPU

| Case | goasm | CPU |
|------|-------|-----|
| `wfi` | Assembled (`0x1050_0073`). | Not decoded: the word does nothing. |
| `fence`, `fence.i`, `ebreak` | Not accepted. | Not decoded (`fence`: illegal instruction). |

Every other instruction of the CPU list is accepted by goasm.

## See also

- [Syntax](syntax.md)
- [Quick reference](quick-reference.md)
- [CPU instruction list](../../console/cpu/README.md)
- [Thread instructions](../../console/cpu/instructions.md)
- [CSR](../../console/cpu/csr.md)
- [Traps](../../console/cpu/traps.md)
