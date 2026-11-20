# Cartridge

This page says what a NATOGA32 cartridge is, which rules its file follows, and how a game reads it while it runs.

## What a cartridge is

A cartridge is one `.n32` file.
It holds the RISC-V code of a game, plus asset chunks and optional debug symbols.

- `goasm` assembles the code into an `RSCV` chunk (and a `dSYM` chunk for the symbols).
- `n32pack` puts these chunks and the asset files together into the `.n32` file.
- The host only inserts the file with [`vc_loadCartridgeFile`](../libvc-api.md#rom-and-cartridge-loading). It never parses it.
- The console checks the file at reset and shows it to the program. See [Insertion](insertion.md).

The format is made of chunks, like PNG.
A reader that finds an unknown optional chunk can skip it.

## General rules

| Rule | Value |
|------|-------|
| Byte order | Little-endian for every multi-byte value |
| CRC-32 | The PNG / zlib CRC-32: polynomial `0xEDB88320` (reflected), initial value `0xFFFF_FFFF`, final XOR `0xFFFF_FFFF`. `crc32("123456789")` is `0xCBF4_3926` |
| File layout | File header, then `IHDR`, `INDX`, other chunks, `IEND` |
| Alignment | Every chunk starts at a file offset that is a multiple of 4 |
| Maximum size | 256 MiB |

`IHDR` is always the first chunk, `INDX` the second and `IEND` the last.
The chunks between `INDX` and `IEND` can come in any order.

## How a game reads its cartridge

The game reads the cartridge through two memory zones.
Both are read-only for programs.

1. It reads `CART_STATUS` to know that a valid cartridge is present.
2. It finds an asset in the `INDX` mirror: the entry of asset `id` is at `0x0022_2000 + id * 16`. The ids come from the `assets.inc` file that `n32pack` writes.
3. It writes the file offset of the data, divided by 4096, into the `cartbank` CSR. This register is private to each thread.
4. It reads the data in the 128 KiB window at `0x0036_0000` and copies it where it is needed.

The window moves in 4 KiB steps, so a block of up to 124 KiB is visible in one go, wherever it starts in the file.
The step-by-step example is in [Console view](console-view.md#example-reading-an-asset).

The host and the debugger do not use the window.
They read the file by absolute offset with [`vc_cartridgeRead8`](../libvc-api.md#rom-and-cartridge-loading).

## Not built

The boot ROM does not read the cartridge yet, and the console has no DMA.
Today a program reads the cartridge itself with the CPU.

## Pages

| Page | Content |
|------|---------|
| [File format](file-format.md) | File header, chunk layout, chunk names, `IHDR`, `INDX`, `IEND` |
| [Chunks](chunks.md) | `RSCV` sections, the asset chunk header, `dSYM` |
| [Console view](console-view.md) | Control zone, status register, error codes, window, `cartbank`, example |
| [Insertion](insertion.md) | Insert and eject, what the device checks at reset |

## See also

- [Console](../README.md)
- [Memory map](../memory-map.md)
- [CSR registers](../cpu/csr.md)
- [libvc C API](../libvc-api.md)
- [n32pack](../../tools/n32pack/README.md)
- [goasm output formats](../../tools/goasm/output-formats.md)
