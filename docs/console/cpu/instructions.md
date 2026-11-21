# Thread instructions

This page describes the six custom instructions that create, switch, sleep and end threads, and the event codes that `sleep.t` and `wake.t` use.

## Overview

The thread instructions use the RISC-V custom-0 opcode `0x0B`.
The field `funct3` selects the instruction.

| Bits | 31 to 25 | 24 to 20 | 19 to 15 | 14 to 12 | 11 to 7 | 6 to 0 |
|------|----------|----------|----------|----------|---------|--------|
| Field | `funct7` (0) | `rs2` | `rs1` | `funct3` | `rd` | opcode (`0x0B`) |

| Instruction | Syntax | goasm format | `funct3` | Fields used | Example |
|-------------|--------|--------------|----------|-------------|---------|
| `new.t` | `new.t rd, rs1` | R | `000` | `rd`, `rs1` | `new.t a1, a0` = `0x0005_058B` |
| `yield.t` | `yield.t` | I | `001` | none | `0x0000_100B` |
| `id.t` | `id.t rd` | I | `010` | `rd` | `id.t a2` = `0x0000_260B` |
| `sleep.t` | `sleep.t rs1, rs2` | R2 | `100` | `rs1`, `rs2` | `sleep.t a0, a1` = `0x00B5_400B` |
| `wake.t` | `wake.t rs1` | R2 | `101` | `rs1` | `wake.t a0` = `0x0005_500B` |
| `end.t` | `end.t` | I | `111` | none | `0x0000_700B` |

The formats are the names that [goasm](../../tools/goasm/instructions.md) uses.
R2 has two source registers and `rd` = 0.
`wake.t` can also be written `wake.t rs1, rs2`: `rs2` is ignored, and it can be `zero`.
The CPU ignores the fields that an instruction does not use.
The two other values of `funct3` (`011` and `110`) do nothing: the PC moves to the next instruction.

Every thread instruction moves the PC to the next instruction, unless it traps (the PC then goes to `mtvec`).
A thread that gives the CPU away resumes there.

## Instructions

### new.t

`new.t rd, rs1` starts a new thread.

1. The CPU looks for the first Free thread, from thread 1 up to thread 7.
2. It resets the slot: registers at 0, `sp` and `tp` set, control block cleared, canary written. See [Threads](threads.md).
3. The new thread starts at the address in `rs1` and its state is Ready.
4. `rd` receives the number of the new thread (1 to 7).

The new thread does not run now: the current thread keeps the CPU.
If no slot is Free, the CPU raises a trap with `mcause` 25 and `rd` is unchanged.
A Dead slot is not Free, see [Threads](threads.md#states).

### yield.t

`yield.t` gives the CPU to the next Ready thread, in the order of the thread numbers.
The current thread becomes Ready.
If it is the only thread that can run, it runs again.
The CPU checks the [stack canary](threads.md#stack-canary) first.

### id.t

`id.t rd` writes the number of the current thread (0 to 7) in `rd`.

### sleep.t

`sleep.t rs1, rs2` puts the current thread to sleep and gives the CPU away at once, like `yield.t`.

| Register | Value | Meaning |
|----------|-------|---------|
| `rs1` | waitkey | The event that wakes the thread. 0 = no event |
| `rs2` | duration | The number of cycles to sleep. 0 = no timer |

- With a waitkey only, the thread sleeps until `wake.t` or `vc_triggerEvent` uses this waitkey.
- With a duration only, the thread sleeps for this number of cycles, counted from the cycle of the `sleep.t`.
- With both, the thread wakes at the first one.
- With both at 0, `sleep.t` does nothing and the thread keeps running.

The registers are read as unsigned 32-bit values.
A duration of 25,000,000 is one second.
The wake-up is not exact: the timers are checked when the CPU switches thread or is Idle, see [Threads](threads.md#switching-threads).

### wake.t

`wake.t rs1` wakes all the sleeping threads whose waitkey is equal to `rs1`.
They become Ready and they run when their turn comes.
`wake.t` does not give the CPU away, and it does nothing when no thread matches.
`wake.t` with `rs1` at 0 matches the threads that sleep on a timer only.

### end.t

`end.t` ends the current thread: its state becomes Dead and the CPU switches to the next Ready thread.
Thread 0 cannot end: `end.t` on thread 0 raises a trap with `mcause` 24.
The stack canary is not checked.

## Waitkey values

A waitkey is a 32-bit value.
The CPU looks at bit 31 only: it tells a hardware event from a software key.
The ranges below are the comment in `core/console/include/vc/events.h`. Nothing in the code enforces them.

| Waitkey | Kind | Wakes when |
|---------|------|------------|
| `0x0000_0000` | No event | `sleep.t` with a waitkey of 0 sleeps on its timer only |
| `0x0000_0000` to `0x01FF_FFFF` | Valid RAM addresses (32 MiB) | A thread runs `wake.t` with the same value |
| `0x1000_0000` to `0x1FFF_FFFF` | Software events, user-defined | A thread runs `wake.t` with the same value, or the host triggers it |
| `0x8000_0000` to `0x8FFF_FFFF` | Hardware events (bit 31 set) | The host triggers it with `vc_triggerEvent` |

A software key has no meaning for the CPU.
Use an address to wait on a variable or a lock, or any number that the program chooses.
The CPU does not tell an address from a software event: only bit 31 counts.

The CPU becomes Idle when every thread sleeps and one of them waits for a hardware event or a timer.
A hardware event wakes the thread even if no thread can call `wake.t`.
If every thread sleeps on a software key without a timer, the CPU raises the deadlock trap (`mcause` 27).

| Event | Value | Use |
|-------|-------|-----|
| `HW_VBLANK` | `0x8000_0001` | Triggered by the host at the end of each frame (416,666 cycles), see [libvc C API](../libvc-api.md#events) |
| `HW_HBLANK` | `0x8000_0002` | Defined, never triggered |

## Example: wait for the vblank

A program that draws a frame, then sleeps until the next vblank:

```asm
.equ HW_VBLANK, 0x80000001

.text
.global main
main:
    li   s1, 0
frame:
    li   a0, HW_VBLANK
    sleep.t a0, zero        # Idle until the next vblank
    addi s1, s1, 1          # runs once per frame
    j    frame
```

The thread sleeps on a hardware event without a timer.
The CPU becomes Idle and `vc_run` skips the time.
When the host triggers the vblank, the thread is Ready and the CPU runs it.
The [boot ROM](../../rom/README.md) waits for the frames the same way.

## Example: a second thread

```asm
.equ KEY, 0x10000001        # a software key

.text
.global main
main:
    la   a0, worker
    new.t s1, a0            # s1 = 1, the number of the new thread
    yield.t                 # the worker runs until it sleeps
    li   a0, KEY
    wake.t a0               # the worker is Ready again
    yield.t                 # the worker runs until it ends
    li   s2, 1
done:
    j    done

worker:
    li   a0, KEY
    sleep.t a0, zero        # sleep until wake.t KEY
    li   s3, 7
    end.t
```

The order is: thread 0 starts the worker, the worker sleeps, thread 0 wakes it, the worker sets its own `s3` and ends.

## See also

- [CPU](README.md)
- [Threads](threads.md)
- [Traps](traps.md)
- [libvc C API](../libvc-api.md)
- [goasm instructions](../../tools/goasm/instructions.md)
- [Boot ROM](../../rom/README.md)
