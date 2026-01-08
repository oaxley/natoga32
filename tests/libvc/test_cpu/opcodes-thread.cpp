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
            switch (test.getCtx(tid).getState())
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
        REQUIRE(test.getCtx(tid).getPC() == value);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("END.T") {
        // load instruction in T0 NEW.T + YIELD
        u32 value = MMAP_CODE + 0x100;
        test.setReg(3, value);
        test.loadInstruction(0x0001810b);       // NEW.T x2, x3
        test.loadInstruction(0x0000100b);       // YIELD.T

        // 1st tick
        test.execute();

        // ensure we get ourself a new thread
        u32 tid = test.getReg(2);
        REQUIRE(test.getCtx(tid).getPC() == value);

        // execute the yield on thread 0
        test.execute();

        // checks
        REQUIRE((u32)test.cpu.getCurrentThreadID() == tid);
        REQUIRE(test.getCtx(tid).getState() == ThreadState::Running);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Ready);

        // load END.T instruction into T1
        test.loadInstruction(0x0000700b, 0x100, tid);
        test.execute();

        // check we are back to T0
        REQUIRE((u32)test.cpu.getCurrentThreadID() == 0);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Running);
        REQUIRE(test.getCtx(tid).getState() == ThreadState::Dead);
    }

    SECTION("ID.T") {
        // load id.t, new.t and then yield.t
        u32 value = MMAP_CODE + 0x100;
        test.loadInstruction(0x0000220B);       // ID.T x4
        test.setReg(3, value);
        test.loadInstruction(0x0001810b);       // NEW.T x2, x3
        test.loadInstruction(0x0000100b);       // YIELD.T

        // ID.T
        test.execute();
        REQUIRE(test.getReg(4) == 0);

        // NEW.T
        test.execute();
        u32 tid = test.getReg(2);
        REQUIRE(tid != 0);
        REQUIRE(test.getCtx(tid).getPC() == value);

        // load ID.T in the second thread
        test.loadInstruction(0x0000220B, 0x100, tid);

        // T0 YIELD.T
        test.execute();

        // T1 ID.T
        test.execute();
        REQUIRE(test.getReg(4, tid) == 1);
    }

    SECTION("SLEEP.T - State verification") {
        // create a new thread
        u32 value = MMAP_CODE + 0x100;
        test.setReg(3, value);
        test.loadInstruction(0x0001810b);       // NEW.T x2, x3
        test.execute();

        // sleep.t with a waitkey
        test.setReg(2, 0x1234);
        test.loadInstruction(0x0001400B);       // SLEEP.T x2, x0
        test.execute();

        // check status
        REQUIRE(test.getCtx(0).getState() == ThreadState::Sleeping);
        REQUIRE(test.getCtx(1).getState() == ThreadState::Running);
    }

    SECTION("SLEEP.T / WAKE.T - Waitkey") {
        // Thread 0: new + sleep
        u32 value = MMAP_CODE + 0x100;
        test.setReg(3, value);
        test.loadInstruction(0x0001810b);       // NEW.T x2, x3
        test.execute();

        test.setReg(2, 0x1234);              // x2 = 0x1234
        test.loadInstruction(0x0001400B);       // SLEEP.T x2, x0
        test.execute();

        // check status
        REQUIRE(test.getCtx(0).getState() == ThreadState::Sleeping);
        REQUIRE(test.getCtx(1).getState() == ThreadState::Running);

        // Thread 1: wake
        test.setReg(10, 0x1234, 1);      // x10 = 0x1234
        test.loadInstruction(0x0005500B, 0x100, 1);       // WAKE.T x10
        test.execute();

        REQUIRE(test.getCtx(0).getState() == ThreadState::Ready);
    }

    SECTION("SLEEP.T - Sleep until") {

    }

    SECTION("YIELD.T") {
        // load instruction in T0 NEW.T + YIELD
        u32 value = MMAP_CODE + 0x100;
        test.setReg(3, value);
        test.loadInstruction(0x0001810b);       // NEW.T x2, x3
        test.loadInstruction(0x0000100b);       // YIELD.T

        // 1st tick
        test.execute();

        // ensure we get ourself a new thread
        u32 tid = test.getReg(2);
        REQUIRE(test.getCtx(tid).getPC() == value);

        // execute the yield on thread 0
        test.execute();

        // checks
        REQUIRE((u32)test.cpu.getCurrentThreadID() == tid);
        REQUIRE(test.getCtx(tid).getState() == ThreadState::Running);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Ready);

        // load YIELD.T instruction into T1
        test.loadInstruction(0x0000100b, 0x100, tid);
        test.execute();

        // check we are back to T0
        REQUIRE((u32)test.cpu.getCurrentThreadID() == 0);
        REQUIRE(test.getCtx(0).getState() == ThreadState::Running);
        REQUIRE(test.getCtx(tid).getState() == ThreadState::Ready);
    }
}
