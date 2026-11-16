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

TEST_CASE("Console timed sleeper frame", "[console][idle]") {
    Console c;
    startConsole(c, makeTimedSleepRom());

    c.runFrame();

    // a loop lasts 4 instructions + 999 idle cycles, the first increment lands on cycle 1002
    REQUIRE(c.getCPU().getTotalCycles() == CYCLES_PER_FRAME);
    REQUIRE(c.getCPU().getThread(0).getRegisters()[12] == (CYCLES_PER_FRAME + 1) / 1003);
}
