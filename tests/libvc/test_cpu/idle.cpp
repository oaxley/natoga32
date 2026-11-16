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
        test.execute();
        REQUIRE(test.cpu.getState() == CPUState::Idle);

        test.execute(5);

        REQUIRE(test.cpu.getState() == CPUState::Running);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Running);
    }
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
