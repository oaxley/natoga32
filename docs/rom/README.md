# Boot ROM

This page describes what the boot ROM does today, how it is built, and what its trap screen shows.

## Overview

The boot ROM is a small program written in assembly: [`rom/boot.s`](../../rom/boot.s).
`n32asm` assembles it into `rombios.sys`, and `vchost` loads that file at `0x0000_0000`.
The console starts thread 0 at `0x0000_0100`, the entry point of the ROM.

The ROM draws everything itself in [Framebuffer A](../console/memory-map.md#framebuffer-a).
The console has no splash screen of its own.

| Property | Value |
|----------|-------|
| Source | `rom/boot.s` |
| Output | `rombios.sys`, assembled with `n32asm -f rom` |
| Entry point | `0x0000_0100` |
| Threads used | Thread 0 only |
| Memory written | Framebuffer A, and about 1.1 KiB at the start of the heap (`0x0080_0000`) |
| Cartridge | Not read yet |

The first 256 bytes of the ROM are the ROM header.
Its layout is in the [memory map](../console/memory-map.md#rom-header).
The ROM fills it with the magic `N32R`, version 0.1, the entry point, the address of the trap handler, no syscall table, and the addresses of the logo and of the font.

## What the ROM does

The ROM runs these steps, in this order:

1. It writes the address of its trap handler in `mtvec`.
   The reset clears all the CSRs, so the ROM does it at every boot. See [Traps](../console/cpu/traps.md#trap-without-a-handler).
2. It fills the 384 rows of Framebuffer A with the navy background.
3. It fades the logo in (see [Logo fade-in](#logo-fade-in)).
4. It shows the message "INSERT A CARTRIDGE AND PRESS RESET" and blinks it forever.

The ROM never ends and never reads the cartridge.
A valid cartridge changes nothing on the screen: the [cartridge device](../console/cartridge/insertion.md) reports its status in registers that the ROM does not read.
To start again, reset the console (the R key of [vchost](../host/vchost.md#keys)).

## Screen layout

| Item | Position (pixels) | Size | Colour (RGB565) |
|------|-------------------|------|-----------------|
| Background | whole screen | 640 x 384 | navy `(8, 10, 24)`, rounded to `0x0843` |
| Logo | `(40, 60)` | 560 x 205 | fades from the background to white `0xFFFF` |
| Message | `(48, 310)` | 34 characters of 16 x 16 = 544 x 16 | `0xCE5A`, a light grey |

`0x0843` is red 1, green 2, blue 3 in 5-6-5 bits.
The message is centred: `(640 - 544) / 2 = 48`.
The logo and the font are converted at build time, see [ROM assets](assets.md).

## Logo fade-in

The logo has 16 grey levels (0 to 15) and 4 bits per pixel.
The ROM turns a level into a colour with a table of 16 colours, and builds a new table at each step.

- There are 15 steps, numbered 1 to 15.
- At the end of each step the ROM waits for 3 frames, so the fade lasts 45 frames (0.75 s at 60 fps).
- At step `s`, level `v` has this colour, with integer divisions:
  `R5 = 1 + 30 * v * s / 225`, `G6 = 2 + 61 * v * s / 225`, `B5 = 3 + 28 * v * s / 225`.
- Level 0 is always the background colour `(1, 2, 3)` in 5-6-5 bits, which is `0x0843`.
- Level 15 at step 15 is `(31, 63, 31)`, which is white.

For each step the ROM also builds a table of 256 words, one for each byte of the logo.
A word holds the two pixels of the byte, so the logo is drawn with one table lookup per 2 pixels.

The tables live at the start of the heap:

| Address | Size | Content |
|---------|------|---------|
| `0x0080_0000` | 32 bytes | The 16 colours of the levels, 16 bits each |
| `0x0080_0040` | 1024 bytes | The 256 words, two pixels each |
| `0x0080_0440` | 9 bytes | 8 hex digits and a 0 byte, used by the trap screen |

## Message and blink

The message is drawn with the 8x8 font, each pixel doubled to 2 x 2 pixels, so a glyph is 16 x 16 pixels.
The text routine writes only the pixels that are set.
To hide the text, the ROM draws it again in the background colour.

The blink loop shows the message for 30 frames, hides it for 30 frames, and repeats.
The period is 60 frames, one second.

## Waiting for a frame

The ROM never loops on a timer.
To wait for a frame it runs `sleep.t` with the waitkey `HW_VBLANK` (`0x8000_0001`) and no duration:

```asm
wait_vblank:
    li a0, WAITKEY_VBLANK
    sleep.t a0, zero
    ret
```

The CPU is Idle until the host triggers the vblank at the end of the frame.
See [`sleep.t`](../console/cpu/instructions.md#sleept) and [Waitkey values](../console/cpu/instructions.md#waitkey-values).

## Trap screen

If the ROM itself traps, for example on an illegal instruction, the CPU jumps to the trap handler.
The handler:

1. sets `sp` to `0x0079_0000`, the top of the stack of thread 0;
2. fills the first 64 rows with red (`0xF800`);
3. writes "SYSTEM ERROR" in white at `(16, 16)`;
4. writes `mcause` as 8 hexadecimal digits (upper case) in white at `(16, 40)`;
5. waits for the vblank in a loop, forever.

The rest of the screen is not cleared, so the background, the logo and the message stay visible under the red band.
The causes are listed in [Traps](../console/cpu/traps.md#causes).
For example, an illegal instruction shows `00000002`.

## Build

The CMake target `rombios` builds the ROM.
It needs `n32asm` and Python 3 with Pillow. Without them, CMake prints a warning and the ROM is not built.

| Step | Command | Output |
|------|---------|--------|
| 1 | `python3 rom/tools/make_assets.py rom/assets/playerone.png <build>/rom` | `logo.raw`, `font.raw` |
| 2 | copy of `rom/boot.s` to `<build>/rom/boot.s` | the source next to the assets |
| 3 | `n32asm <build>/rom/boot.s -f rom -o <build>/core/host/rombios.sys` | `rombios.sys` |

The source is copied next to the assets because the `.incbin "logo.raw"` and `.incbin "font.raw"` paths are relative to the source file.
`rombios.sys` ends up next to `vchost`, which looks for it there by default.
See [Building](../overview/building.md#build-outputs) for the output paths, and [ROM assets](assets.md) for step 1.

To build it by hand, from a directory that holds `boot.s`, `logo.raw` and `font.raw`:

```bash
n32asm boot.s -f rom -o rombios.sys
```

The `-f rom` output writes the `.rom` section as a flat image from `0x0000_0000`.

## Tests

The Catch2 test `Boot ROM` (`tests/libvc/test_bootrom.cpp`) loads `rombios.sys`, runs 300 frames and checks the header, the background, a logo pixel at full white and the blink.
A second part patches the instruction at `0x0000_010C` to `0` (an illegal instruction), runs 2 frames and checks the trap screen.
The test is skipped when the ROM is not built.

## Not built

The ROM has no game loader and no syscall table (the header field is 0).

## See also

- [ROM assets](assets.md)
- [vchost](../host/vchost.md)
- [Memory map: ROM and ROM header](../console/memory-map.md#rom)
- [Traps](../console/cpu/traps.md)
- [Thread instructions](../console/cpu/instructions.md)
- [Cartridge insertion](../console/cartridge/insertion.md)
- [Building](../overview/building.md)
