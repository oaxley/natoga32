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
 * @brief	Tests - stack canary: an overflow is fatal
 */

#include <catch2/catch_all.hpp>

#include "instruction_test.h"
#include "core/csr.h"
#include "vc/exceptions.h"
#include "vc/events.h"

using namespace vc;

// trap handler address used by the tests (in RAM: a test writes code there)
constexpr u32 HANDLER = MMAP_CODE + 0x1000;

// instructions
constexpr u32 NEW_T_X2_X3   = 0x0001810B;   // new.t x2, x3
constexpr u32 NEW_T_X2_X4   = 0x0002010B;   // new.t x2, x4
constexpr u32 YIELD_T       = 0x0000100B;   // yield.t
constexpr u32 SLEEP_T_A0_A1 = 0x00B5400B;   // sleep.t a0, a1
constexpr u32 END_T         = 0x0000700B;   // end.t
constexpr u32 NOP           = 0x00000013;   // addi x0, x0, 0

// entry points of threads 1 and 2
constexpr u32 ENTRY_1 = MMAP_CODE + 0x100;
constexpr u32 ENTRY_2 = MMAP_CODE + 0x200;

static u32 readCsr(InstructionTest& test, u16 id)
{
    return test.mem.read32(MMAP_CSR_REGISTERS + id * sizeof(u32));
}

static void writeCsr(InstructionTest& test, u16 id, u32 value)
{
    test.mem.write32(MMAP_CSR_REGISTERS + id * sizeof(u32), value);
}

static void breakCanary(InstructionTest& test, int tid)
{
    test.mem.write32(MMAP_STACK_BASE + tid * MMAP_THREAD_STACK_SIZE, 0);
}

// thread 0 starts threads 1 and 2, then yields: thread 1 runs `instruction` at ENTRY_1
static void startOnThread1(InstructionTest& test, u32 instruction)
{
    writeCsr(test, CSR_MTVEC, HANDLER);
    test.setReg(3, ENTRY_1);
    test.setReg(4, ENTRY_2);
    test.loadInstruction(NEW_T_X2_X3);
    test.loadInstruction(NEW_T_X2_X4);
    test.loadInstruction(YIELD_T);
    test.loadInstruction(NOP, 0x200, 2);
    test.loadInstruction(instruction, 0x100, 1);
    test.execute(3);
    REQUIRE(test.cpu.getCurrentThreadID() == 1);
}

static void requireFatalOnThread1(InstructionTest& test)
{
    REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_THREAD_STACK_OVERFLOW_ERROR);
    REQUIRE(readCsr(test, CSR_MEPC) == ENTRY_1);
    REQUIRE(test.getPC(1) == HANDLER);
    REQUIRE(test.cpu.getCurrentThreadID() == 1);
    REQUIRE(test.getCtx(1).getState() == ThreadState::Running);
    REQUIRE(test.getCtx(0).getState() == ThreadState::Dead);
    REQUIRE(test.getCtx(2).getState() == ThreadState::Dead);
    for (int tid = 3; tid < THREADS_COUNT; tid++) {
        REQUIRE(test.getCtx(tid).getState() == ThreadState::Dead);
    }
}

TEST_CASE("A broken canary is fatal: trap 26 and every other thread stops", "[cpu][thread][canary]") {
    InstructionTest test;

    SECTION("yield.t") {
        startOnThread1(test, YIELD_T);
        breakCanary(test, 1);
        test.execute();
        requireFatalOnThread1(test);
    }

    SECTION("sleep.t") {
        startOnThread1(test, SLEEP_T_A0_A1);
        test.setReg(10, HW_VBLANK, 1);
        breakCanary(test, 1);
        test.execute();
        requireFatalOnThread1(test);
        REQUIRE(test.getCtx(1).getWaitKey() == 0);
    }

    SECTION("end.t") {
        startOnThread1(test, END_T);
        breakCanary(test, 1);
        test.execute();
        requireFatalOnThread1(test);
    }

    SECTION("sleep.t with both operands at 0 does not sleep and does not check") {
        startOnThread1(test, SLEEP_T_A0_A1);
        breakCanary(test, 1);
        test.execute();
        REQUIRE(readCsr(test, CSR_MCAUSE) == 0);
        REQUIRE(test.getPC(1) == ENTRY_1 + 4);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Ready);
    }
}

TEST_CASE("A broken canary on thread 0 stops the other threads", "[cpu][thread][canary]") {
    InstructionTest test;
    writeCsr(test, CSR_MTVEC, HANDLER);

    SECTION("yield.t") {
        test.setReg(3, ENTRY_1);
        test.loadInstruction(NEW_T_X2_X3);
        test.loadInstruction(YIELD_T);
        test.execute();
        u32 instr_addr = test.getPC();
        breakCanary(test, 0);
        test.execute();

        REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_THREAD_STACK_OVERFLOW_ERROR);
        REQUIRE(readCsr(test, CSR_MEPC) == instr_addr);
        REQUIRE(test.getPC() == HANDLER);
        REQUIRE(test.cpu.getCurrentThreadID() == 0);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Running);
        REQUIRE(test.getCtx(1).getState() == ThreadState::Dead);
    }

    SECTION("end.t reports the overflow before the main-exit error") {
        test.loadInstruction(END_T);
        breakCanary(test, 0);
        test.execute();
        REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_THREAD_STACK_OVERFLOW_ERROR);
    }

    SECTION("without a handler the CPU halts") {
        writeCsr(test, CSR_MTVEC, 0);
        test.loadInstruction(YIELD_T);
        breakCanary(test, 0);
        test.execute();
        REQUIRE(test.cpu.getState() == CPUState::Halted);
        REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_THREAD_STACK_OVERFLOW_ERROR);
    }
}

TEST_CASE("After an overflow the handler can wait for the vblank", "[cpu][thread][canary]") {
    InstructionTest test;
    startOnThread1(test, YIELD_T);
    breakCanary(test, 1);
    test.execute();
    REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_THREAD_STACK_OVERFLOW_ERROR);

    // handler: sleep.t on HW_VBLANK, then a nop
    test.mem.write32(HANDLER, SLEEP_T_A0_A1);
    test.mem.write32(HANDLER + 4, NOP);
    test.setReg(10, HW_VBLANK, 1);
    test.setReg(11, 0, 1);
    writeCsr(test, CSR_MCAUSE, 0);
    test.execute();

    // the canary is not checked again: the thread sleeps, the CPU is idle
    REQUIRE(readCsr(test, CSR_MCAUSE) == 0);
    REQUIRE(test.getCtx(1).getState() == ThreadState::Sleeping);
    REQUIRE(test.cpu.getState() == CPUState::Idle);

    // the vblank wakes the handler, nothing else runs
    test.cpu.wakeThreadOnEvent(HW_VBLANK);
    REQUIRE(test.cpu.getState() == CPUState::Running);
    REQUIRE(test.cpu.getCurrentThreadID() == 1);
    test.execute();
    REQUIRE(test.getPC(1) == HANDLER + 8);
}

TEST_CASE("Reset clears the overflow state", "[cpu][thread][canary]") {
    InstructionTest test;
    startOnThread1(test, YIELD_T);
    breakCanary(test, 1);
    test.execute();
    REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_THREAD_STACK_OVERFLOW_ERROR);

    // the reset writes the canaries again: break thread 0's and check it traps
    test.cpu.reset();
    writeCsr(test, CSR_MCAUSE, 0);
    test.reset();
    test.loadInstruction(YIELD_T);
    breakCanary(test, 0);
    test.execute();
    REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_THREAD_STACK_OVERFLOW_ERROR);
}
