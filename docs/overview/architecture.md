# Architecture

This page shows the parts of NATOGA32, what each one does, and how a frame and a cartridge go through them.

## The parts

| Part | Directory | Language | Role |
|------|-----------|----------|------|
| `libvc` | `core/console` | C++20 | The emulated console: CPU, memory, cartridge device. Static library, shared with `VC_BUILD_SHARED`. |
| `vchost` | `core/host` | C++20, SDL3 | The host: window or headless run, loads the ROM and the cartridge, runs the frames. |
| boot ROM | `rom/` | assembly | `boot.s` is assembled by `n32asm` into `rombios.sys`. It shows the boot screen. |
| `n32asm` | `tools/n32asm` | Go | Assembler and disassembler for the NATOGA32 RISC-V. |
| `n32pack` | `tools/n32pack` | Go | Packs the chunks of `n32asm` and the assets into a `.n32` cartridge. |
| tests | `tests/` | C++ (Catch2) | `vctests`, the cartridge checker `vc_cartcheck` and the `ctest` cases of `vchost`. |

The Go tools have their own tests next to their sources. The ROM asset script has its tests in `rom/tools`.

```
                 +--------------------+
                 |       vchost       |   SDL3 window, or --frames N
                 +---------+----------+
                           | C API (vc.h)
                 +---------v----------+
                 |        libvc       |   CPU + 8 threads, memory, cartridge device
                 +--+--------------+--+
       rombios.sys  |              |  game.n32
        (boot ROM)  |              |  (cartridge)
```

`libvc` knows nothing about `vchost`. `vchost` only uses `vc.h`.
The tests link `libvc` directly and use its internal headers.

The `vcdebug` target (`core/debugger`) is a placeholder: its `main` is empty and it is not part of the built console.

## One frame

The host owns the loop. In a window, one turn of the loop is one frame:

```
vchost                          libvc
------                          -----
poll the window events
vc_runFrame(console) ---------> run CYCLES_PER_FRAME cycles (416,666)
                                  CPU runs the active thread,
                                  or skips time when it is idle
                                trigger HW_VBLANK (0x8000_0001)
                                  threads that sleep on it wake up
vc_getFramebuffer(console) <--- pointer to Framebuffer A
copy 640x384 RGB565 to the SDL texture, present
wait for vsync (or the 60 fps timer)
```

In headless mode (`--frames N`), the host only calls `vc_runFrame` `N` times and checks the console state at the end.

The boot ROM is a good example of the loop: it draws into Framebuffer A, then sleeps on `HW_VBLANK` with `sleep.t` until the next frame.

## The cartridge pipeline

A cartridge is built outside the console, with the two Go tools:

```
game.s ----------------+
                       |  n32asm game.s -f rscv -o game.rscv --sym game.sym
                       v
              game.rscv   game.sym
               (RSCV)     (dSYM)
                  |          |
 game.json -------+----------+--- assets (tiles, ...)
 (manifest)                  |
                             |  n32pack -o game.n32 game.json
                             v
                         game.n32
                             |
                             |  vchost game.n32
                             v
                    libvc: vc_loadCartridgeFile
                    checked at the next vc_reset
                    a program reads the CART_STATUS register
```

If the game source includes `assets.inc`, run `n32pack -inc assets.inc game.json` first: it writes the asset ids that the source uses.
The workflow is described in [n32pack](../tools/n32pack/README.md).

`vc_loadCartridgeFile` only stores the file. The cartridge device checks it at the next `vc_reset()`.
An invalid file does not stop the console: the status register of the device tells what is wrong.
The boot ROM does not read the cartridge yet: it always shows its "insert a cartridge" message.

## See also

- [Design principles](principles.md)
- [Building](building.md)
- [Console](../console/README.md)
- [libvc C API](../console/libvc-api.md)
- [Cartridge](../console/cartridge/README.md)
- [Boot ROM](../rom/README.md)
- [vchost](../host/vchost.md)
- [n32asm](../tools/n32asm/README.md)
- [n32pack](../tools/n32pack/README.md)
- [Documentation index](../README.md)
