# NATOGA32

NATOGA32 ("Not an Atari nor an Amiga, but a bit of both") is a fantasy console emulator, inspired by the hardware of the 1990s and the early 2000s.
The console has a 32-bit RISC-V CPU with 8 cooperative threads and a 32 MiB memory map.
The project holds the emulator library `libvc` (C++20), the SDL3 host `vchost`, a boot ROM in assembly, and two Go tools: the assembler `goasm` and the cartridge packer `n32pack`.

## Quick build

```bash
cmake -S . -B build
cmake --build build -j
build/core/host/vchost
```

See [docs/overview/building.md](docs/overview/building.md) for the requirements and the CMake options.

## Documentation

[docs/README.md](docs/README.md)

## License

MIT, see [LICENSE](LICENSE).
