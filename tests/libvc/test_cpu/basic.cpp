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
