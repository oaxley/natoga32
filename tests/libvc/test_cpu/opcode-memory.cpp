#include <catch2/catch_all.hpp>

#include "instruction_test.h"

using namespace vc;

TEST_CASE("CPU Load/Store Instructions", "[cpu][instructions][memory]") {
    InstructionTest test;

    SECTION("SB and LB/LBU") {
        // SB
        {
            test.setReg(1, MMAP_DATA);
            test.setReg(2, 0xAB);
            test.loadInstruction(0x00208023);       // SB x2, 0(x1)
            test.execute();
            REQUIRE(test.mem.read8(MMAP_DATA) == 0xAB);
            REQUIRE(test.deltaPC() == 4);
        }

        // LB with positive value
        {
            test.mem.write32(MMAP_DATA, 0x12);
            test.setReg(1, MMAP_DATA);
            test.loadInstruction(0x00008183);       // LB x3, 0(x1)
            test.execute();
            u32 value = test.getReg(3);
            REQUIRE(value == 0x12);
            REQUIRE((value & 0x8000'0000) == 0);
            REQUIRE(test.deltaPC() == 4);
        }

        // LB with negative value
        {
            test.mem.write32(MMAP_DATA, static_cast<u32>(-34));
            test.setReg(1, MMAP_DATA);
            test.loadInstruction(0x00008183);       // LB x3, 0(x1)
            test.execute();
            u32 value = test.getReg(3);
            REQUIRE(static_cast<i8>(value) == -34);
            REQUIRE((value & 0x8000'0000) == 0x8000'0000);
            REQUIRE(test.deltaPC() == 4);
        }

        // LBU with a value > 128
        {
            test.mem.write32(MMAP_DATA, 171);
            test.setReg(1, MMAP_DATA);
            test.loadInstruction(0x0000c183);       // LBU x3, 0(x1)
            test.execute();
            u32 value = test.getReg(3);
            REQUIRE(value == 171);
            REQUIRE((value & 0x8000'0000) == 0);
            REQUIRE(test.deltaPC() == 4);
        }
    }

    SECTION("SH and LH") {
        // SH
        {
            test.setReg(1, MMAP_DATA);
            test.setReg(2, 0x1234);
            test.loadInstruction(0x00209023);       // SH x2, 0(x1)
            test.execute();
            REQUIRE(test.mem.read16(MMAP_DATA) == 0x1234);
            REQUIRE(test.deltaPC() == 4);
        }

        // LH with positive value
        {
            test.mem.write32(MMAP_DATA, 16384);
            test.setReg(1, MMAP_DATA);
            test.loadInstruction(0x00009183);       // LH x3, 0(x1)
            test.execute();
            u32 value = test.getReg(3);
            REQUIRE(value == 16384);
            REQUIRE((value & 0x8000'0000) == 0);
            REQUIRE(test.deltaPC() == 4);
        }

        // LH with negative value
        {
            test.mem.write32(MMAP_DATA, static_cast<u32>(-16384));
            test.setReg(1, MMAP_DATA);
            test.loadInstruction(0x00009183);       // LH x3, 0(x1)
            test.execute();
            u32 value = test.getReg(3);
            REQUIRE(static_cast<i16>(value) == -16384);
            REQUIRE((value & 0x8000'0000) == 0x8000'0000);
            REQUIRE(test.deltaPC() == 4);
        }

        // LHU with unsigned value
        {
            test.mem.write32(MMAP_DATA, static_cast<u32>(-16384));
            test.setReg(1, MMAP_DATA);
            test.loadInstruction(0x0000d183);       // LH x3, 0(x1)
            test.execute();
            u32 value = test.getReg(3);
            REQUIRE(value == 49152);            // unsigned(-16384) = 49152
            REQUIRE((value & 0x8000'0000) == 0);
            REQUIRE(test.deltaPC() == 4);
        }
    }

    SECTION("SW and LW") {
        // SW
        {
            test.setReg(1, MMAP_DATA);
            test.setReg(2, 0xCAFEBABE);
            test.loadInstruction(0x0020a023);       // SW x2, 0(x1)
            test.execute();
            REQUIRE(test.mem.read32(MMAP_DATA) == 0xCAFEBABE);
            REQUIRE(test.deltaPC() == 4);
        }

        // LW
        {
            test.mem.write32(MMAP_DATA, 0xBAADF00D);
            test.setReg(1, MMAP_DATA);
            test.loadInstruction(0x0000a183);       // LW x3, 0(x1)
            test.execute();
            REQUIRE(test.getReg(3) == 0xBAADF00D);
            REQUIRE(test.deltaPC() == 4);
        }
    }
}

TEST_CASE("x0 should remain 0 at all time", "[cpu][registers]") {
    InstructionTest test;

    test.loadInstruction(0x23400013);       // ADDI x0, x0, 564
    test.execute();
    REQUIRE(test.getReg(0) == 0);
}
