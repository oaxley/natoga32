# Cartridge: console view

This page describes what a program sees of the cartridge: the control zone with its registers and mirrors, the status and error codes, the 128 KiB window, bank switching, and the faults.

The file itself is described in [File format](file-format.md).
The address of each zone is in the [memory map](../memory-map.md).

## Control zone

The cartridge control zone is `0x0022_0000` to `0x0025_FFFF` (256 KiB).
It is read-only: a store raises a store access fault.
An instruction fetch from it raises an instruction access fault.

| Part | Start | End | Size | Content |
|------|-------|-----|------|---------|
| Control registers | `0x0022_0000` | `0x0022_0FFF` | 4 KiB | The three registers below, the rest reads 0 |
| Header and IHDR mirror | `0x0022_1000` | `0x0022_1FFF` | 4 KiB | Copy of the file header and of the `IHDR` data |
| INDX mirror | `0x0022_2000` | `0x0025_FFFF` | 248 KiB | Copy of the `INDX` data |

### Registers

The registers are 32 bits wide. Read them with `lw`.

| Address | Name | Content |
|---------|------|---------|
| `0x0022_0000` | `CART_STATUS` | Bit 0 = present, bit 1 = valid, bits 8-15 = error code |
| `0x0022_0004` | `CART_SIZE` | Size of the file in bytes |
| `0x0022_0008` | `CART_INDX_COUNT` | Number of entries in `INDX` |

`CART_STATUS`, by case:

| Case | Value | Other registers |
|------|-------|-----------------|
| No cartridge | `0x0000_0000` | All 0 |
| Present, valid | `0x0000_0003` | `CART_SIZE`, `CART_INDX_COUNT` and the mirrors are filled |
| Present, invalid | `0x0000_0001 \| (error << 8)` | Only `CART_SIZE` is filled. The mirrors and `CART_INDX_COUNT` read 0 |

### Error codes

The error code is in bits 8-15 of `CART_STATUS`.
For example, a bad header CRC gives `0x0000_0301`.
The errors 1, 2, 3, 13 and 4 are checked in this order. Then the chunks are checked one by one, and the first bad chunk gives the code.

| Code | Name | Meaning |
|------|------|---------|
| 0 | None | No error |
| 1 | Bad magic | The file is shorter than 16 bytes, or the magic is wrong |
| 2 | Bad version | The format version is not `0x10` |
| 3 | Header CRC | The CRC of the header does not match |
| 13 | Bad flags | A reserved flag bit is set, or the mode is not valid |
| 4 | Bad size | The size in the header is not the file size, or it is over 256 MiB |
| 5 | Bad `IHDR` | `IHDR` is missing or not first, appears again later, its length is not 64, or a field is not valid |
| 6 | `IHDR` CRC | The CRC of `IHDR` does not match |
| 7 | Bad `INDX` | `INDX` is missing or not second, appears again later, or its length is not a multiple of 16 |
| 8 | `INDX` CRC | The CRC of `INDX` does not match |
| 9 | `INDX` too big | `INDX` has more than 15,872 entries |
| 10 | Bad `IEND` | `IEND` is missing, not last, not empty, or a chunk runs past the end of the file |
| 11 | Bad chunk | A chunk type is not valid, or it is a critical type that is not known |
| 12 | `INDX` mismatch | `INDX` does not match the chunks |

The code 13 is checked before the code 4, so the table is not sorted by code.
The fields of `IHDR` and the rules of the chunk names are in [File format](file-format.md).

### Mirrors

The mirrors are filled only for a valid cartridge.

| Address | Content |
|---------|---------|
| `0x0022_1000` | The 16 bytes of the file header |
| `0x0022_1010` | The 64 bytes of `IHDR` data. The rest up to `0x0022_1FFF` reads 0 |
| `0x0022_2000` | The `INDX` data. The entry of asset `id` is at `0x0022_2000 + id * 16` |

So the fields are at fixed addresses.
For example, the RISC-V entry point is at `0x0022_1010 + 0x0C`, which is `0x0022_101C`.

## Window

The cartridge window is `0x0036_0000` to `0x0037_FFFF` (128 KiB).
It shows a part of the file.
It is read-only and not executable:

| Access | Result |
|--------|--------|
| Load | The byte of the file, see below |
| Store | Store access fault (`mcause` 7) |
| Instruction fetch | Instruction access fault (`mcause` 1) |

Code from the cartridge must be copied to `.text` before it runs.

### Bank switching

Which part of the file is visible depends on the `cartbank` CSR (`0x7C0`).
See [CSR registers](../cpu/csr.md#cartbank) for the register.

```
file offset = (cartbank << 12) + (address - 0x0036_0000)
```

- The bank moves in 4 KiB steps, and the window is 128 KiB. A block of up to 124 KiB is visible in one go, wherever it starts.
- 65,536 banks of 4 KiB give 256 MiB, the maximum size of a cartridge.
- Each thread has its own `cartbank`, so a thread that changes its bank does not disturb another thread that copies data.
- A load reads the bank of the thread that runs it.
- A byte past the end of the file reads 0.
- The window reads 0 when there is no cartridge, or when the cartridge is not valid.

A load of 2 or 4 bytes reads consecutive bytes of the file, from the same bank.
A load must be entirely inside the window: a load that crosses the edge of the window raises a load access fault.

### Access faults

A store, or an instruction fetch, that touches the control zone or the window raises an access fault. A load that crosses the edge of a zone raises a load access fault.
The table is in [Traps](../cpu/traps.md#access-faults), and the rules for a program in [Memory map](../memory-map.md).

## Example: reading an asset

The sample cartridge has one asset, `LEVEL1_TILES`.
`n32pack -inc assets.inc hello.json` generates the id:

```asm
# assets.inc (generated)
.equ ASSET_LEVEL1_TILES, 1
```

This program checks the status, finds the asset in the `INDX` mirror, selects the bank and reads the first words of the asset chunk.
It assembles with `goasm cart_read.s -f rscv -o cart_read.rscv`.

```asm
.include "assets.inc"

.equ CART_STATUS,   0x00220000
.equ CART_INDX,     0x00222000
.equ CART_WINDOW,   0x00360000
.equ CSR_CARTBANK,  0x7C0

.text
.global main
main:
    li   t0, CART_STATUS
    lw   t1, 0(t0)                  # CART_STATUS
    andi t1, t1, 2                  # bit 1: the cartridge is valid
    beqz t1, no_cart

    li   t0, CART_INDX
    li   t1, ASSET_LEVEL1_TILES
    slli t1, t1, 4                  # id * 16
    add  t0, t0, t1                 # INDX entry of the asset
    lw   t2, 8(t0)                  # file offset of the chunk data
    lw   t3, 12(t0)                 # data length

    srli t4, t2, 12                 # bank = offset / 4 KiB
    csrw CSR_CARTBANK, t4           # this thread only
    li   t5, 0xFFF
    and  t2, t2, t5                 # offset inside the bank
    li   t5, CART_WINDOW
    add  t2, t2, t5                 # address of the data in the window

    lw   a0, 0(t2)                  # asset header: slot (u16) and format (u8)
    lw   a1, 4(t2)                  # asset header: count (u16)
    lw   a2, 8(t2)                  # first 4 bytes of the payload
done:
    j    done
no_cart:
    j    no_cart
```

Run as a test ROM (assembled with `-f rom`, with `.text` replaced by `.rom` and `.org 0x100`) with `hello.n32` inserted, the registers end with:

| Register | Value | Meaning |
|----------|-------|---------|
| `a0` | `0x0004_0000` | Slot 0, format 4, flags 0 |
| `a1` | `0x0000_0001` | Count 1 |
| `a2` | `0x4556_454C` | `"LEVE"`, the start of the asset file |

A copy loop would then read `t3` bytes from the address in `t2`.
The data starts less than 4 KiB into the window, so up to 124 KiB of it fit in this one bank.

## See also

- [Cartridge](README.md)
- [File format](file-format.md)
- [Insertion](insertion.md)
- [Memory map](../memory-map.md)
- [CSR registers](../cpu/csr.md)
- [Traps](../cpu/traps.md)
