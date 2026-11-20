# Building

This page lists what you need to build NATOGA32, the CMake options, the commands, the files they produce, and how to run the tests.

## Requirements

| Tool | Version | Needed for |
|------|---------|------------|
| C++ compiler | C++20 | `libvc`, `vchost`, tests |
| CMake | 3.25 or newer | the whole build |
| Go | 1.21 or newer (`go.mod` of `goasm` and `n32pack`) | `goasm`, `n32pack` |
| Python 3 with Pillow (`PIL`) | no minimum in the CMake files | the boot ROM assets (`rom/tools/make_assets.py`) |
| SDL3 | found by `find_package(SDL3)` | `vchost` |

CMake downloads two more libraries at configure time with `FetchContent`: Catch2 (tests) and argparse (`vchost`).
This needs network access the first time.

Some parts are skipped with a warning when a requirement is missing:

- No Go: `goasm`, `n32pack`, the boot ROM and the sample cartridge are not built.
- No Python 3 with Pillow: the boot ROM (`rombios.sys`) is not built.
- No SDL3: the configuration fails when `VC_BUILD_HOST` is `ON`. Use `-DVC_BUILD_HOST=OFF`.

## CMake options

| Option | Default | Effect |
|--------|---------|--------|
| `VC_BUILD_DEBUGGER` | `ON` | Adds `core/debugger`. It needs an `imgui` target from `third_party/`, and without it CMake prints a warning and skips it. The debugger is a placeholder, not documented here. |
| `VC_BUILD_TESTS` | `ON` | Builds `vctests`, `vc_cartcheck` and registers the `ctest` cases. |
| `VC_BUILD_HOST` | `ON` | Builds `vchost`. Needs SDL3. |
| `VC_BUILD_SHARED` | `OFF` | Builds `libvc` as a shared library instead of a static one. |

The compiler flags are `-Wall -Wextra -pedantic` (`/W4` with MSVC).

## Build

From the root of the repository:

```bash
cmake -S . -B build
cmake --build build -j
```

Add `-DCMAKE_BUILD_TYPE=Release` to the first command for an optimized build.
To build without the window host and the tests, for example:

```bash
cmake -S . -B build -DVC_BUILD_HOST=OFF -DVC_BUILD_TESTS=OFF
```

## Build outputs

All paths are relative to the build directory (`build/`).

| Output | Path | What it is |
|--------|------|------------|
| `vchost` | `core/host/vchost` | The host program |
| `rombios.sys` | `core/host/rombios.sys` | The boot ROM, next to `vchost` because the host looks for it there |
| `goasm` | `goasm` | The assembler |
| `n32pack` | `n32pack` | The cartridge packer |
| `vctests` | `tests/vctests` | The Catch2 test program |
| `vc_cartcheck` | `tests/vc_cartcheck` | Loads a cartridge and prints its status |
| `hello.n32` | `tests/cartridge/hello.n32` | A sample cartridge, packed and checked by the tests |
| `libvc` | `core/console/libvc.a` | The console library (`libvc.so` with `VC_BUILD_SHARED`) |

The Go tools are built by `go build` through custom targets (`vc_goasm`, `vc_n32pack`).
They are rebuilt when a Go source changes.

## Run

```bash
build/core/host/vchost                    # boot screen in a window
build/core/host/vchost game.n32           # with a cartridge
build/core/host/vchost --frames 60        # headless: run 60 frames, then exit
```

See [vchost](../host/vchost.md) for the options.

## Tests

```bash
ctest --test-dir build                    # vctests, cartridge checks, vchost headless runs
(cd tools/goasm && go test ./...)         # goasm tests
(cd tools/n32pack && go test ./...)       # n32pack tests
python3 -m pytest rom/tools               # boot ROM asset script
```

`ctest` runs the cases of `vctests` and the extra cases registered in `tests/CMakeLists.txt`.
You can also start `build/tests/vctests` directly and pass Catch2 filters.

## See also

- [Architecture](architecture.md)
- [Design principles](principles.md)
- [vchost](../host/vchost.md)
- [goasm](../tools/goasm/README.md)
- [n32pack](../tools/n32pack/README.md)
- [Boot ROM](../rom/README.md)
- [Documentation index](../README.md)
