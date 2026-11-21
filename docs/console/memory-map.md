# Memory map

This page lists every zone of the 32 MiB address space of the console, with its address, its size and who may read or write it.

## Overview

The CPU sees one flat address space of 32 MiB (`0x0000_0000` to `0x01FF_FFFF`).
Addresses outside it read as 0, and writes to them are ignored.
All the values are little-endian.

```
0x0000_0000  +--------------------------+
             | ROM / boot loader   2 MiB|
0x0020_0000  +--------------------------+
             | Interrupts vector  128 B |
0x0020_1000  +--------------------------+
             | CSR registers      16 KiB|
0x0020_5000  +--------------------------+
             | Private thread zone 4 KiB|  access fault for programs
0x0020_6000  +--------------------------+
             | Thread TLS          2 KiB|
0x0020_6800  +--------------------------+
             | Unused            102 KiB|
0x0022_0000  +--------------------------+
             | Cartridge control 256 KiB|  read-only
0x0026_0000  +--------------------------+
             | S-RAM             512 KiB|  reserved, not implemented
0x002E_0000  +--------------------------+
             | Audio RAM         512 KiB|  reserved, not implemented
0x0036_0000  +--------------------------+
             | Cartridge window  128 KiB|  read-only
0x0038_0000  +--------------------------+
             | Video RAM           4 MiB|  Framebuffer A at its start
0x0078_0000  +--------------------------+
             | Stacks        8 x 64 KiB |
0x0080_0000  +--------------------------+
             | Heap                8 MiB|
0x0100_0000  +--------------------------+
             | .bss                2 MiB|
0x0120_0000  +--------------------------+
             | .rodata             2 MiB|
0x0140_0000  +--------------------------+
             | .data               6 MiB|
0x01A0_0000  +--------------------------+
             | .text (code)        6 MiB|
0x0200_0000  +--------------------------+  end of memory
```

## Zones

| Zone | Start | End | Size | Access rules |
|------|-------|-----|------|--------------|
| ROM / boot loader | `0x0000_0000` | `0x001F_FFFF` | 2 MiB | Read-only: writes are ignored. See [ROM](#rom). |
| Interrupts vector | `0x0020_0000` | `0x0020_007F` | 128 B | 32 words. Plain memory, not used by the CPU. |
| Unused | `0x0020_0080` | `0x0020_0FFF` | 3,968 B | Plain memory. |
| CSR registers | `0x0020_1000` | `0x0020_4FFF` | 16 KiB | 4096 words, the value of CSR `n` is at `0x0020_1000 + 4 * n`. Programs use the CSR instructions. |
| Private thread zone | `0x0020_5000` | `0x0020_5FFF` | 4 KiB | Emulator only. Any load or store raises an access fault. See [Private thread zone](#private-thread-zone). |
| Thread TLS | `0x0020_6000` | `0x0020_67FF` | 2 KiB | Plain memory. 256 B per thread. |
| Unused | `0x0020_6800` | `0x0021_FFFF` | 102 KiB | Plain memory. |
| Cartridge control | `0x0022_0000` | `0x0025_FFFF` | 256 KiB | Read-only, not executable. See [Cartridge zones](#cartridge-zones). |
| S-RAM | `0x0026_0000` | `0x002D_FFFF` | 512 KiB | Reserved, not implemented. |
| Audio RAM | `0x002E_0000` | `0x0035_FFFF` | 512 KiB | Reserved, not implemented. |
| Cartridge window | `0x0036_0000` | `0x0037_FFFF` | 128 KiB | Read-only, not executable. See [Cartridge zones](#cartridge-zones). |
| Video RAM | `0x0038_0000` | `0x0077_FFFF` | 4 MiB | Plain memory. Only Framebuffer A is used. See [Framebuffer A](#framebuffer-a). |
| Stacks | `0x0078_0000` | `0x007F_FFFF` | 512 KiB | Plain memory. 64 KiB per thread. |
| Heap | `0x0080_0000` | `0x00FF_FFFF` | 8 MiB | Plain memory. |
| `.bss` | `0x0100_0000` | `0x011F_FFFF` | 2 MiB | Plain memory. |
| `.rodata` | `0x0120_0000` | `0x013F_FFFF` | 2 MiB | Read-only by convention. See [.rodata](#rodata). |
| `.data` | `0x0140_0000` | `0x019F_FFFF` | 6 MiB | Plain memory. |
| `.text` | `0x01A0_0000` | `0x01FF_FFFF` | 6 MiB | Plain memory. Code runs from here. |

"Plain memory" means that loads and stores work and no check is made.
Reserved zones are plain memory too, but nothing in the console uses them.

The reset clears all the memory except the ROM.

## ROM

The ROM is 2 MiB, loaded by the host with [`vc_loadRomFile` or `vc_loadRomData`](libvc-api.md#rom-and-cartridge-loading).
A ROM larger than 2 MiB is refused.
The ROM survives `vc_reset`.

The memory ignores every write to the ROM (byte, half-word or word), with no fault.
A write that is partly in the ROM is ignored as a whole.

On reset, thread 0 starts at `0x0000_0100`, the entry point of the ROM.

### ROM header

The first 256 bytes of the ROM (`0x0000_0000` to `0x0000_00FF`) are the ROM header.
The code starts at `0x0000_0100`.
The boot ROM writes this header in [`rom/boot.s`](../../rom/boot.s).
The console does not read or check it.

| Offset | Type | Field |
|--------|------|-------|
| `+0x00` | u32 | Magic `N32R` (`0x5232334E`) |
| `+0x04` | u16 | Version major (0) |
| `+0x06` | u16 | Version minor (1) |
| `+0x08` | u32 | Entry point (`0x0000_0100`) |
| `+0x0C` | u32 | Address of the trap handler |
| `+0x10` | u32 | Address of the syscall table (0 = none) |
| `+0x14` | u32 | Address of the logo |
| `+0x18` | u32 | Address of the font |
| `+0x1C` | | Reserved, zero up to `+0xFF` |

## Private thread zone

The zone `0x0020_5000` to `0x0020_5FFF` holds the state of the 8 threads.
Only the emulator uses it.

| Part | Start | End | Size | Content |
|------|-------|-----|------|---------|
| Registers save | `0x0020_5000` | `0x0020_53FF` | 1 KiB | 32 registers of 4 B for each of the 8 threads |
| Thread control blocks | `0x0020_5400` | `0x0020_54FF` | 256 B | 32 B for each thread |
| Reserved | `0x0020_5500` | `0x0020_5FFF` | 2,816 B | Not used |

A program load or store that touches this zone, even by one byte, raises an access fault (a load access fault or a store access fault, see [Traps](cpu/traps.md)).
The host can still read it with `vc_memRead*` for debugging.
The content of the blocks is described in [Threads](cpu/threads.md).

## Cartridge zones

Two zones show the inserted cartridge to the program.
Both are read-only for programs.

| Zone | Content |
|------|---------|
| Cartridge control (`0x0022_0000`) | The control registers, a mirror of the cartridge header, the IHDR and the INDX. |
| Cartridge window (`0x0036_0000`) | A 128 KiB view of the cartridge file, moved with the `cartbank` CSR. |

A store to these zones raises a store access fault.
A load that crosses the edge of one of them raises a load access fault.
An instruction fetch that touches them raises an instruction access fault.

The registers and the window are described in [Cartridge: console view](cartridge/console-view.md).

## Framebuffer A

Framebuffer A is the only video memory in use. It is at the start of Video RAM.

| Property | Value |
|----------|-------|
| Address | `0x0038_0000` |
| Size in pixels | 640 x 384 |
| Pixel format | RGB565, 16 bits, little-endian |
| Row size | 1280 bytes (640 x 2) |
| Size | 491,520 bytes, `0x0038_0000` to `0x003F_7FFF` |

The pixel `(x, y)` is at `0x0038_0000 + y * 1280 + x * 2`.
A pixel has red in bits 15 to 11, green in bits 10 to 5, and blue in bits 4 to 0.
A host reads the frame with [`vc_getFramebuffer`](libvc-api.md#framebuffer).

The rest of Video RAM is reserved, not implemented.

## .rodata

`.rodata` is read-only by convention: the emulator raises no fault on a write to it.
Keep the constants of a program there and do not write to them.

## See also

- [Console](README.md)
- [libvc C API](libvc-api.md)
- [Threads](cpu/threads.md)
- [Traps](cpu/traps.md)
- [Cartridge: console view](cartridge/console-view.md)
- [Boot ROM](../rom/README.md)
- [Architecture](../overview/architecture.md)
