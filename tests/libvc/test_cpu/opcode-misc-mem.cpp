#include <catch2/catch_all.hpp>

#include "instruction_test.h"
#include "core/csr.h"
#include "vc/exceptions.h"

using namespace vc;

// FENCE and FENCE.I (opcode 0x0F): one hart, no cache, no reordering, so both
// only move the PC. They are decoded so that code using them keeps working if
// a DMA or a cache is added later.
TEST_CASE("MISC-MEM Instructions", "[cpu][instruction][fence]") {
    InstructionTest test;

    SECTION("FENCE moves to the next instruction and changes no register") {
        test.setReg(5, 0x1234);
        test.loadInstruction(0x0FF0000F);   // FENCE iorw, iorw
        test.execute();
        REQUIRE(test.deltaPC() == 4);
        REQUIRE(test.getReg(5) == 0x1234);
        REQUIRE(test.getCtx().getTotalCycles() == 1);
    }

    SECTION("FENCE with other sets, rd and rs1 fields is still a no-op") {
        test.setReg(1, 0xAAAA);
        test.loadInstruction(0x0330000F);   // FENCE rw, rw
        test.loadInstruction(0x8330808F);   // fm = 8 (FENCE.TSO), rd = x1, rs1 = x1
        test.execute(2);
        REQUIRE(test.deltaPC() == 8);
        REQUIRE(test.getReg(1) == 0xAAAA);
    }

    SECTION("FENCE.I moves to the next instruction") {
        test.loadInstruction(0x0000100F);   // FENCE.I
        test.execute();
        REQUIRE(test.deltaPC() == 4);
        REQUIRE(test.getCtx().getTotalCycles() == 1);
    }

    SECTION("code written to memory runs after FENCE.I") {
        // sw x6, 8(x5) writes "addi x7, x0, 42" right after the FENCE.I
        test.loadInstruction(0x0062A423);   // SW x6, 8(x5)
        test.loadInstruction(0x0000100F);   // FENCE.I
        test.loadInstruction(0x00000013);   // NOP, overwritten by the store
        test.setReg(5, test.getPC());       // x5 = address of the SW
        test.setReg(6, 0x02A00393);         // ADDI x7, x0, 42
        test.execute(3);
        REQUIRE(test.getReg(7) == 42);
    }
}

TEST_CASE("MISC-MEM with a reserved funct3 traps", "[cpu][instruction][fence][trap]") {
    InstructionTest test;
    constexpr u32 HANDLER = 0x0000'0200;
    test.mem.write32(MMAP_CSR_REGISTERS + CSR_MTVEC * sizeof(u32), HANDLER);

    for (u32 funct3 = 2; funct3 < 8; funct3++) {
        DYNAMIC_SECTION("funct3 " << funct3) {
            test.loadInstruction(0x0000000F | (funct3 << 12));
            u32 instr_addr = test.getPC();
            test.execute();
            REQUIRE(test.mem.read32(MMAP_CSR_REGISTERS + CSR_MCAUSE * sizeof(u32)) == CPU_MAIN_ILLEGAL_INSTRUCTION);
            REQUIRE(test.mem.read32(MMAP_CSR_REGISTERS + CSR_MEPC * sizeof(u32)) == instr_addr);
            REQUIRE(test.getPC() == HANDLER);
        }
    }
}
