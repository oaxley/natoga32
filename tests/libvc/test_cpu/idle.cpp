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
 * @brief	Tests - CPU idle state
 */

#include <catch2/catch_all.hpp>

#include "instruction_test.h"
#include "vc/events.h"

using namespace vc;

TEST_CASE("Idle CPU", "[cpu][threads][idle]") {
    InstructionTest test;

    SECTION("idle tick counts cycles") {
        test.setReg(10, HW_VBLANK);
        test.setReg(11, 0);
        test.loadInstruction(0x00B5400B);       // SLEEP.T a0, a1
        test.execute();
        REQUIRE(test.cpu.getState() == CPUState::Idle);

        u32 pc = test.getPC();
        u64 cycles = test.cpu.getTotalCycles();
        test.execute(10);

        REQUIRE(test.cpu.getTotalCycles() == cycles + 10);
        REQUIRE(test.getPC() == pc);
    }

    SECTION("event wakes and resumes") {
        test.setReg(10, HW_VBLANK);
        test.setReg(11, 0);
        test.loadInstruction(0x00B5400B);       // SLEEP.T a0, a1
        test.execute();
        REQUIRE(test.cpu.getState() == CPUState::Idle);

        test.cpu.wakeThreadOnEvent(HW_VBLANK);

        REQUIRE(test.cpu.getState() == CPUState::Running);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Running);
        REQUIRE(test.getCtx(0).getWaitKey() == 0);
    }

    SECTION("unrelated event") {
        test.setReg(10, HW_VBLANK);
        test.setReg(11, 0);
        test.loadInstruction(0x00B5400B);       // SLEEP.T a0, a1
        test.execute();
        REQUIRE(test.cpu.getState() == CPUState::Idle);

        test.cpu.wakeThreadOnEvent(HW_HBLANK);

        REQUIRE(test.cpu.getState() == CPUState::Idle);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Sleeping);
    }

    SECTION("timed sleep ends while idle") {
        test.setReg(10, 0);
        test.setReg(11, 5);
        test.loadInstruction(0x00B5400B);       // SLEEP.T a0, a1
        test.loadInstruction(0x00000013);       // NOP, run once awake
        test.execute();
        REQUIRE(test.cpu.getState() == CPUState::Idle);

        test.execute(5);

        REQUIRE(test.cpu.getState() == CPUState::Running);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Running);
    }
}

// T0 and T2 sleep until cycle 22, T1 (the last to run) sleeps on vblank:
// the CPU goes idle with current thread 1
static void idleWithTwoTimedSleepers(InstructionTest& test)
{
    test.setReg(3, MMAP_CODE + 0x100);
    test.setReg(4, MMAP_CODE + 0x200);
    test.setReg(11, 20);
    test.loadInstruction(0x0001810b);               // T0: NEW.T x2, x3    (T1)
    test.loadInstruction(0x0002010b);               //     NEW.T x2, x4    (T2)
    test.loadInstruction(0x00B0400B);               //     SLEEP.T zero, a1 (cycle 2 -> 22)
    test.loadInstruction(0x0000100b, 0x100, 1);     // T1: YIELD.T
    test.loadInstruction(0x0005400B, 0x100, 1);     //     SLEEP.T a0, zero (vblank)
    test.loadInstruction(0x00B0400B, 0x200, 2);     // T2: SLEEP.T zero, a1 (cycle 4 -> 22)

    test.execute(2);
    test.setReg(10, HW_VBLANK, 1);
    test.setReg(11, 18, 2);
    test.execute(4);

    REQUIRE(test.cpu.getState() == CPUState::Idle);
    REQUIRE(test.cpu.getCurrentThreadID() == 1);
    REQUIRE(test.cpu.getTotalCycles() == 6);
}

TEST_CASE("Idle CPU resumes round robin after the current thread", "[cpu][threads][idle]") {
    InstructionTest test;
    idleWithTwoTimedSleepers(test);

    SECTION("cycle by cycle") {
        test.execute(15);
        REQUIRE(test.cpu.getState() == CPUState::Idle);
        test.execute();
    }

    SECTION("fast-forward") {
        REQUIRE(test.cpu.skipIdle(100) == 16);
    }

    // both wake at cycle 22: T2 follows T1, T0 waits its turn
    REQUIRE(test.cpu.getTotalCycles() == 22);
    REQUIRE(test.cpu.getState() == CPUState::Running);
    REQUIRE(test.cpu.getCurrentThreadID() == 2);
    REQUIRE(test.getCtx(2).getState() == ThreadState::Running);
    REQUIRE(test.getCtx(0).getState() == ThreadState::Ready);
    REQUIRE(test.getCtx(1).getState() == ThreadState::Sleeping);
}

TEST_CASE("Idle CPU fast-forward", "[cpu][threads][idle]") {
    InstructionTest test;

    SECTION("not idle consumes nothing") {
        REQUIRE(test.cpu.getState() == CPUState::Running);
        REQUIRE(test.cpu.skipIdle(100) == 0);
        REQUIRE(test.cpu.getTotalCycles() == 0);
    }

    SECTION("event sleeper consumes the whole budget") {
        test.setReg(10, HW_VBLANK);
        test.setReg(11, 0);
        test.loadInstruction(0x00B5400B);       // SLEEP.T a0, a1
        test.execute();
        REQUIRE(test.cpu.getState() == CPUState::Idle);

        u64 cycles = test.cpu.getTotalCycles();
        REQUIRE(test.cpu.skipIdle(100) == 100);
        REQUIRE(test.cpu.getTotalCycles() == cycles + 100);
        REQUIRE(test.cpu.getState() == CPUState::Idle);
    }

    SECTION("timed sleeper stops the jump at its wake cycle") {
        test.setReg(10, 0);
        test.setReg(11, 5);
        test.loadInstruction(0x00B5400B);       // SLEEP.T a0, a1
        test.execute();
        REQUIRE(test.cpu.getState() == CPUState::Idle);

        // the sleep instruction already consumed one of the 5 cycles
        u64 cycles = test.cpu.getTotalCycles();
        REQUIRE(test.cpu.skipIdle(100) == 4);
        REQUIRE(test.cpu.getTotalCycles() == cycles + 4);
        REQUIRE(test.cpu.getState() == CPUState::Running);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Running);
    }

    SECTION("budget shorter than the sleep") {
        test.setReg(10, 0);
        test.setReg(11, 5);
        test.loadInstruction(0x00B5400B);       // SLEEP.T a0, a1
        test.execute();

        u64 cycles = test.cpu.getTotalCycles();
        REQUIRE(test.cpu.skipIdle(3) == 3);
        REQUIRE(test.cpu.getTotalCycles() == cycles + 3);
        REQUIRE(test.cpu.getState() == CPUState::Idle);

        REQUIRE(test.cpu.skipIdle(3) == 1);
        REQUIRE(test.cpu.getState() == CPUState::Running);
    }
}
