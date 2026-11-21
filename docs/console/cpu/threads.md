# Threads

This page describes the 8 threads of the CPU: their states, their control blocks, their stacks and how the CPU moves from one to the next.

## Overview

- The console has one core and 8 threads, numbered 0 to 7.
- Thread 0 is the main thread. It starts at `0x0000_0100` at reset and it cannot end.
- The threads are cooperative. The running thread keeps the CPU until it calls `yield.t`, `sleep.t` or `end.t`.
- There is no scheduler. The next thread is the first ready one after the current thread, in the order 0, 1, ..., 7, 0, ... (round robin).
- Each thread has its own 32 registers, stack, TLS block and control block.

At reset, thread 0 is Running and the threads 1 to 7 are Free.
A thread is created with `new.t` and it uses the first Free slot.
The instructions are described in [Instructions](instructions.md).

## States

The state of a thread is one byte in its control block.
The values are bits, a thread has one state at a time.

| Value | State | Meaning |
|-------|-------|---------|
| 0 | Free | The slot is not in use. `new.t` can take it |
| 1 | Ready | The thread can run. It waits for its turn |
| 2 | Running | The thread runs now. Only one thread is Running |
| 4 | Sleeping | The thread waits for an event, a timer, or both |
| 8 | Dead | The thread has ended |

A Dead slot is not reused: only a Free slot can be taken by `new.t`.
A program can start at most 7 threads between two resets, even if they all ended.

| From | To | Cause |
|------|----|-------|
| Free | Ready | `new.t` takes the slot |
| Ready | Running | The CPU gives the thread its turn |
| Running | Ready | `yield.t` |
| Running | Sleeping | `sleep.t` |
| Running | Dead | `end.t` |
| Sleeping | Ready | `wake.t`, an event, or the end of the timer |

When the CPU is Idle, no thread is Running, see [CPU](README.md#idle-cpu).

## Registers

Each thread has its own 32 registers.
They are saved in the [private thread zone](../memory-map.md#private-thread-zone): the register `n` of the thread `tid` is at `0x0020_5000 + tid * 128 + n * 4`.
A program cannot read this zone, it only sees its own registers.

A new thread starts with all its registers at 0, except:

| Register | Value |
|----------|-------|
| `sp` (`x2`) | The top of the stack of the thread (see [Stacks](#stacks)) |
| `tp` (`x4`) | The start of the TLS block of the thread (see [TLS](#tls)) |

`new.t` does not pass an argument. A new thread gets its work from memory.

## Thread control block

The state of a thread is kept in a control block (TCB) of 32 bytes, at `0x0020_5400 + tid * 32`.
The TCBs are in the private thread zone: a program cannot read or write them.
A host reads them with `vc_memRead*`.

| Offset | Type | Field | Content |
|--------|------|-------|---------|
| `+0x00` | u8 | `state` | State of the thread (see [States](#states)) |
| `+0x01` | u8[3] | reserved | Not used |
| `+0x04` | u32 | `pc` | Program counter. For a thread that is not running, the address where it resumes |
| `+0x08` | u32 | `waitkey` | Event the thread waits for, 0 = none |
| `+0x0C` | u32 | `cartbank` | Cartridge bank register, 16 bits used. See [CSR registers](csr.md#cartbank) |
| `+0x10` | u64 | `sleep_until` | Global cycle at which the thread wakes, 0 = none |
| `+0x18` | u64 | `total_cycles` | Instructions completed by the thread |

All the fields are little-endian.
`new.t` and the reset clear the whole block before they fill `state` and `pc`.
`sp`, `tp` and the canary address are not stored: they follow from the thread number.

## Stacks

Each thread has a stack of 64 KiB (`0x1_0000`) in the stack zone `0x0078_0000` to `0x007F_FFFF`.
The stack grows downward.
The first `sp` is the address just after the stack, so the first push writes at top - 4.

| Thread | Stack base | Stack top (first `sp`) |
|--------|------------|------------------------|
| 0 | `0x0078_0000` | `0x0079_0000` |
| 1 | `0x0079_0000` | `0x007A_0000` |
| 2 | `0x007A_0000` | `0x007B_0000` |
| 3 | `0x007B_0000` | `0x007C_0000` |
| 4 | `0x007C_0000` | `0x007D_0000` |
| 5 | `0x007D_0000` | `0x007E_0000` |
| 6 | `0x007E_0000` | `0x007F_0000` |
| 7 | `0x007F_0000` | `0x0080_0000` |

The base is `0x0078_0000 + tid * 0x1_0000` and the top is `0x0078_0000 + (tid + 1) * 0x1_0000`.

### Stack canary

The first word of each stack (at the base) holds the canary `0xBAAD_F00D`.
The reset and `new.t` write it.
`yield.t`, `sleep.t` (when the thread really sleeps) and `end.t` read it before the thread gives the CPU away.
If the value changed, the stack has overflowed and memory is corrupted, so the overflow is fatal:

1. Every other thread becomes Dead, thread 0 included.
2. The CPU raises a trap (`mcause` 26, see [Traps](traps.md#thread-traps)) on the thread that overflowed. It stays Running and the instruction has no effect: no switch, no sleep, no end.
3. From then on the canary is no longer checked, until the next reset. The trap handler is the only code that runs, and it can wait for the vblank with `sleep.t` (the boot ROM does this to show its error screen).

- With no handler (`mtvec` = 0) the CPU halts, as for any trap.
- A handler that runs `mret` goes back to the thread that overflowed, alone. Do not do it: its memory is corrupted.
- The check finds an overflow that reached the base. It does not stop the overflow.
- An overflow of thread `n` runs into the stack of thread `n - 1`. An overflow of thread 0 runs into Video RAM.

## TLS

Each thread has a thread local storage block of 256 bytes (`0x100`) in the zone `0x0020_6000` to `0x0020_67FF`.
The block of a thread starts at `0x0020_6000 + tid * 0x100`, and `tp` points to it.
The block is plain memory. The console does not clear it when `new.t` takes a slot.

| Thread | TLS block |
|--------|-----------|
| 0 | `0x0020_6000` |
| 1 | `0x0020_6100` |
| 2 | `0x0020_6200` |
| 3 | `0x0020_6300` |
| 4 | `0x0020_6400` |
| 5 | `0x0020_6500` |
| 6 | `0x0020_6600` |
| 7 | `0x0020_6700` |

## Switching threads

The CPU switches thread in these cases only.

| Event | What happens |
|-------|--------------|
| `yield.t` | The current thread becomes Ready. The CPU picks the next Ready thread. If it is the only one, it runs again |
| `sleep.t` with a waitkey or a duration | The current thread becomes Sleeping. The CPU picks the next Ready thread |
| `end.t` | The current thread becomes Dead. The CPU picks the next Ready thread |

The thread that gives the CPU away resumes at the instruction after the one that switched.
The thread that gets the CPU resumes at its saved `pc`. A new thread starts at the address given to `new.t`.

Before it picks the next thread, a switch wakes the threads whose timer is over.
The timers are checked only then, and while the CPU is Idle: a thread that runs for a long time delays the wake-up of the others.
If no thread is Ready after a `sleep.t` or an `end.t`:

- if a thread sleeps with a timer or on a hardware event, the CPU becomes Idle;
- if not, nothing can wake any thread and the CPU raises the deadlock trap (`mcause` 27).

A trap does not switch thread: the handler runs as the thread that trapped.

## Sleeping and waking

A sleeping thread has a waitkey (an event code), a wake-up cycle, or both.
It wakes at the first of these:

- `wake.t` with the same waitkey, from another thread;
- the host triggers the same event with `vc_triggerEvent`;
- the global cycle counter reaches `sleep_until`.

A woken thread becomes Ready, and its `waitkey` and `sleep_until` return to 0.
It does not run at once: it waits for the running thread to give the CPU away, or for the Idle CPU to pick it.

The waitkey ranges are in [Instructions](instructions.md#waitkey-values).
A waitkey of 0 means "no event": `wake.t` with a register at 0 wakes the threads that sleep on a timer only.

## See also

- [CPU](README.md)
- [Instructions](instructions.md)
- [CSR registers](csr.md)
- [Traps](traps.md)
- [Memory map](../memory-map.md)
- [libvc C API](../libvc-api.md)
