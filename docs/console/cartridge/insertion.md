# Cartridge insertion

This page explains how a cartridge enters the console: how the host inserts and ejects it, what the device checks at reset, and what happens with no cartridge or an invalid one.

The functions are in the [libvc C API](../libvc-api.md#rom-and-cartridge-loading).
The registers the program reads are in [Console view](console-view.md).

## Insert and eject

The host gives a file to the console with `vc_loadCartridgeFile` or `vc_loadCartridgeData`.
It asks to remove the cartridge with `vc_ejectCartridge`.
These calls do not change the console at once: the change is kept aside until the next `vc_reset`.

| Step | What happens |
|------|--------------|
| `vc_loadCartridgeFile` / `vc_loadCartridgeData` | The file is stored as pending. It is not checked. A call replaces the previous pending file |
| `vc_ejectCartridge` | The pending change becomes "no cartridge" |
| `vc_reset` | The pending change is applied, then the device checks the file and fills the control zone |

Until the reset:

- the control zone and the window show the cartridge of the last reset, and a running game is not disturbed;
- `vc_cartridgeRead8` and `vc_getCartridgeSize` show that same cartridge.

Without a new call, a reset keeps the cartridge that is already inserted and checks it again.
`vc_shutdown` drops the cartridge, inserted or pending.

The load calls fail only when the file cannot be read, is empty, or when the data pointer is null or the size is 0.
A file that is not a cartridge is accepted by them.
The check at reset reports it.

## What the device checks

At each reset the memory is cleared, then the device checks the file and fills the control zone.
The order of the checks is the order of the error codes in [Console view](console-view.md#error-codes):

| Checked | Error |
|---------|-------|
| File of at least 16 bytes, magic `7F 4E 33 32` | 1 |
| Format version `0x10` | 2 |
| CRC-32 of the header | 3 |
| Header flags: reserved bits, mode 0 to 3, mode bit | 13 |
| Size in the header equal to the file size, at most 256 MiB | 4 |
| `IHDR` first, length 64, valid fields (regions, reserved fields, save RAM = 0, RISC-V entry in `.text`) | 5 |
| CRC-32 of `IHDR` | 6 |
| `INDX` second, length a multiple of 16 | 7 |
| CRC-32 of `INDX` | 8 |
| At most 15,872 `INDX` entries | 9 |
| Every chunk fits in the file, `IEND` last and empty, nothing after it | 10 |
| Valid chunk names, no unknown critical chunk | 11 |
| `INDX` has one correct entry for each chunk (type, id, flags, offset, length) | 12 |

The first error stops the check.
The comparison with `INDX` is made last, when the walk reaches `IEND`.

## What the device does not check

- The CRC-32 of the chunks other than `IHDR` and `INDX`.
- The content of `RSCV`, of the asset chunks and of `dSYM`. See [Chunks](chunks.md).
- The reserved byte `+0x05` of the file header.
- That a chunk such as `RSCV` is present, or that the boot flags of `INDX` are set.
- The minimum boot ROM version and the regions of `IHDR`. They are not compared with anything.
- The meaning of the graphics mode in the flags.

`n32pack` checks the CRC and the content of the chunks when it builds the cartridge.
A program can also check a chunk CRC itself, by reading the chunk through the window.

## No cartridge, or an invalid one

| Case | `CART_STATUS` | Control zone | Window |
|------|---------------|--------------|--------|
| No cartridge | `0x0000_0000` | All 0 | Reads 0 |
| Invalid cartridge | `0x0000_0001 \| (error << 8)` | `CART_SIZE` only, the rest is 0 | Reads 0 |
| Valid cartridge | `0x0000_0003` | Registers and mirrors filled | Shows the file |

An invalid cartridge is not an error of the console: the state stays Ready, and the CPU starts as usual.
Only the status register tells the program.

The boot ROM does not read the cartridge yet.
It always shows its "insert a cartridge" message, even with a valid cartridge inserted.

## Host and debugger access

The host does not use the window and does not change `cartbank`.
It reads the file by absolute offset with `vc_cartridgeRead8(console, offset)`.
The call works for a valid and for an invalid cartridge, and returns 0 past the end of the file or with no cartridge.

The host can also read the control zone with `vc_memRead32`.
For example, `vc_cartcheck` in the tests resets the console and prints the result of `CART_STATUS` and `CART_INDX_COUNT`.

## Not built

The boot ROM has no cartridge loader: no check of the minimum boot ROM version or of the regions, no loading of the sections, no jump to the entry point of the game.

## See also

- [Cartridge](README.md)
- [Console view](console-view.md)
- [File format](file-format.md)
- [libvc C API](../libvc-api.md)
- [vchost](../../host/vchost.md)
- [Boot ROM](../../rom/README.md)
