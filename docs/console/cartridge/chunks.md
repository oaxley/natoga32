# Cartridge chunks

This page gives the data layout of the chunks that carry the game: `RSCV` (RISC-V code and data), the asset chunks, and `dSYM` (debug symbols).

The file header, the chunk layout, `IHDR`, `INDX` and `IEND` are in [File format](file-format.md).
All values are little-endian.

## Who checks what

The cartridge device checks the chain of chunks, not their content (see [Insertion](insertion.md#what-the-device-checks)).
The content of `RSCV` and `dSYM` is checked by `n32pack` when it packs the cartridge.
The device never looks at it, and a bad `RSCV` is not an error of the device.

## RSCV

`RSCV` holds the code and the data of the game for the RISC-V CPU.
Its data is a list of **sections**, one after the other.
Each section has a 12-byte header, followed by its bytes.

| Offset | Size | Field |
|--------|------|-------|
| `+0x00` | 1 | Section kind, see below |
| `+0x01` | 1 | Flags (reserved, 0) |
| `+0x02` | 2 | Reserved (0) |
| `+0x04` | 4 | Load address |
| `+0x08` | 4 | Size in bytes |
| `+0x0C` | size | Data, padded with zeros to a multiple of 4. No data for `.bss` |

The section kinds:

| Kind | Section | Load region | Data in the file |
|------|---------|-------------|------------------|
| `0x01` | `.text` | `0x01A0_0000` to `0x01FF_FFFF` | Yes |
| `0x02` | `.data` | `0x0140_0000` to `0x019F_FFFF` | Yes |
| `0x03` | `.rodata` | `0x0120_0000` to `0x013F_FFFF` | Yes |
| `0x04` | `.bss` | `0x0100_0000` to `0x011F_FFFF` | No: only the size, to be zero-filled at load |

The regions are the code and data zones of the [memory map](../memory-map.md).

The rules, checked by `n32pack`:

- Any other kind is refused.
- The flags and the reserved bytes must be 0.
- A section must fit in its region: its load address and its end are inside it.
- Sections must not overlap. A section of size 0 overlaps nothing.
- There must be a `.text` section.
- The sections must fill the chunk exactly.

A kind can appear more than once, for example two `.data` sections at two addresses.
There is no relocation: `n32asm` writes the final addresses.
`n32asm` writes the non-empty sections in this order: `.text`, `.data`, `.rodata`, `.bss`.

The RISC-V entry point is in the [IHDR](file-format.md#ihdr), not in `RSCV`.

Example, the `RSCV` data of the sample cartridge (the sample has 12 bytes of code):

| Section | Header | Data |
|---------|--------|------|
| `.text` | `01 00 0000`, `0x01A0_0000`, `0x0000_000C` | 12 bytes of code |
| `.data` | `02 00 0000`, `0x0140_0000`, `0x0000_0004` | 4 bytes |
| `.rodata` | `03 00 0000`, `0x0120_0000`, `0x0000_0004` | 4 bytes |
| `.bss` | `04 00 0000`, `0x0100_0000`, `0x0000_0010` | none |

Nothing loads these sections today: the boot ROM has no cartridge loader.

## Asset chunks

The asset chunks are `TILS`, `TMAP`, `TPAL`, `TMAT`, `SPRS`, `SOAM`, `SPAL`, `SMAT`, `FNTS`, `FPAL`, `ASMD`, `ATMD`, `ATFM` and `AIFM`.
`n32pack` accepts these 14 types in a manifest.
Each one starts with the same 8-byte header, followed by the payload:

| Offset | Size | Field |
|--------|------|-------|
| `+0x00` | 2 | Slot, from the manifest |
| `+0x02` | 1 | Format, from the manifest |
| `+0x03` | 1 | Flags (reserved, 0) |
| `+0x04` | 2 | Count, from the manifest |
| `+0x06` | 2 | Reserved (0) |
| `+0x08` | | Payload: the content of the asset file, copied as it is |

The slot, the format and the count are numbers that `n32pack` copies from the manifest.
Their meaning, the payload format and the destination of an asset are reserved: nothing in the console uses them.

Example, the asset chunk `TILS` of the sample cartridge: the header is `00 00 04 00 01 00 00 00` (slot 0, format 4, count 1), then the bytes of the asset file.

## dSYM

`dSYM` holds the labels of the program and their addresses.
It is an ancillary chunk: the console never reads it, and a loader can skip it.
`n32asm --sym FILE` writes it, and `n32pack` takes it from the `dsym` field of the manifest.
`n32pack` uses it to find the entry point when the manifest names a label.

The data is a count followed by the entries:

| Offset | Size | Field |
|--------|------|-------|
| `+0x00` | 4 | Number of entries |
| `+0x04` | | The entries, one after the other |

Each entry:

| Offset | Size | Field |
|--------|------|-------|
| `+0x00` | 4 | Address of the label |
| `+0x04` | 1 | Section kind, 1 to 4, the same values as in `RSCV` |
| `+0x05` | 1 | Flags: bit 0 = global. Other bits are 0 |
| `+0x06` | 2 | Length of the name, 1 to 65535 |
| `+0x08` | n | Name in UTF-8, then zero bytes up to a multiple of 4 |

- `n32asm` sets the global bit only when the `.global` line comes after the label. A `.global` line before the label is lost: the bit stays 0 (an `n32asm` bug).
- `n32asm` writes the labels of `.text`, `.data`, `.rodata` and `.bss`, sorted by address, then by name.
- `n32pack` refuses an empty name, a kind outside 1 to 4, a duplicate name, and extra bytes after the last entry.

Example, the last entry of the sample cartridge: address `0x01A0_0008`, kind 1, flags 0, name length 4, name `loop`.

## See also

- [Cartridge](README.md)
- [File format](file-format.md)
- [Console view](console-view.md)
- [Insertion](insertion.md)
- [n32asm output formats](../../tools/n32asm/output-formats.md)
- [n32pack](../../tools/n32pack/README.md)
- [Memory map](../memory-map.md)
