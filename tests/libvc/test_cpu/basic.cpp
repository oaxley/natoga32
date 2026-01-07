#include <catch2/catch_all.hpp>

#include "instruction_test.h"


using namespace vc;

TEST_CASE("CPU Basics", "[cpu][basic]") {
    InstructionTest test;

    SECTION("Thread 0 is main thread after reset") {
        REQUIRE(test.cpu.getCurrentThreadID() == 0);
    }

    SECTION("Thread 0 starts at RESET_DEFAULT_ADDR") {
        REQUIRE(test.getCtx().pc == RESET_DEFAULT_ADDR);
    }

    SECTION("All threads are free except thread 0") {
        REQUIRE(test.getCtx().state == ThreadState::Running);

        for (int tid = 1; tid < THREADS_COUNT; tid++) {
            REQUIRE(test.getCtx(tid).state == ThreadState::Free);
        }
    }

    SECTION("All registers are set to 0") {
        for (int tid = 0; tid < THREADS_COUNT; tid++) {
            auto& thread = test.getCtx(tid);

            for (int r = 0; r < THREADS_REGISTERS; r++) {
                switch (r)
                {
                    case 2:    // specific value for SP (x2)
                    {
                        u32 value = MMAP_STACK_BASE + ((tid + 1) * MMAP_STACK_SIZE) - 1;
                        REQUIRE(thread.sp_base == value);
                        REQUIRE(thread.sp == value);
                        REQUIRE(thread.registers[r] == value);
                        break;
                    }
                    case 4:    // specific value for TP (x4)
                    {
                        u32 value = MMAP_THREADS_TLS_BASE + tid * MMAP_TTLS_SIZE;
                        REQUIRE(thread.tp == value);
                        REQUIRE(thread.registers[r] == value);
                        break;
                    }
                    default:
                        REQUIRE(thread.registers[r] == 0);
                        break;
                }
            }
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
