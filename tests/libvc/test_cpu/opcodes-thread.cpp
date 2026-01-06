#include <catch2/catch_all.hpp>
#include <iostream>

#include "instruction_test.h"

using namespace vc;

TEST_CASE("Thread instructions", "[cpu][instruction][thread]") {
    InstructionTest test;

    SECTION("NEW.T") {
        u32 value = MMAP_CODE + 0x100;
        test.setReg(3, value);
        test.loadInstruction(0x0001810b);          // NEW.T x2, x3
        test.execute();

        // count the number of threads
        int count = 0;
        for (int tid = 0; tid < THREADS_COUNT; tid++) {
            switch (test.getCtx(tid).state)
            {
                case ThreadState::Running:
                case ThreadState::Ready:
                    count++;
                    break;
                default:
                    break;
            }
        }

        int tid = test.getReg(2, 0);
        REQUIRE(count == 2);
        REQUIRE(tid != 0);
        REQUIRE(test.getCtx(tid).pc == value);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("END.T") {
        // load instruction in T0 NEW.T + YIELD
        u32 value = MMAP_CODE + 0x100;
        test.setReg(3, value);
        test.loadInstruction(0x0001810b);       // NEW.T x2, x3
        test.loadInstruction(0x0000100b, 4);       // YIELD.T

        // 1st tick
        test.execute();

        // ensure we get ourself a new thread
        u32 tid = test.getReg(2);
        REQUIRE(test.getCtx(tid).pc == value);

        // execute the yield on thread 0
        test.execute();

        // checks
        REQUIRE((u32)test.cpu.getCurrentThreadID() == tid);
        REQUIRE(test.getCtx(tid).state == ThreadState::Running);
        REQUIRE(test.getCtx(0).state == ThreadState::Ready);

        // load END.T instruction into T1
        test.loadInstruction(0x0000700b, 0x100, tid);
        test.execute();

        // check we are back to T0
        REQUIRE((u32)test.cpu.getCurrentThreadID() == 0);
        REQUIRE(test.getCtx(0).state == ThreadState::Running);
        REQUIRE(test.getCtx(tid).state == ThreadState::Dead);
    }

    SECTION("ID.T") {

    }

    SECTION("SLEEP.T") {

    }

    SECTION("WAKE.T") {

    }

    SECTION("YIELD.T") {
        // load instruction in T0 NEW.T + YIELD
        u32 value = MMAP_CODE + 0x100;
        test.setReg(3, value);
        test.loadInstruction(0x0001810b);       // NEW.T x2, x3
        test.loadInstruction(0x0000100b, 4);       // YIELD.T

        // 1st tick
        test.execute();

        // ensure we get ourself a new thread
        u32 tid = test.getReg(2);
        REQUIRE(test.getCtx(tid).pc == value);

        // execute the yield on thread 0
        test.execute();

        // checks
        REQUIRE((u32)test.cpu.getCurrentThreadID() == tid);
        REQUIRE(test.getCtx(tid).state == ThreadState::Running);
        REQUIRE(test.getCtx(0).state == ThreadState::Ready);

        // load YIELD.T instruction into T1
        test.loadInstruction(0x0000100b, 0x100, tid);
        test.execute();

        // check we are back to T0
        REQUIRE((u32)test.cpu.getCurrentThreadID() == 0);
        REQUIRE(test.getCtx(0).state == ThreadState::Running);
        REQUIRE(test.getCtx(tid).state == ThreadState::Ready);
    }
}
