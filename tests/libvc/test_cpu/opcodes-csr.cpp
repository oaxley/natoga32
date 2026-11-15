#include <catch2/catch_all.hpp>

#include "instruction_test.h"

using namespace vc;

TEST_CASE("CSR Instructions", "[cpu][instruction][csr]") {
    InstructionTest test;

    SECTION("CSRRW") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x300 * sizeof(u32);
        test.mem.write32(csr_addr, 0xAABBCCDD);  // Initialize CSR 0x300 with a value
        test.setReg(6, 0x12345678);               // x6 = 0x12345678
        test.loadInstruction(0x300312F3);   // CSRRW x5, 0x300, x6
        test.execute();
        REQUIRE(test.getReg(5) == 0xAABBCCDD);           // x5 should contain old CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0x12345678); // CSR should contain value from x6
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CSRRS") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x301 * sizeof(u32);
        test.mem.write32(csr_addr, 0xF0F0F0F0);  // Initialize CSR 0x301
        test.setReg(8, 0x0F0F0F0F);               // x8 = bits to set
        test.loadInstruction(0x301423F3);   // CSRRS x7, 0x301, x8
        test.execute();
        REQUIRE(test.getReg(7) == 0xF0F0F0F0);           // x7 should contain old CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0xFFFFFFFF); // CSR = 0xF0F0F0F0 | 0x0F0F0F0F = 0xFFFFFFFF
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CSRRS - Read Only") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x303 * sizeof(u32);
        test.mem.write32(csr_addr, 0x12345678);  // Initialize CSR 0x303
        test.loadInstruction(0x303025F3);   // CSRRS x11, 0x303, x0
        test.execute();
        REQUIRE(test.getReg(11) == 0x12345678);           // x11 should contain CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0x12345678); // CSR should be UNCHANGED (read-only)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CSRRC") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x302 * sizeof(u32);
        test.mem.write32(csr_addr, 0xFFFFFFFF);  // Initialize CSR 0x302 (all bits set)
        test.setReg(10, 0x0F0F0F0F);              // x10 = bits to clear
        test.loadInstruction(0x302534F3);   // CSRRC x9, 0x302, x10
        test.execute();
        REQUIRE(test.getReg(9) == 0xFFFFFFFF);           // x9 should contain old CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0xF0F0F0F0); // CSR = 0xFFFFFFFF & ~0x0F0F0F0F = 0xF0F0F0F0
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CSRRC - Read Only") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x304 * sizeof(u32);
        test.mem.write32(csr_addr, 0xABCDEF00);  // Initialize CSR 0x304
        test.loadInstruction(0x30403673);   // CSRRC x12, 0x304, x0
        test.execute();
        REQUIRE(test.getReg(12) == 0xABCDEF00);           // x12 should contain CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0xABCDEF00); // CSR should be UNCHANGED (read-only)
        REQUIRE(test.deltaPC() == 4);
    }
}

TEST_CASE("CSR Instructions Immediate", "[cpu][instruction][csr]") {
    InstructionTest test;

    SECTION("CSRRWI") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x305 * sizeof(u32);
        test.mem.write32(csr_addr, 0xDEADBEEF);  // Initialize CSR 0x305 with a value
        test.loadInstruction(0x3057D6F3);   // CSRRWI x13, 0x305, 15
        test.execute();
        REQUIRE(test.getReg(13) == 0xDEADBEEF);        // x13 should contain old CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0x0000000F); // CSR should contain immediate value 15
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CSRRSI") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x306 * sizeof(u32);
        test.mem.write32(csr_addr, 0xF0F0F0F0);  // Initialize CSR 0x306
        test.loadInstruction(0x3067E773);   // CSRRSI x14, 0x306, 0x0F
        test.execute();
        REQUIRE(test.getReg(14) == 0xF0F0F0F0);           // x14 should contain old CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0xF0F0F0FF); // CSR = 0xF0F0F0F0 | 0x0F = 0xF0F0F0FF
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CSRRSI - Read Only") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x307 * sizeof(u32);
        test.mem.write32(csr_addr, 0x87654321);  // Initialize CSR 0x307
        test.loadInstruction(0x307067F3);   // CSRRSI x15, 0x307, 0
        test.execute();
        REQUIRE(test.getReg(15) == 0x87654321);           // x15 should contain CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0x87654321); // CSR should be UNCHANGED (read-only)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CSRRCI") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x308 * sizeof(u32);
        test.mem.write32(csr_addr, 0xFFFFFFFF);  // Initialize CSR 0x308 (all bits set)
        test.loadInstruction(0x3087F873);   // CSRRCI x16, 0x308, 0x0F
        test.execute();
        REQUIRE(test.getReg(16) == 0xFFFFFFFF);           // x16 should contain old CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0xFFFFFFF0); // CSR = 0xFFFFFFFF & ~0x0F = 0xFFFFFFF0
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CSRRCI - Read Only") {
        u32 csr_addr = MMAP_CSR_REGISTERS + 0x309 * sizeof(u32);
        test.mem.write32(csr_addr, 0xCAFEBABE);  // Initialize CSR 0x309
        test.loadInstruction(0x309078F3);   // CSRRCI x17, 0x309, 0
        test.execute();
        REQUIRE(test.getReg(17) == 0xCAFEBABE);           // x17 should contain CSR value
        REQUIRE(test.mem.read32(csr_addr) == 0xCAFEBABE); // CSR should be UNCHANGED (read-only)
        REQUIRE(test.deltaPC() == 4);
    }

}
