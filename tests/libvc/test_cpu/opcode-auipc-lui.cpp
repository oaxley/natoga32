#include <catch2/catch_all.hpp>

#include "instruction_test.h"

using namespace vc;

TEST_CASE("", "[cpu][instructions][u_type]") {
    InstructionTest test;

    SECTION("AUIPC") {
        // PC = MMAP_CODE upon initialization
        u32 expected = MMAP_CODE + (0x12345 << 12);

        test.loadInstruction(0x12345097);       // AUIPC x1, 0x12345
        test.execute();

        REQUIRE(test.getReg(1) == expected);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("Load Upper Immediate") {
        test.loadInstruction(0x123450B7);       // LUI x1, 0x12345
        test.execute();

        REQUIRE(test.getReg(1) == 0x12345000);
        REQUIRE(test.deltaPC() == 4);
    }
}
