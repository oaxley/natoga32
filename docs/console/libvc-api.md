# libvc C API

This page lists the functions of `libvc` that a host program calls, grouped by use, with a minimal host example.

The whole API is in one header, [`core/console/include/vc/vc.h`](../../core/console/include/vc/vc.h).
It is plain C and works from C++ too.
Every function takes the console handle (`VC_Console_t*`) as its first argument.
A `NULL` handle never crashes: the function does nothing and returns `false`, 0, `NULL` or an "empty" value.
The exceptions: `vc_getCpuState` returns `VC_CPU_HALTED`, `vc_getState` returns `VC_STATE_UNINITIALIZED`, `vc_getCurrentThreadId` returns -1, and `vc_getLastError` returns `"Invalid console instance"`.

The host `vchost` uses only this API. See [vchost](../host/vchost.md).

## Lifecycle

A console goes through these calls: create, initialize, load the ROM, reset, run, destroy.

| Function | What it does |
|----------|--------------|
| `VC_Console_t* vc_create(void)` | Allocates a console. Returns `NULL` on failure. The console is not usable yet. |
| `bool vc_initialize(console)` | Allocates the 32 MiB of memory and creates the CPU. Returns `true` at once if already initialized. |
| `void vc_reset(console)` | Clears the memory except the ROM, applies the pending cartridge change, resets the 8 threads (thread 0 at `0x0000_0100`), clears the error and sets the state to Ready. Does nothing before `vc_initialize`. |
| `void vc_shutdown(console)` | Frees the memory and the CPU and drops the cartridge. The state is Uninitialized again. |
| `void vc_destroy(console)` | Shuts the console down and frees the handle. |

The ROM, the cartridge and the framebuffer need an initialized console.
Do not call the memory and thread inspection functions, `vc_getCpuState` and `vc_getCurrentThreadId` on a console that is not initialized, or after `vc_shutdown`: they have no check.
After `vc_reset` the console is Ready, not Running: call `vc_resume` to start it.

## ROM and cartridge loading

| Function | What it does |
|----------|--------------|
| `bool vc_loadRomFile(console, path)` | Reads a ROM file and copies it at `0x0000_0000`. Fails if the file cannot be read, is empty, or is larger than 2 MiB. |
| `bool vc_loadRomData(console, data, size)` | Same, from a memory buffer. |
| `bool vc_loadCartridgeFile(console, path)` | Reads a cartridge file and keeps it aside. Fails only if the file cannot be read or is empty. |
| `bool vc_loadCartridgeData(console, data, size)` | Same, from a memory buffer (copied). Fails if `data` is `NULL` or `size` is 0. |
| `void vc_ejectCartridge(console)` | Asks to remove the cartridge. |
| `uint8_t vc_cartridgeRead8(console, offset)` | One byte of the inserted cartridge file. 0 beyond the end, or with no cartridge. |
| `size_t vc_getCartridgeSize(console)` | Size of the inserted cartridge file in bytes. 0 with no cartridge. |

The ROM is copied at once and survives `vc_reset`. Load it before the console runs.

Loading or ejecting a cartridge takes effect at the next `vc_reset`.
Until then `vc_cartridgeRead8` and `vc_getCartridgeSize` still show the cartridge that was inserted at the last reset.
The device checks the file at that reset: an invalid file is accepted here and reported by the status register of the device, not by these functions.
See [Cartridge insertion](cartridge/insertion.md).

## Running

| Function | What it does |
|----------|--------------|
| `void vc_tick(console)` | Runs one CPU cycle. |
| `void vc_run(console, cycles)` | Runs `cycles` cycles. A stretch where every thread sleeps is skipped in one jump. Stops early if the state leaves Running. |
| `void vc_runFrame(console)` | Runs one frame, 416,666 cycles (25 MHz / 60), then triggers the vblank event `0x8000_0001`. |
| `void vc_pause(console)` | Running becomes Paused. Does nothing in another state. |
| `void vc_resume(console)` | Paused or Ready becomes Running. Does nothing in another state. |

`vc_tick`, `vc_run` and `vc_runFrame` execute cycles only in the Running state.
`vc_runFrame` triggers the vblank event even if the console is not running.
If the CPU halts on an exception, the state becomes Error and the running calls stop.

## State and errors

| Function | What it does |
|----------|--------------|
| `VC_ConsoleState_t vc_getState(console)` | Console state, see the table below. |
| `bool vc_isRunning(console)` | `true` in the Running state. |
| `bool vc_isPaused(console)` | `true` in the Paused state. |
| `uint64_t vc_getTotalCycles(console)` | Cycles since the last reset, idle cycles included. |
| `VC_CPUState_t vc_getCpuState(console)` | CPU state: `VC_CPU_RUNNING` (0), `VC_CPU_IDLE` (1, every thread sleeps) or `VC_CPU_HALTED` (2). |
| `int vc_getCurrentThreadId(console)` | Thread that runs now (0 to 7), or -1 for a `NULL` handle. |
| `const char* vc_getLastError(console)` | Text of the last error. Valid until the next call on the console. |
| `void vc_clearError(console)` | Clears the text. An Error state goes back to Ready. |

| `VC_ConsoleState_t` | Value | Meaning |
|---------------------|-------|---------|
| `VC_STATE_UNINITIALIZED` | 0 | Created, or shut down. |
| `VC_STATE_READY` | 1 | Initialized or reset, not started. |
| `VC_STATE_RUNNING` | 2 | Executes cycles. |
| `VC_STATE_PAUSED` | 3 | Stopped by `vc_pause`. |
| `VC_STATE_ERROR` | 4 | A call failed, or the CPU halted. |

A failed load (a missing file, a ROM that is too large, no initialization) sets the Error state too.
`vc_reset` and `vc_clearError` bring the console back to Ready.
The text tells what failed, for example `Failed to open ROM file: <path>` or `CPU halted due to exception`.

## Events

| Function | What it does |
|----------|--------------|
| `void vc_triggerEvent(console, event)` | Wakes the threads that sleep on this event code (see `sleep.t` in [Threads](cpu/threads.md)). |

An event code is a 32-bit value.
Codes with bit 31 set are hardware events:

| Code | Name | Use |
|------|------|-----|
| `0x8000_0001` | `HW_VBLANK` | Triggered by `vc_runFrame` at the end of each frame. |
| `0x8000_0002` | `HW_HBLANK` | Defined, never triggered. |

The names are in the C++ header [`events.h`](../../core/console/include/vc/events.h), which a C host cannot include: write the values yourself.

## Framebuffer

| Function | What it does |
|----------|--------------|
| `const uint16_t* vc_getFramebuffer(console)` | Pointer to the first pixel of Framebuffer A, or `NULL` if the console is not initialized. |

The buffer has `VC_VIDEO_WIDTH` x `VC_VIDEO_HEIGHT` pixels (640 x 384), 16 bits each, RGB565, row by row (1280 bytes per row).
It is the live memory of the console, not a copy.
The pointer stays valid until `vc_shutdown` or `vc_destroy`, so you can keep it between frames.
See [Framebuffer A](memory-map.md#framebuffer-a).

## Memory and thread inspection

These functions are for debugging. They do not run any cycle.

| Function | What it does |
|----------|--------------|
| `uint8_t vc_memRead8(console, addr)` | Reads a byte. |
| `uint16_t vc_memRead16(console, addr)` | Reads a halfword, little-endian. |
| `uint32_t vc_memRead32(console, addr)` | Reads a word, little-endian. |
| `void vc_memWrite8(console, addr, value)` | Writes a byte. |
| `void vc_memWrite16(console, addr, value)` | Writes a halfword. |
| `void vc_memWrite32(console, addr, value)` | Writes a word. |
| `VC_ThreadState_t vc_getThreadState(console, thread_id)` | State of a thread. |
| `uint32_t vc_getThreadPc(console, thread_id)` | Program counter of a thread. |
| `uint32_t vc_getThreadRegister(console, thread_id, reg_num)` | Register `reg_num` (0 to 31) of a thread. |

A read outside the 32 MiB returns 0, and a write outside is ignored.
A write that touches the ROM, the cartridge control zone or the cartridge window is ignored as a whole.
These accesses do not raise faults and do not use the `cartbank` CSR, so the host can read the [private thread zone](memory-map.md#private-thread-zone).
Writes to the ROM are filtered as the memory does it (see [ROM](memory-map.md#rom)).
A `thread_id` outside 0 to 7, or a `reg_num` outside 0 to 31, returns 0.

| `VC_ThreadState_t` | Value |
|--------------------|-------|
| `VC_THREAD_FREE` | 0 |
| `VC_THREAD_READY` | 1 |
| `VC_THREAD_RUNNING` | 2 |
| `VC_THREAD_SLEEPING` | 4 |
| `VC_THREAD_DEAD` | 8 |

## Minimal host

This program loads a ROM, runs one second (60 frames) and prints the first pixel of Framebuffer A.

```c
#include <stdio.h>
#include "vc/vc.h"

int main(int argc, char** argv)
{
    VC_Console_t* vc = vc_create();
    if (argc < 2 || !vc || !vc_initialize(vc) || !vc_loadRomFile(vc, argv[1])) {
        fprintf(stderr, "error: %s\n", vc ? vc_getLastError(vc) : "no console");
        return 1;
    }
    vc_reset(vc);
    vc_resume(vc);                                 /* Ready -> Running */
    for (int i = 0; i < 60 && vc_getState(vc) == VC_STATE_RUNNING; i++)
        vc_runFrame(vc);                           /* one frame */
    printf("pixel(0,0) = 0x%04x\n", vc_getFramebuffer(vc)[0]);
    vc_destroy(vc);
    return 0;
}
```

To load a cartridge, call `vc_loadCartridgeFile(vc, path)` before `vc_reset`.
Link with `libvc` and the C++ runtime. From the build directory, for a static `libvc`:

```bash
cc -I ../core/console/include vc_example.c core/console/libvc.a -lstdc++ -o vc_example
./vc_example core/host/rombios.sys
```

## See also

- [Console](README.md)
- [Memory map](memory-map.md)
- [Cartridge insertion](cartridge/insertion.md)
- [vchost](../host/vchost.md)
- [Architecture](../overview/architecture.md)
- [Building](../overview/building.md)
