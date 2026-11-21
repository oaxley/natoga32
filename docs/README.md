# NATOGA32 documentation

This is the entry point of the NATOGA32 documentation: it says what the project is, what works today, and links to every page.

## What is NATOGA32?

NATOGA32 is a portmanteau built from the sentence "Not an Atari nor an Amiga, but a bit of both".
It is a fantasy console emulator, inspired by the hardware of the 1990s and the early 2000s, with a couple of twists.
The console is designed to be fun to program, yet challenging.

The project has four parts:

- `libvc`: the emulated console, a C++20 library with a C API (`core/console`).
- `vchost`: an SDL3 program that runs the console in a window (`core/host`).
- a boot ROM, written in assembly (`rom/`).
- two Go tools, `n32asm` (assembler) and `n32pack` (cartridge packer) (`tools/`).

## What works today

- A 32-bit RISC-V CPU (RV32I and M, Zba, Zbb, Zbs, Zicond and part of Zbkb) with 8 cooperative threads, CSR registers and traps.
- A 32 MiB memory map with a private zone for the thread state.
- A cartridge device: the console checks a `.n32` file at reset, reads its chunk index and exposes it through a control block and a bank-switched window.
- A boot ROM that draws a logo with a fade-in and an "insert a cartridge" message in Framebuffer A (640x384, RGB565), and a trap screen.
- `vchost`: window or headless run at 25 MHz and 60 fps, with the boot ROM and a cartridge given on the command line.
- `n32asm`: assembler and disassembler with a preprocessor, and the `rscv` and `rom` outputs.
- `n32pack`: builds a `.n32` cartridge from a JSON manifest, the chunks of `n32asm` and asset files.

Everything else (video modes, GPU, inputs, audio, DMA, DSP, memory card) is not built and is not documented here.

## Pages

### Overview

| Page | Content |
|------|---------|
| [Design principles](overview/principles.md) | The rules the code applies: host and console, clock, cycles, determinism |
| [Architecture](overview/architecture.md) | The parts of the project, one frame, the cartridge pipeline |
| [Building](overview/building.md) | Requirements, CMake options, build outputs, tests |

### Console (`libvc`)

| Page | Content |
|------|---------|
| [Console](console/README.md) | What the emulated console contains |
| [Memory map](console/memory-map.md) | Zones, addresses and sizes of the 32 MiB address space |
| [libvc C API](console/libvc-api.md) | The functions a host calls |

### CPU

| Page | Content |
|------|---------|
| [CPU](console/cpu/README.md) | The instruction set, the cycle model and the idle state |
| [Threads](console/cpu/threads.md) | The 8 cooperative threads, their states, control blocks and stacks |
| [Instructions](console/cpu/instructions.md) | The six thread instructions and the event codes of `sleep.t` and `wake.t` |
| [CSR registers](console/cpu/csr.md) | The control and status registers |
| [Traps](console/cpu/traps.md) | How a trap is taken, its causes, `ecall` and `mret` |

### Cartridge

| Page | Content |
|------|---------|
| [Cartridge](console/cartridge/README.md) | What a cartridge is and how the parts fit |
| [File format](console/cartridge/file-format.md) | The `.n32` layout: header, index, end marker |
| [Chunks](console/cartridge/chunks.md) | Each chunk type and its fields |
| [Console view](console/cartridge/console-view.md) | The control registers, the status codes, the 128 KiB window and bank switching |
| [Insertion](console/cartridge/insertion.md) | Load, eject, reset and the status the console reports |

### Boot ROM

| Page | Content |
|------|---------|
| [Boot ROM](rom/README.md) | The ROM header, the start-up code, the logo fade-in and the trap screen |
| [ROM assets](rom/assets.md) | How the logo and the font are converted and included |

### Host

| Page | Content |
|------|---------|
| [vchost](host/vchost.md) | The host program: command line, window, keys, headless mode, frame pacing, exit codes |

### Tools

| Page | Content |
|------|---------|
| [n32asm](tools/n32asm/README.md) | The assembler: build, command-line options and disassembler |
| [n32asm syntax](tools/n32asm/syntax.md) | Sections, labels, registers, numbers and data directives |
| [n32asm preprocessor](tools/n32asm/preprocessor.md) | Constants, includes, conditions, loops, macros |
| [n32asm instructions](tools/n32asm/instructions.md) | Supported instructions and pseudo-instructions |
| [n32asm output formats](tools/n32asm/output-formats.md) | The `rscv` and `rom` outputs, the symbol file and the optimizer |
| [n32asm quick reference](tools/n32asm/quick-reference.md) | One-screen summary |
| [n32pack](tools/n32pack/README.md) | The cartridge packer: build, options, place in the cartridge build, checks |
| [n32pack manifest](tools/n32pack/manifest.md) | The JSON manifest and the `assets.inc` file |

## Conventions

- All values are little-endian.
- Addresses are written `0x0020_6800`.
- Paths are relative to the root of the repository.

## See also

- [Design principles](overview/principles.md)
- [Architecture](overview/architecture.md)
- [Building](overview/building.md)
