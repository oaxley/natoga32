#include <catch2/catch_all.hpp>

#include "instruction_test.h"
#include "core/csr.h"
#include "vc/exceptions.h"

using namespace vc;

// trap handler address used by the tests
constexpr u32 HANDLER = 0x0000'0200;

static u32 readCsr(InstructionTest& test, u16 id)
{
    return test.mem.read32(MMAP_CSR_REGISTERS + id * sizeof(u32));
}

static void setHandler(InstructionTest& test)
{
    test.mem.write32(MMAP_CSR_REGISTERS + CSR_MTVEC * sizeof(u32), HANDLER);
}

TEST_CASE("Traps jump to the handler and the instruction has no effect", "[cpu][trap]") {
    InstructionTest test;
    setHandler(test);

    SECTION("unknown opcode raises an illegal instruction trap") {
        test.loadInstruction(0x0000007F);               // opcode 0x7F does not exist
        u32 instr_addr = test.getPC();
        test.execute();
        REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_MAIN_ILLEGAL_INSTRUCTION);
        REQUIRE(readCsr(test, CSR_MEPC) == instr_addr);
        REQUIRE(test.getPC() == HANDLER);
    }

    SECTION("a trapping instruction is not counted in the thread cycles") {
        test.loadInstruction(0x0000007F);
        test.execute();
        REQUIRE(test.getCtx().getTotalCycles() == 0);
    }

    SECTION("writing a read-only CSR traps and leaves rd unchanged") {
        test.setReg(1, 0x55);
        test.setReg(3, 0x1234);
        test.loadInstruction(0xC00091F3);               // CSRRW x3, cycle, x1
        u32 instr_addr = test.getPC();
        test.execute();
        REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_MAIN_ILLEGAL_INSTRUCTION);
        REQUIRE(readCsr(test, CSR_MEPC) == instr_addr);
        REQUIRE(test.getPC() == HANDLER);
        REQUIRE(test.getReg(3) == 0x1234);
    }

    SECTION("END.T on the main thread traps") {
        test.loadInstruction(0x0000700B);               // END.T
        u32 instr_addr = test.getPC();
        test.execute();
        REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_THREAD_MAIN_EXIT_ERROR);
        REQUIRE(readCsr(test, CSR_MEPC) == instr_addr);
        REQUIRE(test.getPC() == HANDLER);
    }

    SECTION("NEW.T without a free slot traps") {
        test.setReg(3, MMAP_CODE + 0x100);
        for (int i = 0; i < THREADS_COUNT; i++) {
            test.loadInstruction(0x0001810b);           // NEW.T x2, x3
        }
        test.execute(THREADS_COUNT - 1);                // fills slots 1..7
        u32 instr_addr = test.getPC();
        test.setReg(2, 0xAAAA);
        test.execute();                                 // 8th NEW.T: no slot left
        REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_THREAD_SPAWN_ERROR);
        REQUIRE(readCsr(test, CSR_MEPC) == instr_addr);
        REQUIRE(test.getPC() == HANDLER);
        REQUIRE(test.getReg(2) == 0xAAAA);
    }
}

TEST_CASE("A trap without a handler halts the CPU", "[cpu][trap]") {
    InstructionTest test;                               // mtvec = 0: no handler

    test.loadInstruction(0x0000007F);                   // illegal instruction
    test.loadInstruction(0x00100093);                   // ADDI x1, x0, 1 (never executed)
    u32 instr_addr = test.getPC();
    test.execute();

    REQUIRE(test.cpu.getState() == CPUState::Halted);
    REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_MAIN_ILLEGAL_INSTRUCTION);
    REQUIRE(readCsr(test, CSR_MEPC) == instr_addr);
    REQUIRE(test.getPC() == instr_addr);

    // a halted CPU stays stopped
    test.execute(10);
    REQUIRE(test.cpu.getState() == CPUState::Halted);
    REQUIRE(test.getReg(1) == 0);
    REQUIRE(test.cpu.skipIdle(100) == 0);
}
