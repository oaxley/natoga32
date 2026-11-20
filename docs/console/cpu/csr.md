# CSR registers

This page lists the control and status registers (CSRs) of the CPU, how a program reads and writes them, and which ones do something.

## Overview

A CSR has a number from 0 to 4095 (12 bits).
A program reads and writes the CSRs with the six CSR instructions: `csrrw`, `csrrs`, `csrrc`, `csrrwi`, `csrrsi`, `csrrci`.
goasm also knows the short forms `csrr`, `csrw`, `csrs`, `csrc`, `csrwi`, `csrsi` and `csrci`.
It does not know the CSR names: write the number, for example `csrr t0, 0xC00`, or define it with `.equ`.

## Shared CSR memory

The CSR values are stored in the CSR zone of the [memory map](../memory-map.md).
The value of the CSR `n` is the word at `0x0020_1000 + n * 4`, and the zone covers the 4096 numbers (16 KiB).

- The CSRs are shared by the 8 threads, except `cartbank`.
  A thread that sets `mtvec` sets it for all of them. `mepc`, `mcause` and `mtval` are shared too.
- The reset clears all the CSRs to 0.
- The zone is plain memory: a `lw` or a `sw` works on it, and the host can read it with `vc_memRead*`.
  Use the CSR instructions in a program.
  The counters and `cartbank` are not kept in this memory, so a `lw` does not return their value.
- The CSRs that are listed below as "storage only" hold what a program writes, and nothing else uses it.

## Access rules

- A CSR from `0xC00` to `0xFFF` (the two top bits are 1) is read-only.
- A CSR instruction that writes a read-only CSR raises an illegal instruction trap (`mcause` 2).
  Nothing is written, and `rd` is unchanged. See [Traps](traps.md).
- `csrrw` and `csrrwi` always write.
  `csrrs`, `csrrc`, `csrrsi` and `csrrci` write only when `rs1` (or the 5-bit immediate) is not 0.
  So `csrr` (a `csrrs` with `rs1` = `zero`) can read a read-only CSR.
- `rd` receives the old value of the CSR.
- The 5-bit immediate of `csrrwi`, `csrrsi` and `csrrci` is in the `rs1` field.

## List of the CSRs

| Number | Name | Access | Notes |
|--------|------|--------|-------|
| `0x300` | `mstatus` | Read/write | Bits 3 (MIE) and 7 (MPIE) are changed by `mret`. See [Traps](traps.md#mret) |
| `0x304` | `mie` | Read/write | Storage only |
| `0x305` | `mtvec` | Read/write | Address of the trap handler. 0 = no handler. See [Traps](traps.md#trap-entry) |
| `0x340` | `mscratch` | Read/write | Storage only. A trap handler can use it to save a register |
| `0x341` | `mepc` | Read/write | Address of the instruction that trapped. `mret` resumes at `mepc + 4` |
| `0x342` | `mcause` | Read/write | Cause of the last trap |
| `0x343` | `mtval` | Read/write | Value that goes with the cause (an address, or 0) |
| `0x344` | `mip` | Read/write | Storage only |
| `0x7C0` | `cartbank` | Read/write | Per thread, 16 bits. See [cartbank](#cartbank) |
| `0xC00` | `cycle` | Read-only | Global cycle counter, low 32 bits |
| `0xC01` | `time` | Read-only | Low 32 bits of the virtual time. Always 0 today |
| `0xC02` | `instret` | Read-only | Instructions completed by the current thread, low 32 bits |
| `0xC80` | `cycleh` | Read-only | Global cycle counter, high 32 bits |
| `0xC81` | `timeh` | Read-only | High 32 bits of the virtual time. Always 0 today |
| `0xC82` | `instreth` | Read-only | Instructions completed by the current thread, high 32 bits |

Any other number from 0 to `0xBFF` is a free storage word: it reads what was written, and nothing uses it.
Any other number from `0xC00` to `0xFFF` is read-only and reads 0.

No interrupt is raised today.
`mie`, `mip` and the interrupt bits of `mstatus` do not change what the CPU does.

## Counters

The six counters are read-only and computed when they are read.

| Pair | Counter | Value |
|------|---------|-------|
| `cycle` / `cycleh` | Global | Every tick since the reset, idle ticks and trapped instructions included |
| `instret` / `instreth` | Per thread | Instructions completed by the thread that runs the read. A thread has its own value |
| `time` / `timeh` | Virtual time | Not driven by the console: 0 |

Read the 64-bit value with the usual sequence: read `cycleh`, `cycle`, then `cycleh` again, and start over if the high word changed.
The value read is the counter before the reading instruction is counted.
See [CPU](README.md#cycle-counters).

```asm
.equ CSR_CYCLE,   0xC00
.equ CSR_INSTRET, 0xC02

    csrr t0, CSR_CYCLE      # global cycle counter, low 32 bits
    csrr t1, CSR_INSTRET    # instructions completed by this thread
```

## cartbank

`cartbank` (`0x7C0`) is the bank register of the cartridge window.
It is the only per-thread CSR: each thread has its own value, stored in its [thread control block](threads.md#thread-control-block) at `+0x0C`.

- It is 16 bits. A write keeps the low 16 bits (the value is masked with `0xFFFF`).
- It is 0 after the reset and for a new thread.
- A write does not change the value of another thread.
- The byte at offset `o` of the window `0x0036_0000` is the byte `(cartbank << 12) + o` of the cartridge file.
  It is 0 beyond the end of the file, without a valid cartridge, and with no cartridge.

```asm
.equ CSR_CARTBANK, 0x7C0

    li   t2, 3
    csrw CSR_CARTBANK, t2   # the window now starts at byte 3 * 4096 of the file
    li   t3, 0x00360000     # cartridge window
    lw   a0, 0(t3)          # first word of the bank
```

The window and the control registers are described in [Cartridge: console view](../cartridge/console-view.md).

## See also

- [CPU](README.md)
- [Threads](threads.md)
- [Traps](traps.md)
- [Memory map](../memory-map.md)
- [Cartridge: console view](../cartridge/console-view.md)
