# Console

This page describes the NATOGA32 console as a program sees it, and links to the pages of each part.

## What a program sees

| Part | What it is |
|------|------------|
| CPU | One 32-bit RISC-V core running at 25 MHz, with 8 cooperative threads. |
| Memory | 32 MiB in one flat address space, with fixed zones for the ROM, the registers, the stacks and the program sections. |
| Cartridge device | Shows the inserted `.n32` file to the program through control registers and a 128 KiB window. |
| Framebuffer A | A 640 x 384 RGB565 picture in memory. The host shows it on the screen. |

There is no operating system. The boot ROM runs first, and a program talks to the hardware by reading and writing memory and CSR registers.

Video RAM has room for more than Framebuffer A, but only Framebuffer A is built.
The other memories of the map (S-RAM, Audio RAM) are reserved, not implemented.

## The CPU

- One core, 8 threads. Thread 0 starts at `0x0000_0100` at reset.
- Threads are cooperative: the running thread keeps the CPU until it sleeps, yields or ends.
- The clock is 25 MHz and a frame is 416,666 cycles (60 frames per second).
- Time is counted in cycles. When every thread sleeps, the console skips the time instead of running it.

The details are in [CPU](cpu/README.md), [Threads](cpu/threads.md), [Instructions](cpu/instructions.md), [CSR registers](cpu/csr.md) and [Traps](cpu/traps.md).

## The memory

The 32 MiB are split in zones, from the 2 MiB ROM at `0x0000_0000` to the 6 MiB of code that end at `0x01FF_FFFF`.
Some zones have rules: the private thread zone raises an access fault, and the cartridge zones are read-only.
See the [memory map](memory-map.md).

## The cartridge

A cartridge is a `.n32` file built with `n32asm` and `n32pack`.
The console checks it at reset, then a program reads its chunk index and its data through the control zone and the window.
See [Cartridge](cartridge/README.md).

## Running the console

The host owns the loop. It creates the console, loads the ROM and a cartridge, resets it, then asks for one frame at a time.
After each frame, the vblank event wakes the threads that wait for it, and the host copies Framebuffer A to the screen.
See the [libvc C API](libvc-api.md) for the functions, and [vchost](../host/vchost.md) for the host program.

## See also

- [Memory map](memory-map.md)
- [libvc C API](libvc-api.md)
- [CPU](cpu/README.md)
- [Cartridge](cartridge/README.md)
- [Boot ROM](../rom/README.md)
- [Design principles](../overview/principles.md)
- [Architecture](../overview/architecture.md)
- [Documentation index](../README.md)
