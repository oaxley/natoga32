# n32pack

This page says what n32pack is, how to build it, what each option does, how it fits in the build of a cartridge, and what it checks.

n32pack is the cartridge packer of NATOGA32, written in Go.
It reads a JSON [manifest](manifest.md), the `RSCV` code chunk and the `dSYM` symbol chunk written by [goasm](../goasm/README.md), and the asset files.
It writes a `.n32` [cartridge](../../console/cartridge/README.md).
It can also write `assets.inc`, a file with the asset ids that the game source includes.

## Build

```bash
cd tools/n32pack
go build -o n32pack .
```

The CMake build does this too, see [Building](../../overview/building.md).

## Usage

```bash
n32pack [-o game.n32] [-inc assets.inc] game.json
```

| Option | Meaning |
|--------|---------|
| `-o FILE` | Write the cartridge to `FILE`. |
| `-inc FILE` | Write the asset ids to `FILE`. See [assets.inc](manifest.md#assetsinc). |

- The options come before the manifest.
- Exactly one manifest is required, and at least one of `-o` and `-inc`.
  Otherwise n32pack prints the usage and exits with status 2.
- With `-inc` alone, only the manifest is read.
  The `rscv`, `dsym` and asset files may not exist yet.
- An error prints `n32pack: <message>` on stderr and exits with status 1.
  Each message names one file: the manifest, or the `rscv`, `dsym` or asset file.
- Everything is built in memory first, so no file is written when the packing fails.
  If `-o` and `-inc` are both given and writing the second file fails, the first one stays.

## Workflow

A game source that uses asset ids includes `assets.inc`.
So the file is generated first, then the code is assembled, then the cartridge is packed.
These are the commands that `tests/CMakeLists.txt` runs for the sample cartridge:

```bash
cd build/tests/cartridge     # the build copies hello.s, hello.json and level1.tiles here

n32pack -inc assets.inc hello.json                       # 1. the asset ids
goasm hello.s -f rscv -o hello.rscv --sym hello.sym      # 2. the RSCV and dSYM chunks
n32pack -o hello.n32 hello.json                          # 3. the cartridge
```

1. `hello.s` has the line `.include "assets.inc"`. The include path is relative to the source file, see the [preprocessor](../goasm/preprocessor.md).
2. `-f rscv` is the default format of goasm. `--sym` writes the labels, so the manifest can name a label as the entry point. See [goasm output formats](../goasm/output-formats.md).
3. The manifest names `hello.rscv` and `hello.sym`. All its paths are relative to the directory of the manifest.

To check the result, load it in the emulator with `vc_cartcheck`, a test program of the build:

```bash
../vc_cartcheck hello.n32      # cartridge valid, 3 chunks indexed
```

## What n32pack writes

The cartridge follows the [file format](../../console/cartridge/file-format.md).
This is what n32pack chooses:

- The chunks are in this order: `IHDR`, `INDX`, `RSCV`, the assets in the order of the manifest, `dSYM` (when given), `IEND`.
- `INDX` lists every chunk between `INDX` and `IEND`. The ids are the positions: `RSCV` is 0, the assets are 1 to n, `dSYM` is n + 1.
- The boot flag is set on `RSCV`, and on the assets with `"boot": true`.
- Each asset chunk is an 8-byte header made from `slot`, `format` and `count`, then the content of the asset file. See [Chunks](../../console/cartridge/chunks.md#asset-chunks).
- The `IHDR` fields come from the manifest. The DSP entry point and the save RAM size are 0 (reserved, not implemented). See [IHDR](../../console/cartridge/file-format.md#ihdr).
- With `graphics_mode`, the header flags are the mode and bit 3 (required). Without it, the flags are 0.
- The output is deterministic: the same manifest with an explicit `date` and the same input files give the same bytes.

## Checks

n32pack refuses to build a cartridge when:

- The manifest is not one valid JSON object, has an unknown field, a missing required field, or a value out of range. The rules of each field are in [Manifest](manifest.md).
- Two assets have the same name.
- The `rscv` or `dsym` file is not exactly one chunk of the right type: length, CRC-32 and zero padding are checked.
- An `RSCV` section has an unknown kind, non-zero reserved bytes, runs past the chunk, is outside the region of its kind, or overlaps another section. There must be a `.text` section. See [RSCV](../../console/cartridge/chunks.md#rscv).
- The `dSYM` data is malformed: kind outside 1 to 4, empty name, duplicate name, entry past the end, or extra bytes at the end. See [dSYM](../../console/cartridge/chunks.md#dsym).
- The entry point is an unknown label, is not 4-byte aligned, or is not inside a `.text` section.
- The date is before 2026-01-01 or more than 2^32 - 1 seconds after it.
- An asset file can't be read.
- There are more than 15,872 `INDX` entries, or the cartridge is over 256 MiB.

The checks of the `RSCV` and `dSYM` content are the only ones: the console never checks them, see [Chunks](../../console/cartridge/chunks.md#who-checks-what).

## Testing

```bash
cd tools/n32pack
go test ./...
```

The tests build cartridges in memory and decode them again.
The build also packs the sample cartridge and `ctest` loads it with `vc_cartcheck`.

## See also

- [Manifest](manifest.md)
- [goasm](../goasm/README.md)
- [Cartridge](../../console/cartridge/README.md)
- [Cartridge file format](../../console/cartridge/file-format.md)
- [Cartridge chunks](../../console/cartridge/chunks.md)
- [Building](../../overview/building.md)
