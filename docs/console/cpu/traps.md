# Traps

This page describes what the CPU does when an instruction cannot complete: how a trap is taken, which causes exist, and how `ecall` and `mret` work.

## Overview

A trap stops the current instruction and jumps to a handler written in the program.
The CPU raises traps for errors and for `ecall`.
It raises no interrupt: the interrupt bit of `mcause` is never set today.

A trap runs in the thread that trapped.
It does not switch thread, and the handler sees and changes the registers of that thread.

## Trap entry

When an instruction traps, the CPU does these steps:

1. `mepc` receives the address of the instruction that trapped.
2. `mcause` receives the cause code (see [Causes](#causes)).
3. `mtval` receives a value that goes with the cause: the faulting address for an access fault, 0 for the other causes.
4. The PC becomes the value of `mtvec`.

The registers are not saved and `mstatus` is not changed.
The handler must save what it uses.

The instruction that trapped has no other effect:

- the destination register is unchanged and memory is not written;
- the PC does not move to the next instruction;
- the instruction is not counted in `instret`. See [CPU](README.md#cycle-counters).

If a second trap happens in the handler, it overwrites `mepc`, `mcause` and `mtval`.
The CSRs are described in [CSR registers](csr.md).

## Trap without a handler

`mtvec` is 0 at reset.
A trap taken while `mtvec` is 0 halts the CPU:

- `mepc`, `mcause` and `mtval` are set as usual;
- the PC stays on the instruction that trapped;
- the CPU stays Halted until the next reset;
- the console goes to its error state with the text `CPU halted due to exception`.

For this reason, a program sets `mtvec` first, with `csrw`.
The address 0 cannot be a handler.

## Causes

| `mcause` | Name | Raised by | `mtval` |
|----------|------|-----------|---------|
| 1 | Instruction access fault | A fetch where the 4 bytes at the PC touch the cartridge control zone or the cartridge window | The PC |
| 2 | Illegal instruction | An opcode that does not exist, an unassigned encoding of a known opcode, a write to a read-only CSR | 0 |
| 5 | Load access fault | A load that touches the private thread zone, or that crosses the edge of a cartridge zone | The address of the load |
| 7 | Store access fault | A store that touches the private thread zone or a cartridge zone | The address of the store |
| 11 | Environment call from M-mode | `ecall` | 0 |
| 24 | Main thread exit | `end.t` in thread 0 | 0 |
| 25 | Thread spawn error | `new.t` with no Free slot | 0 |
| 26 | Stack overflow | `yield.t` when the stack canary has changed | 0 |
| 27 | Deadlock | `sleep.t` or `end.t` when no thread is Ready and none can wake up | 0 |

The values 1, 2, 5, 7 and 11 are the RISC-V ones.
The values 24 to 27 are specific to the console.

### Access faults

The rules for the zones are in the [memory map](../memory-map.md).
For a program:

| Access | Result |
|--------|--------|
| Load or store that touches one byte of the private thread zone `0x0020_5000` to `0x0020_5FFF` | Load access fault (5) or store access fault (7) |
| Load fully inside the cartridge control zone or the cartridge window | Works. The window is read through `cartbank` |
| Load that starts outside a cartridge zone and ends inside it, or the reverse | Load access fault (5) |
| Store that touches the cartridge control zone or the cartridge window | Store access fault (7) |
| Fetch that touches the cartridge control zone or the cartridge window | Instruction access fault (1) |

"Touches" includes an access that overlaps the zone by one byte: a word at `0x0020_4FFE` faults.
The check uses the size of the access (1, 2 or 4 bytes).
A fetch from the private thread zone is not checked.
Outside these rules, there is no access fault: an address beyond the 32 MiB reads 0 and a write is ignored.

### Illegal instruction

The CPU raises an illegal instruction trap (2) for:

- an opcode that is not in the [instruction list](README.md#instructions-decoded);
- opcode `0x0F` with a `funct3` other than 0 (`fence`) and 1 (`fence.i`);
- a load with `funct3` 3, 6 or 7, and a store with `funct3` above 2. The CPU checks this before any access;
- a CSR instruction that writes a read-only CSR, see [CSR registers](csr.md#access-rules);
- a known opcode with an unassigned `funct3`, `funct7` or `rs2` field, see [CPU](README.md#what-the-cpu-does-with-the-rest).

`ebreak` and `wfi` do not trap: they do nothing for now.

### Thread traps

| `mcause` | When | State after the trap |
|----------|------|----------------------|
| 24 | Thread 0 calls `end.t` | Thread 0 keeps running |
| 25 | `new.t` finds no Free slot | `rd` is unchanged |
| 26 | `yield.t` finds the canary changed | The thread is not switched, it stays Running |
| 27 | After `sleep.t` or `end.t`, no thread is Ready and no thread waits for a hardware event or a timer | The trap runs in the thread that tried to switch. Its state is already Sleeping or Dead |

For 24 to 27, `mepc` is the address of the thread instruction.
The instructions are in [Instructions](instructions.md), the stacks in [Threads](threads.md).

## ecall

`ecall` raises the trap 11 and the PC jumps to `mtvec`.
The CPU does not look at the registers: the handler does, in software.

`mepc` is the address of the `ecall` itself.
The handler ends with `mret`, which resumes at `mepc + 4`, the instruction after the `ecall`.
The handler does not add 4 to `mepc` (see [mret](#mret)).

### Calling convention

The call follows the standard RISC-V convention.
The CPU does not check it, the program and its handler agree on it.

| Register | Role |
|----------|------|
| `a7` | Call number |
| `a0` to `a6` | Arguments |
| `a0` | Return value (`a1` for a second value, or the high half of a 64-bit value) |

A call behaves like a function call:

- the handler may overwrite `t0` to `t6`, so the caller saves them if it needs them;
- `a0` (and `a1` when used) receive the result;
- every other register is preserved, including `ra`, `sp`, `gp`, `tp` and `s0` to `s11`.

Syscall numbers are not defined yet.
The [boot ROM](../../rom/README.md) has no syscall handler: its trap handler shows an error screen and the value of `mcause`, for every trap.

## mret

`mret` (`0x3020_0073`) ends a handler.

- The PC becomes `mepc + 4`.
- In `mstatus`, bit 3 (MIE) receives bit 7 (MPIE), then bit 7 (MPIE) is set to 1.

The `+ 4` is part of the instruction, whatever the cause of the trap.
It is what a handler for `ecall` needs: the program resumes after the `ecall`.
After another trap (a fault, an illegal instruction), `mret` also skips the instruction that trapped.
A handler that wants to run it again, or to resume somewhere else, writes `mepc` before `mret`: to resume at the address `x`, it writes `x - 4`.

The trap entry does not change `mstatus`, so its bits only follow what the program wrote.

## Example: a handler for ecall

```asm
.equ CSR_MTVEC, 0x305

.text
.global main
main:
    la   t0, handler
    csrw CSR_MTVEC, t0
    li   a7, 1              # call number, chosen by the program
    li   a0, 5              # argument
    ecall
    mv   s1, a0             # s1 = 6, the handler added 1
done:
    j    done

handler:
    addi a0, a0, 1          # result in a0
    mret                    # resumes after the ecall
```

## See also

- [CPU](README.md)
- [Threads](threads.md)
- [Instructions](instructions.md)
- [CSR registers](csr.md)
- [Memory map](../memory-map.md)
- [Boot ROM](../../rom/README.md)
