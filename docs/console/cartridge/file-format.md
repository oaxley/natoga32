# Cartridge file format

This page gives the byte layout of a `.n32` file: the file header, the chunk layout, the chunk names, and the three chunks that every cartridge has (`IHDR`, `INDX`, `IEND`).

All values are little-endian.
The `RSCV`, asset and `dSYM` chunks are in [Chunks](chunks.md).

## File header (16 bytes)

| Offset | Size | Field |
|--------|------|-------|
| `+0x00` | 4 | Magic: `0x7F`, then `"N32"` (`0x7F 0x4E 0x33 0x32`) |
| `+0x04` | 1 | Format version, BCD: `0x10` is version 1.0 |
| `+0x05` | 1 | Reserved (0, not checked) |
| `+0x06` | 2 | Flags, see below |
| `+0x08` | 4 | Total size of the file in bytes |
| `+0x0C` | 4 | CRC-32 of the bytes `0x00` to `0x0B` |

The flags:

| Bits | Meaning |
|------|---------|
| 0-2 | Graphics mode required by the game, 0 to 3. Values 4 to 7 are invalid |
| 3 | 1 = the game requires the mode of bits 0-2. 0 = no requirement, and bits 0-2 must be 0 |
| 4-15 | Reserved, must be 0 |

The device checks the flags but does nothing with the mode.

Example, the first 16 bytes of the sample cartridge `hello.n32`:

```
7F 4E 33 32  10 00 09 00  7C 01 00 00  9C 2E A9 51
magic        ver  - flags size=0x17C   CRC-32
```

The flags `0x0009` mean "mode 1 required".

## Chunk layout

After the header comes a series of chunks.
Every chunk has the same five parts:

| Size | Field |
|------|-------|
| 4 | Length of the chunk data in bytes, padding not included |
| 4 | Chunk type, a FourCC (four ASCII letters) |
| length | Chunk data |
| 0-3 | Padding of zero bytes, so that the next field is 4-byte aligned |
| 4 | CRC-32 of the type and the data (the padding is not included) |

The size of a chunk is `12 + align4(length)`, where `align4(n)` rounds `n` up to a multiple of 4.
The first chunk starts at the file offset 16, and each chunk starts where the previous one ends.

## Chunk names

A chunk type is four letters, upper or lower case.
The case of each letter tells a reader what to do with a chunk it does not know.

| Letter | Uppercase | Lowercase |
|--------|-----------|-----------|
| 1st | Critical: the game needs it. A reader that does not know it refuses the cartridge | Ancillary: an unknown one is skipped |
| 2nd | Public: defined by this format | Private: free for tools |

The 3rd and the 4th letters must be uppercase.

Digits and spaces are not letters and are not allowed.

The device refuses a cartridge in these cases (error 11 in [Console view](console-view.md#error-codes)):

- a chunk type that does not follow the letter rules;
- a critical chunk (first letter uppercase) that is not in the list below.

A lowercase first letter, as in `dSYM`, is accepted by the device even if it does not know the chunk.
The chunk still has to be listed in `INDX`.

## Chunk list

| Type | Kind | Count | Content |
|------|------|-------|---------|
| `IHDR` | critical | 1, first | Game information, see [IHDR](#ihdr) |
| `INDX` | critical | 1, second | Index of the chunks that follow, see [INDX](#indx) |
| `IEND` | critical | 1, last | End of the file, no data |
| `RSCV` | critical | 1 (`n32pack` always writes it) | RISC-V code and data sections |
| `TILS` `TMAP` `TPAL` `TMAT` | critical | 0 or more | Asset chunks |
| `SPRS` `SOAM` `SPAL` `SMAT` | critical | 0 or more | Asset chunks |
| `FNTS` `FPAL` | critical | 0 or more | Asset chunks |
| `ASMD` `ATMD` `ATFM` `AIFM` | critical | 0 or more | Asset chunks |
| `dSYM` | ancillary | 0 or 1 | Debug symbols |

The device also accepts the critical type `DSPC`.
Reserved, not implemented: `n32pack` does not write it.

The device does not count the chunks or require one of them, except `IHDR`, `INDX` and `IEND`.
`n32pack` writes `RSCV`, the assets in the order of the manifest, then `dSYM` when given.
The meaning of the asset types is not defined yet, see [Chunks](chunks.md#asset-chunks).

## IHDR

`IHDR` has 64 bytes of data, and every field has a fixed position.
The chunk is at the file offset 16, so its data starts at the file offset `0x18`.

| Offset | Size | Field | Rule checked by the device |
|--------|------|-------|----------------------------|
| `+0x00` | 2 | Game version, BCD (`0x0100` is 1.00) | None |
| `+0x02` | 1 | Minimum boot ROM version, major | None |
| `+0x03` | 1 | Minimum boot ROM version, minor | None |
| `+0x04` | 1 | Regions: bit 0 Americas, bit 1 Europe, bit 2 Asia | Not 0, and no bit above bit 2 |
| `+0x05` | 1 | Reserved | Must be 0 |
| `+0x06` | 2 | Languages: bit 0 English | None |
| `+0x08` | 4 | Creation date, in seconds since 2026-01-01 00:00 UTC | None |
| `+0x0C` | 4 | RISC-V entry point | In `.text`, `0x01A0_0000` to `0x01FF_FFFF` |
| `+0x10` | 4 | DSP entry point | Any value. `n32pack` writes 0. Reserved, not implemented |
| `+0x14` | 2 | Save RAM size | Must be 0. Reserved, must be 0 |
| `+0x16` | 2 | Reserved | Must be 0 |
| `+0x18` | 32 | Game name, UTF-8, padded with zeros | None |
| `+0x38` | 8 | Reserved | All 0 |

The game name has no terminator when it is 32 bytes long.
The device does not compare the minimum boot ROM version or the regions with anything: it only checks the rules of the last column.
The `IHDR` chunk must have a length of exactly 64, and a valid CRC.

## INDX

`INDX` lists every chunk between `INDX` and `IEND`, in file order.
It has one 16-byte entry per chunk:

| Offset | Size | Field |
|--------|------|-------|
| `+0x00` | 4 | Chunk type |
| `+0x04` | 2 | Asset id: the position of the entry in the index (0, 1, 2, ...) |
| `+0x06` | 2 | Flags. Bit 0 = boot. The other bits must be 0 |
| `+0x08` | 4 | File offset of the chunk data (the chunk starts 8 bytes before) |
| `+0x0C` | 4 | Length of the chunk data, the same as in the chunk |

- The number of entries is `length / 16`. The length must be a multiple of 16.
- The maximum is **15,872** entries, the size of the index mirror in the [control zone](console-view.md#mirrors). `n32pack` refuses more.
- The asset id is the position of the entry. It is a consistency check, and it is the number a game uses.
- The boot flag marks a chunk that the loader should load at start-up. `n32pack` sets it on `RSCV`, and on the assets that have `"boot": true`. No boot loader is built, so nothing reads it today.
- `dSYM` is listed like any other chunk, without the boot flag.

The device compares the index with the real chunks: type, id, offset, length, and the number of entries.
An entry or a chunk that does not match gives error 12.

With `n32pack`, `RSCV` is id 0, the assets are 1 to n in the manifest order, and `dSYM` is id n + 1.

## IEND

`IEND` ends the cartridge.
Its length is 0, and nothing may follow it.
The end of `IEND` must be the end of the file.

## See also

- [Cartridge](README.md)
- [Chunks](chunks.md)
- [Console view](console-view.md)
- [Insertion](insertion.md)
- [Memory map](../memory-map.md)
- [n32pack](../../tools/n32pack/README.md)
