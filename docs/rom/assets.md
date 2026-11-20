# ROM assets

This page describes how the logo and the font of the boot ROM are converted by `make_assets.py`, the layout of the two output files, and how to test the script.

## Overview

The boot ROM includes two binary files with `.incbin`:

| File | Size | Content | Used as |
|------|------|---------|---------|
| `logo.raw` | 57,400 bytes | The logo, 560 x 205 pixels, 4 bits per pixel | `logo` in [`rom/boot.s`](../../rom/boot.s) |
| `font.raw` | 512 bytes | 64 glyphs of 8 x 8 pixels | `font` in `rom/boot.s` |

The two files are build products.
They are not committed: CMake creates them in `<build>/rom` from the committed picture `rom/assets/playerone.png`.
The addresses of `logo` and `font` are written in the [ROM header](../console/memory-map.md#rom-header) at `+0x14` and `+0x18`.

## Usage

The script needs Python 3 and Pillow.

```bash
python3 rom/tools/make_assets.py rom/assets/playerone.png <outdir>
```

It creates `<outdir>` (and the parents) if needed and writes `logo.raw` and `font.raw` in it.
It exits with 2 and prints a usage line if the number of arguments is not 2.
It fails if the picture has no pixel darker than the crop threshold.

## Logo

The script converts the picture in these steps:

1. It opens the picture and converts it to 256 grey levels.
   The alpha channel is ignored.
2. It looks for the content: the pixels with a grey level below 215.
   The smallest box around them is the crop box.
3. It crops the picture to this box.
4. It resizes the result to 560 x 205 pixels with the Lanczos filter.
   The ratio is not preserved by the code, so the box must already have the ratio of the target.
5. It inverts each pixel and keeps 4 bits: `level = (255 - grey) >> 4`, from 0 to 15.
   White (the paper) is level 0 and black (the ink) is level 15.
6. It packs two pixels per byte.

### 4 bpp layout

| Property | Value |
|----------|-------|
| Pixels | 560 x 205 |
| Bits per pixel | 4, a level from 0 (background) to 15 (solid ink) |
| Bytes per row | 280 |
| Order | Row by row, the top row first, no padding between rows |
| Left pixel | High nibble (bits 7 to 4) |
| Right pixel | Low nibble (bits 3 to 0) |

The pixel `(x, y)` is in the byte `y * 280 + x / 2`.
It is in the high nibble if `x` is even, and in the low nibble if `x` is odd.

The ROM does not use the levels as grey values.
It maps each level to a colour that goes from the background to white, see [Logo fade-in](README.md#logo-fade-in).

## Font

The font is `font8x8_basic`, a public domain font by Daniel Hepper, based on the IBM VGA fonts.
The glyphs are written in the script itself, so the script does not read any font file.

| Property | Value |
|----------|-------|
| Glyphs | 64, for the characters `0x20` (space) to `0x5F` (underscore) |
| Glyph size | 8 bytes, one byte per row, the top row first |
| Bit order | Bit 0 is the leftmost pixel, bit 7 the rightmost |
| Set bit | The pixel is drawn |

The glyph of the character `c` starts at `font + (c - 0x20) * 8`.
For example `A` (`0x41`) is `0C 1E 33 33 3F 33 33 00`.

The font has digits, the upper case letters and the punctuation up to `_`.
It has no lower case letters: the text routine of the ROM does not check the range, so a character outside `0x20` to `0x5F` reads outside the font.
This is why all the texts of the ROM are in upper case.

The ROM draws a glyph at twice its size, 16 x 16 pixels, see [Message and blink](README.md#message-and-blink).

## Tests

The tests are in `rom/tools/test_make_assets.py`.

```bash
python3 -m pytest rom/tools
```

`python3 rom/tools/test_make_assets.py` also works, because the tests use `unittest`.

They check that:

- `logo.raw` is 57,400 bytes, its first byte is 0 (background) and at least 5% of the pixels are level 15;
- `font.raw` is 512 bytes, the space is blank and the glyph of `A` is the one above;
- the command line creates the missing directories and both files.

## See also

- [Boot ROM](README.md)
- [Memory map: ROM header](../console/memory-map.md#rom-header)
- [Building](../overview/building.md)
- [Documentation index](../README.md)
