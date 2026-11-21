# vchost

This page describes `vchost`, the host program: its command line, the window, the keys, the headless mode, the frame pacing and the exit codes.

## Overview

`vchost` (version 0.1.0) is the only host of the console today.
It is a C++20 program in `core/host` that uses [`libvc`](../console/libvc-api.md) for the console and SDL3 for the window.

At start-up it does these steps:

1. It creates and initializes a console.
2. It loads the boot ROM (`--rombios`, or `rombios.sys` next to the executable).
3. It loads the cartridge, if a file is given.
4. It resets the console and resumes it.
5. It runs the frames, in a window or without one.

Without a cartridge, the console shows the [boot ROM](../rom/README.md) screen.
The build puts `rombios.sys` next to `vchost`, see [Building](../overview/building.md#build-outputs).

## Command line

```text
vchost [--help] [--version] [--debug] [--debug-port PORT] [--rombios FILENAME] [--superscalar] [--frames N] [file]
```

| Option | Default | Meaning |
|--------|---------|---------|
| `file` | none | Cartridge file (`.n32`). At most one. See [Cartridge](#cartridge) |
| `--rombios FILENAME` | `rombios.sys` in the directory of the executable | Boot ROM file |
| `--superscalar` | off | Opens a 1280 x 768 window: the display is scaled x2 |
| `--frames N` | `0` | `N > 0`: run N frames without a window, then exit. `0`: open a window. `N < 0` is an error |
| `--debug` | off | Accepted, no effect yet |
| `--debug-port PORT` | `2600` | Accepted, no effect yet |
| `-h`, `--help` | | Prints the help and exits |
| `-v`, `--version` | | Prints the version and exits |

The options and the file can come in any order.
A bad command line (a value that is not an integer, a second file) prints the error and the usage, and exits with 1.

The default of `--rombios` is the directory where the executable is, wherever `vchost` is started from.
`--help` prints that directory in full.

`--debug` and `--debug-port` are read and stored, but nothing uses them: the host starts no debug server.

### Examples

```bash
build/core/host/vchost                       # boot screen in a window
build/core/host/vchost game.n32              # with a cartridge
build/core/host/vchost --superscalar         # 1280 x 768 window
build/core/host/vchost --frames 60           # headless, 60 frames
build/core/host/vchost --rombios my.sys      # another boot ROM
```

## Cartridge

`vchost` gives the file to `vc_loadCartridgeFile` before the first reset.
The call fails only if the file cannot be read or is empty: the host then prints `Failed to load cartridge: <path>: <reason>` and exits with 1.

A file that is not a valid cartridge is accepted.
The cartridge device reports it in its status register, and the console runs as usual.
See [Cartridge insertion](../console/cartridge/insertion.md#no-cartridge-or-an-invalid-one).

The boot ROM does not read the cartridge yet: the screen is the same with or without one.

## Window

The windowed mode starts when `--frames` is 0 (the default).

| Property | Value |
|----------|-------|
| Title | `NATOGA32` |
| Size | 640 x 384, or 1280 x 768 with `--superscalar` |
| Content | [Framebuffer A](../console/memory-map.md#framebuffer-a), RGB565, read with `vc_getFramebuffer` |
| Scaling | Nearest neighbour, over the whole window |

One loop turn does these steps:

1. It reads the pending events (keys, close button).
2. It runs one console frame with `vc_runFrame`.
3. It copies the framebuffer into an SDL texture and shows it.
4. It waits for the next frame, see [Frame pacing](#frame-pacing).

The window has no other option: it is not resizable by the code, and `--superscalar` is a host option only.
The console does not know about the scale.

### Keys

| Key | Action |
|-----|--------|
| `R` | Reset the console and resume it. The key repeat is ignored |
| `Esc` | Quit |

The close button of the window quits too.
The reset applies the pending cartridge change as any `vc_reset` does, see [libvc C API](../console/libvc-api.md#lifecycle).
`vchost` never changes the cartridge after start-up, so a reset checks the same file again.

## Headless mode

`--frames N` with `N > 0` runs without a window.
The code makes no SDL call in this mode, so it works without a display.

The host calls `vc_runFrame` N times, then checks the console.
The exit code is 0 if the console is not in the Error state, and 1 otherwise.
For example, a ROM that traps without a handler gives `Console error: CPU halted due to exception`.
The console state and the errors are described in [libvc C API](../console/libvc-api.md#state-and-errors).

The headless mode does not wait: it runs as fast as the machine allows.
`--superscalar` has no effect with `--frames`.

In a window, the host does not check the state of the console after a frame.
If the console goes to the Error state, the window stays open and no message is printed.

## Frame pacing

The console runs 60 frames per second (416,666 cycles of the 25 MHz clock per frame).
Only the windowed mode paces the frames.

| Case | What paces the frames |
|------|-----------------------|
| VSync available | `SDL_SetRenderVSync(renderer, 1)` worked: the display refresh paces the loop, one console frame per refresh |
| VSync unavailable | A timer at 60 fps, from `FramePacer` |

In the first case the host does not check the refresh rate of the display.
A display that is not at 60 Hz changes the speed of the console.

In the second case the host prints `VSync unavailable (<reason>), using a timer at 60 fps` and waits with `SDL_DelayPrecise`.
`FramePacer` (`core/host/src/frame_pacer.h`) has no SDL call and works in nanoseconds:

- the frame N ends at `start + N / 60` seconds, computed from the start, so the timer does not drift;
- a frame a little late is caught up by the next ones;
- if a frame is late by more than one frame (a stall), the cadence restarts from now, instead of running frames back to back.

## Exit codes

| Code | Case |
|------|------|
| 0 | Normal exit: the window was closed or `Esc` was pressed, or a headless run ended without a console error. `--help` and `--version` too |
| 1 | The console cannot start, the ROM or the cartridge cannot be loaded, `--frames` is negative, the command line is bad, an SDL call fails, or a headless run ends in the Error state |

The error messages go to the standard error.

## Tests

`ctest` runs `vchost` headless: a missing ROM, a negative `--frames`, a missing cartridge, a ROM that traps, an invalid cartridge (the boot ROM still starts), the default ROM found through `PATH`, and 5 frames with the boot ROM, with and without `--superscalar`.
The frame pacer has its own unit tests in `tests/host/test_frame_pacer.cpp`.
See [Building](../overview/building.md#tests).

## See also

- [Boot ROM](../rom/README.md)
- [libvc C API](../console/libvc-api.md)
- [Cartridge insertion](../console/cartridge/insertion.md)
- [Architecture](../overview/architecture.md)
- [Building](../overview/building.md)
