/* -*- coding: utf-8 -*-
 * vim: filetype=cpp
 *
 * This source file is subject to the MIT License
 * that is bundled with this package in the file LICENSE.txt.
 * It is also available through the Internet at this address:
 * https://opensource.org/license/mit
 *
 * @author	Sebastien LEGRAND
 * @license	MIT License
 *
 * @brief	Tests - Console frame clock, vblank and framebuffer
 */

#include <catch2/catch_all.hpp>
#include <vector>

#include "console.h"
#include "core/cpu.h"
#include "core/memory.h"
#include "core/constants.h"
#include "vc/events.h"

using namespace vc;

namespace
{

// 0x110-byte ROM: a loop sleeping until the vblank event
std::vector<u8> makeSleepLoopRom()
{
    std::vector<u8> rom(0x110, 0);
    const u32 code[] = {
        0x80000537,     // lui   a0, 0x80000
        0x00150513,     // addi  a0, a0, 1
        0x0005400B,     // sleep.t a0, zero
        0xFF5FF06F,     // jal   zero, -12
    };
    for (size_t i = 0; i < 4; ++i) {
        for (size_t b = 0; b < 4; ++b) {
            rom[0x100 + i * 4 + b] = static_cast<u8>(code[i] >> (8 * b));
        }
    }
    return rom;
}

// 0x110-byte ROM: a loop counting in a2 after each 1000-cycle timed sleep
std::vector<u8> makeTimedSleepRom()
{
    std::vector<u8> rom(0x110, 0);
    const u32 code[] = {
        0x3E800593,     // addi  a1, zero, 1000
        0x00B0400B,     // sleep.t zero, a1
        0x00160613,     // addi  a2, a2, 1
        0xFF5FF06F,     // jal   zero, -12
    };
    for (size_t i = 0; i < 4; ++i) {
        for (size_t b = 0; b < 4; ++b) {
            rom[0x100 + i * 4 + b] = static_cast<u8>(code[i] >> (8 * b));
        }
    }
    return rom;
}

// 0x164-byte ROM: thread 0 spawns 3 threads, each loop counts its wake-ups in s0
//   T0: sleeps on vblank          T1: timed sleep 1000 cycles
//   T2: timed sleep 777 cycles    T3: vblank or 2000 cycles, whichever comes first
std::vector<u8> makeMultiSleeperRom()
{
    const u32 code[] = {
        0x12C00293,     // 0x100 addi  t0, zero, 0x12C
        0x0002830B,     //       new.t t1, t0
        0x13C00293,     //       addi  t0, zero, 0x13C
        0x0002830B,     //       new.t t1, t0
        0x14C00293,     //       addi  t0, zero, 0x14C
        0x0002830B,     //       new.t t1, t0
        0x80000537,     // 0x118 lui   a0, 0x80000
        0x00150513,     //       addi  a0, a0, 1
        0x0005400B,     //       sleep.t a0, zero
        0x00140413,     //       addi  s0, s0, 1
        0xFF1FF06F,     //       j     0x118
        0x3E800593,     // 0x12C addi  a1, zero, 1000
        0x00B0400B,     //       sleep.t zero, a1
        0x00140413,     //       addi  s0, s0, 1
        0xFF5FF06F,     //       j     0x12C
        0x30900593,     // 0x13C addi  a1, zero, 777
        0x00B0400B,     //       sleep.t zero, a1
        0x00140413,     //       addi  s0, s0, 1
        0xFF5FF06F,     //       j     0x13C
        0x80000537,     // 0x14C lui   a0, 0x80000
        0x00150513,     //       addi  a0, a0, 1
        0x7D000593,     //       addi  a1, zero, 2000
        0x00B5400B,     //       sleep.t a0, a1
        0x00140413,     //       addi  s0, s0, 1
        0xFEDFF06F,     //       j     0x14C
    };
    std::vector<u8> rom(0x100 + sizeof(code), 0);
    for (size_t i = 0; i < std::size(code); ++i) {
        for (size_t b = 0; b < 4; ++b) {
            rom[0x100 + i * 4 + b] = static_cast<u8>(code[i] >> (8 * b));
        }
    }
    return rom;
}

// the two CPUs are in exactly the same state
void requireSameCPU(CPU& f, CPU& s)
{
    REQUIRE(f.getTotalCycles() == s.getTotalCycles());
    REQUIRE(f.getState() == s.getState());
    REQUIRE(f.getCurrentThreadID() == s.getCurrentThreadID());
    for (int tid = 0; tid < THREADS_COUNT; ++tid) {
        CAPTURE(tid);
        REQUIRE(f.getThread(tid).getState() == s.getThread(tid).getState());
        REQUIRE(f.getThread(tid).getPC() == s.getThread(tid).getPC());
        REQUIRE(f.getThread(tid).getSleepUntil() == s.getThread(tid).getSleepUntil());
        REQUIRE(f.getThread(tid).getWaitKey() == s.getThread(tid).getWaitKey());
        REQUIRE(f.getThread(tid).getTotalCycles() == s.getThread(tid).getTotalCycles());
        REQUIRE(f.getThread(tid).getRegisters()[8] == s.getThread(tid).getRegisters()[8]);
    }
}

void startConsole(Console& c, const std::vector<u8>& rom)
{
    REQUIRE(c.initialize());
    REQUIRE(c.loadROM(rom.data(), rom.size()));
    c.reset();
    c.resume();
}

} // namespace


TEST_CASE("Console frame", "[console]") {
    Console c;
    REQUIRE(c.initialize());

    auto rom = makeSleepLoopRom();
    REQUIRE(c.loadROM(rom.data(), rom.size()));
    c.reset();
    c.resume();

    SECTION("constants") {
        REQUIRE(CPU_FREQUENCY_HZ == 25'000'000);
        REQUIRE(FRAMES_PER_SECOND == 60);
        REQUIRE(CYCLES_PER_FRAME == 416666);
        REQUIRE(VIDEO_WIDTH == 640);
        REQUIRE(VIDEO_HEIGHT == 384);
        REQUIRE(MMAP_FRAMEBUFFER_A == MMAP_VIDEO_RAM);
    }

    SECTION("a frame lasts CYCLES_PER_FRAME cycles and ends with vblank") {
        c.runFrame();

        REQUIRE(c.getCPU().getTotalCycles() == CYCLES_PER_FRAME);
        REQUIRE(c.getCPU().getThread(0).getState() != ThreadState::Sleeping);
    }

    SECTION("framebuffer maps the video RAM") {
        c.getMemory().write16(MMAP_VIDEO_RAM + 2 * (640 * 1 + 2), 0xABCD);

        REQUIRE(c.getFramebuffer() != nullptr);
        REQUIRE(c.getFramebuffer()[640 + 2] == 0xABCD);
    }

    SECTION("reset after a frame restarts at the reset address") {
        c.runFrame();
        c.reset();
        c.resume();

        REQUIRE(c.getCPU().getThread(0).getPC() == RESET_DEFAULT_ADDR);
        REQUIRE(c.getCPU().getState() == CPUState::Running);
    }
}


TEST_CASE("Console idle fast-forward matches cycle-by-cycle execution", "[console][idle]") {
    auto rom = makeTimedSleepRom();
    auto budget = GENERATE(as<u32>{}, 1u, 2u, 3u, 999u, 1000u, 1001u, 1002u, 1003u, 1004u, 5000u, 12345u, CYCLES_PER_FRAME);

    Console fast;
    Console slow;
    startConsole(fast, rom);
    startConsole(slow, rom);

    fast.run(budget);
    for (u32 i = 0; i < budget; ++i) {
        slow.tick();
    }

    CPU& f = fast.getCPU();
    CPU& s = slow.getCPU();
    REQUIRE(f.getTotalCycles() == budget);
    REQUIRE(f.getTotalCycles() == s.getTotalCycles());
    REQUIRE(f.getState() == s.getState());
    REQUIRE(f.getThread(0).getState() == s.getThread(0).getState());
    REQUIRE(f.getThread(0).getPC() == s.getThread(0).getPC());
    REQUIRE(f.getThread(0).getSleepUntil() == s.getThread(0).getSleepUntil());
    REQUIRE(f.getThread(0).getRegisters()[12] == s.getThread(0).getRegisters()[12]);
}

TEST_CASE("Console idle fast-forward matches cycle-by-cycle execution with several sleepers", "[console][idle]") {
    auto rom = makeMultiSleeperRom();
    // budgets around the wake-ups of T1 (1000), T2 (777) and T3 (2000)
    auto budget = GENERATE(as<u32>{}, 1u, 20u, 776u, 777u, 790u, 1000u, 1010u, 1554u, 1999u, 2000u, 2020u, 7777u, 12345u);

    Console fast;
    Console slow;
    startConsole(fast, rom);
    startConsole(slow, rom);

    // part of a frame, a vblank event (wakes T0 and T3), then the rest of the frame
    fast.run(budget);
    fast.triggerEvent(HW_VBLANK);
    fast.run(CYCLES_PER_FRAME - budget);
    for (u32 i = 0; i < budget; ++i) {
        slow.tick();
    }
    slow.triggerEvent(HW_VBLANK);
    for (u32 i = budget; i < CYCLES_PER_FRAME; ++i) {
        slow.tick();
    }

    REQUIRE(fast.getCPU().getTotalCycles() == CYCLES_PER_FRAME);
    REQUIRE(fast.getState() == ConsoleState::Running);
    requireSameCPU(fast.getCPU(), slow.getCPU());

    // every thread woke up at least once (T0 only when it slept before the vblank)
    for (int tid = (budget < 20 ? 1 : 0); tid < 4; ++tid) {
        CAPTURE(tid);
        REQUIRE(fast.getCPU().getThread(tid).getRegisters()[8] > 0);
    }
}

TEST_CASE("Console stops in error when the CPU traps without a handler", "[console]") {
    // the ROM executes an illegal instruction without setting mtvec
    std::vector<u8> rom(0x104, 0);

    Console c;
    startConsole(c, rom);
    c.runFrame();

    REQUIRE(c.getCPU().getState() == CPUState::Halted);
    REQUIRE(c.getState() == ConsoleState::Error);
    REQUIRE(c.getLastError() == "CPU halted due to exception");
    REQUIRE(c.getCPU().getTotalCycles() == 1);
}

TEST_CASE("Console total cycles is the CPU global counter", "[console]") {
    Console c;
    startConsole(c, makeTimedSleepRom());

    // the thread executes a few instructions then sleeps: idle cycles count too
    c.run(5000);

    REQUIRE(c.getCPU().getTotalCycles() == 5000);
    REQUIRE(c.getTotalCycles() == 5000);
}

TEST_CASE("Console timed sleeper frame", "[console][idle]") {
    Console c;
    startConsole(c, makeTimedSleepRom());

    c.runFrame();

    // a loop lasts 4 instructions + 999 idle cycles, the first increment lands on cycle 1002
    REQUIRE(c.getCPU().getTotalCycles() == CYCLES_PER_FRAME);
    REQUIRE(c.getCPU().getThread(0).getRegisters()[12] == (CYCLES_PER_FRAME + 1) / 1003);
}
