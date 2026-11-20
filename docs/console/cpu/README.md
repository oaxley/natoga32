# CPU

This page describes the CPU of the console: its instruction set, how it counts cycles, and what it does when every thread sleeps.

## Overview

| Property | Value |
|----------|-------|
| Core | One 32-bit RISC-V core in machine mode |
| Clock | 25 MHz, one instruction per cycle |
| Threads | 8, cooperative, thread 0 is the main thread. See [Threads](threads.md) |
| Instruction size | 4 bytes, no compressed instructions |
| Registers | `x0` to `x31` for each thread, `x0` is always 0 |
| Reset | Thread 0 starts at `0x0000_0100`, the other threads are free |
| Interrupts | None are raised. The CPU only takes [traps](traps.md) |

The CPU runs one instruction per tick.
There is no pipeline, no cache and no wait state: a multiplication, a division and a load all cost one cycle.
Atomic and floating point instructions do not exist.

## Instructions decoded

This is the list of the instructions that `cpu.cpp` executes.
Everything else is not an instruction of the console.

| Set | Instructions |
|-----|--------------|
| RV32I upper and jumps | `lui`, `auipc`, `jal`, `jalr` |
| RV32I branches | `beq`, `bne`, `blt`, `bge`, `bltu`, `bgeu` |
| RV32I loads | `lb`, `lh`, `lw`, `lbu`, `lhu` |
| RV32I stores | `sb`, `sh`, `sw` |
| RV32I with an immediate | `addi`, `slti`, `sltiu`, `xori`, `ori`, `andi`, `slli`, `srli`, `srai` |
| RV32I with registers | `add`, `sub`, `sll`, `slt`, `sltu`, `xor`, `srl`, `sra`, `or`, `and` |
| M | `mul`, `mulh`, `mulhsu`, `mulhu`, `div`, `divu`, `rem`, `remu` |
| Zbb | `clz`, `ctz`, `cpop`, `sext.b`, `sext.h`, `zext.h`, `min`, `minu`, `max`, `maxu`, `rol`, `ror`, `rori`, `rev8` |
| Zbs | `bset`, `bclr`, `binv`, `bext`, `bseti`, `bclri`, `binvi`, `bexti` |
| Zbkb | `pack`, `packh` (`rev8` is shared with Zbb) |
| Zicond | `czero.eqz`, `czero.nez` |
| CSR | `csrrw`, `csrrs`, `csrrc`, `csrrwi`, `csrrsi`, `csrrci`. See [CSR registers](csr.md) |
| System | `ecall`, `mret`. See [Traps](traps.md) |
| Custom (opcode `0x0B`) | `new.t`, `yield.t`, `id.t`, `sleep.t`, `wake.t`, `end.t`. See [Instructions](instructions.md) |

The other instructions of Zbb, Zbs and Zbkb (`andn`, `orn`, `xnor`, `orc.b`, `brev8`, `zip`, `unzip`, ...) are not decoded.
`fence` and `fence.i` are not decoded either.

### What the CPU does with the rest

| Case | Result |
|------|--------|
| Opcode that does not exist (including `fence`) | Illegal instruction trap |
| Load with an unused `funct3`, store with an unused `funct3` | Illegal instruction trap |
| Write to a read-only CSR | Illegal instruction trap |
| Known opcode with an unassigned `funct3`, `funct7` or immediate (for example `ebreak`, `wfi`, or a branch with `funct3` 2 or 3) | Does nothing: the PC moves to the next instruction and `rd` is unchanged |

`div` and `divu` by zero return `0xFFFF_FFFF`, `rem` and `remu` by zero return the dividend.
`div` of `0x8000_0000` by -1 returns `0x8000_0000` and `rem` returns 0.
`jalr` clears bit 0 of the target.
There is no check on the alignment of an address.

Known problem: `mulhsu` multiplies `rs1` by the number of the register `rs2` instead of by its value.
Do not use it until it is fixed.

## Cycle counters

The CPU has two counters, and they can differ.

| Counter | Where | What it counts |
|---------|-------|----------------|
| Global | `CPU::getTotalCycles`, [`vc_getTotalCycles`](../libvc-api.md), CSR `cycle` | Every tick since the reset: executed instructions, instructions that trap, and idle ticks |
| Per thread | `total_cycles` in the [thread control block](threads.md#thread-control-block), CSR `instret` | Only the instructions of this thread that complete. An instruction that traps is not counted |

A program reads them with the [counter CSRs](csr.md#counters).
The value read is the counter before the reading instruction is counted.

A frame is 416,666 cycles (25 MHz / 60).
The host triggers the vblank event after each frame, see [libvc C API](../libvc-api.md#events).

## CPU states

| State | Meaning |
|-------|---------|
| Running | A thread executes instructions |
| Idle | Every thread sleeps. Nothing runs and the time goes on |
| Halted | A trap happened with no handler. The CPU stays stopped until the reset |

A host reads the state with `vc_getCpuState`.
When the CPU halts, the console goes to its error state with the text `CPU halted due to exception`.

## Idle CPU

A thread that calls `sleep.t` leaves the CPU.
If no thread is ready and at least one thread sleeps with a timer or on a hardware event, the CPU becomes Idle.
If nothing can wake a thread, it is a deadlock instead, see [Traps](traps.md#thread-traps).

While the CPU is Idle:

- every tick adds 1 to the global counter, no thread counter changes;
- a thread with a timer wakes when the global counter reaches its `sleep_until` value;
- an event (`vc_triggerEvent`, for example the vblank) wakes the threads that wait for it.

As soon as a thread is ready again, the CPU returns to Running and runs the first ready thread after the last one that ran.
The console does not run the idle ticks one by one: `vc_run` jumps to the next wake-up, or to the end of the requested cycles.
The result is the same as running every tick.

## See also

- [Threads](threads.md)
- [Instructions](instructions.md)
- [CSR registers](csr.md)
- [Traps](traps.md)
- [Memory map](../memory-map.md)
- [libvc C API](../libvc-api.md)
- [goasm instructions](../../tools/goasm/instructions.md)
