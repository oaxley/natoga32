#include <catch2/catch_all.hpp>

#include "instruction_test.h"


using namespace vc;

TEST_CASE("CPU Basics", "[cpu][basic]") {
    InstructionTest test;

    SECTION("Thread 0 is main thread after reset") {
        REQUIRE(test.cpu.getCurrentThreadID() == 0);
    }

    SECTION("Thread 0 starts at RESET_DEFAULT_ADDR") {
        REQUIRE(test.getCtx().getPC() == RESET_DEFAULT_ADDR);
    }

    SECTION("All threads are free except thread 0") {
        REQUIRE(test.getCtx().getState() == ThreadState::Running);

        for (int tid = 1; tid < THREADS_COUNT; tid++) {
            REQUIRE(test.getCtx(tid).getState() == ThreadState::Free);
        }
    }

    SECTION("All registers are set to 0") {
        for (int tid = 0; tid < THREADS_COUNT; tid++) {
            auto& thread = test.getCtx(tid);
            auto& registers = thread.getRegisters();

            for (int r = 0; r < THREADS_REGISTERS; r++) {
                switch (r)
                {
                    case 2:    // specific value for SP (x2)
                    {
                        // top of the thread's 64KiB stack, 16-byte aligned (RISC-V ABI)
                        u32 value = 0x0078'0000 + (tid + 1) * 0x1'0000;
                        REQUIRE(thread.getSP() == value);
                        REQUIRE((thread.getSP() % 16) == 0);
                        REQUIRE(registers[r] == value);
                        break;
                    }
                    case 4:    // specific value for TP (x4)
                    {
                        // 256B of thread local storage per thread
                        u32 value = 0x0020'6000 + tid * 256;
                        REQUIRE(thread.getTP() == value);
                        REQUIRE(registers[r] == value);
                        break;
                    }
                    default:
                        REQUIRE(registers[r] == 0);
                        break;
                }
            }
        }
    }

    SECTION("Stack canary is written at the bottom of each thread stack") {
        for (int tid = 0; tid < THREADS_COUNT; tid++) {
            u32 canary_addr = 0x0078'0000 + tid * 0x1'0000;
            REQUIRE(test.mem.read32(canary_addr) == STACK_CANARY_VALUE);
        }
    }

    SECTION("CPU State should be running") {
        REQUIRE(test.cpu.getState() == CPUState::Running);
    }

    SECTION("Check the Memory vs Registers span view") {
        int thread_id = 6;
        int reg_num = 14;
        u32 value = 0xBAADF00D;

        // retrieve the base address for Thread 6
        u32 reg_size = THREADS_REGISTERS * sizeof(u32);
        u32 base_addr = MMAP_THREADS_REG_BASE + thread_id * reg_size;
        u32 reg_addr = base_addr + reg_num * sizeof(u32);

        // write a value in memory
        test.mem.write32(reg_addr, value);

        // check for the register thread
        REQUIRE(test.getReg(reg_num, thread_id) == value);
    }
}

// address of the Thread Control Block of a thread
static u32 tcb(int tid)
{
    return MMAP_THREADS_TCB_BASE + tid * MMAP_THREAD_TCB_SIZE;
}

TEST_CASE("Thread Control Block lives in memory", "[cpu][threads]") {
    InstructionTest test;

    SECTION("reset writes the main thread state and PC in its TCB") {
        REQUIRE(test.mem.read8(tcb(0)) == static_cast<u8>(ThreadState::Running));
        REQUIRE(test.mem.read32(tcb(0) + 0x04) == RESET_DEFAULT_ADDR);
        REQUIRE(test.mem.read8(tcb(3)) == static_cast<u8>(ThreadState::Free));
    }

    SECTION("setPC is stored in the TCB") {
        test.getCtx(0).setPC(0x01A0'0010);
        REQUIRE(test.mem.read32(tcb(0) + 0x04) == 0x01A0'0010);
    }

    SECTION("sleep / wake / end are stored in the TCB") {
        test.getCtx(1).sleep(100, 0x42, 50);
        REQUIRE(test.mem.read8(tcb(1)) == static_cast<u8>(ThreadState::Sleeping));
        REQUIRE(test.mem.read32(tcb(1) + 0x08) == 0x42);
        REQUIRE(test.mem.read32(tcb(1) + 0x10) == 150);
        REQUIRE(test.mem.read32(tcb(1) + 0x14) == 0);

        test.getCtx(1).wake();
        REQUIRE(test.mem.read8(tcb(1)) == static_cast<u8>(ThreadState::Ready));
        REQUIRE(test.mem.read32(tcb(1) + 0x08) == 0);
        REQUIRE(test.mem.read32(tcb(1) + 0x10) == 0);

        test.getCtx(1).end();
        REQUIRE(test.mem.read8(tcb(1)) == static_cast<u8>(ThreadState::Dead));
    }

    SECTION("total cycles are stored in the TCB") {
        test.getCtx(0).incrementCycles();
        test.getCtx(0).incrementCycles();
        REQUIRE(test.mem.read32(tcb(0) + 0x18) == 2);
        REQUIRE(test.getCtx(0).getTotalCycles() == 2);
    }

    SECTION("memory writes (debugger path) are seen by the thread") {
        test.mem.write32(tcb(2) + 0x08, 0x99);
        REQUIRE(test.getCtx(2).getWaitKey() == 0x99);
        test.mem.write8(tcb(2), static_cast<u8>(ThreadState::Sleeping));
        REQUIRE(test.getCtx(2).getState() == ThreadState::Sleeping);
    }

    SECTION("init resets a dirty TCB") {
        test.mem.write32(tcb(5) + 0x08, 0x1234);
        test.mem.write32(tcb(5) + 0x10, 77);
        test.mem.write32(tcb(5) + 0x18, 88);
        test.getCtx(5).init(test.mem, 5, 0x01A0'0100);
        REQUIRE(test.mem.read32(tcb(5) + 0x08) == 0);
        REQUIRE(test.mem.read32(tcb(5) + 0x10) == 0);
        REQUIRE(test.mem.read32(tcb(5) + 0x18) == 0);
        REQUIRE(test.mem.read32(tcb(5) + 0x04) == 0x01A0'0100);
        REQUIRE(test.getCtx(5).getState() == ThreadState::Free);
    }
}
