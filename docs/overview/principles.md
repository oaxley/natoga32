# Design principles

This page lists the rules that the code of NATOGA32 applies today, and why they matter.

## The host drives the console

The console does not run by itself. A host program creates it, loads the ROM and the cartridge, and asks it to run.
`vchost` is the host that exists today. It calls `vc_runFrame` once per frame, then reads the framebuffer.

The console never knows that the host exists. `libvc` has no window, no SDL call and no clock of the real world.
It only has a C API (`core/console/include/vc/vc.h`) that the host calls.

This separation is strict. Everything the host needs goes through the API, and nothing flows the other way.

## Headless mode

Because the console does not depend on the host, it can run without a screen.
`vchost --frames N` runs `N` frames, opens no window (no `SDL_Init`), and exits. The tests use it, and so can a script or a CI job.

In headless mode, if the CPU stops in error, the host prints `Console error: ...` and exits with a failure status.

## The clock

The global clock is 25 MHz and the console runs at 60 frames per second. The values are in `core/console/src/core/constants.h`:

| Constant | Value |
|----------|-------|
| `CPU_FREQUENCY_HZ` | 25,000,000 |
| `FRAMES_PER_SECOND` | 60 |
| `CYCLES_PER_FRAME` | 416,666 (`CPU_FREQUENCY_HZ / FRAMES_PER_SECOND`, integer division) |

A frame is `CYCLES_PER_FRAME` cycles of the CPU, followed by the `HW_VBLANK` event (`0x8000_0001`).
The vblank event wakes the threads that sleep on it.

## Cycles drive everything

- The CPU executes an instruction and consumes cycles. Today every instruction costs one cycle.
- Time inside the console is counted in cycles, never in seconds. `sleep.t` durations are in cycles too.
- When all the threads sleep, the CPU is idle. The cycles still pass, and the console skips the idle stretch in one jump instead of looping one cycle at a time.
- The vblank event is raised from the cycle count, not from the clock of the host.
- The only device, the cartridge, has no time of its own. The vblank is the only time-driven event.

The host decides how to show the frames: with vsync, or with a timer at 60 fps.
This never changes what the console computes, only when the host asks for the next frame.

## Determinism

The same ROM, the same cartridge and the same number of frames give the same result.
The console reads no real clock and uses no random source. Its time is the cycle counter, so a run can be replayed.

## See also

- [Architecture](architecture.md)
- [Building](building.md)
- [libvc C API](../console/libvc-api.md)
- [vchost](../host/vchost.md)
- [CPU](../console/cpu/README.md)
- [Documentation index](../README.md)
