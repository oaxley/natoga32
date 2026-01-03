#include <catch2/catch_all.hpp>

#include "instruction_test.h"

using namespace vc;

TEST_CASE("Branch Instructions", "[cpu][instructions][control-flow]") {
    InstructionTest test;

    SECTION("BEQ - Branch taken") {
        test.setReg(1, 42);
        test.setReg(2, 42);
        test.loadInstruction(0x00208463);   // BEQ x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 8);
    }

    SECTION("BEQ - Branch not taken") {
        test.setReg(1, 42);
        test.setReg(2, 24);
        test.loadInstruction(0x00208463);   // BEQ x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BNE - Branch taken") {
        test.setReg(1, 42);
        test.setReg(2, 24);
        test.loadInstruction(0x00209463);   // BNE x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 8);
    }

    SECTION("BNE - Branch not taken") {
        test.setReg(1, 42);
        test.setReg(2, 42);
        test.loadInstruction(0x00209463);   // BNE x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BLT - Branch taken") {
        test.setReg(1, -24);
        test.setReg(2, 42);
        test.loadInstruction(0x0020c463);   // BLT x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 8);
    }

    SECTION("BLT - Branch not taken") {
        test.setReg(1, 42);
        test.setReg(2, -24);
        test.loadInstruction(0x0020c463);   // BLT x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BGE - Branch taken") {
        test.setReg(1, 42);
        test.setReg(2, -24);
        test.loadInstruction(0x0020d463);   // BGE x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 8);
    }

    SECTION("BGE - Branch not taken") {
        test.setReg(1, -24);
        test.setReg(2, 42);
        test.loadInstruction(0x0020d463);   // BGE x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BLTU - Branch taken") {
        test.setReg(1, 24);
        test.setReg(2, -42);        //< unsigned value considered
        test.loadInstruction(0x0020e463);   // BLTU x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 8);
    }

    SECTION("BLTU - Branch not taken") {
        test.setReg(1, -42);        //< unsigned value considered
        test.setReg(2, 24);
        test.loadInstruction(0x0020e463);   // BLTU x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BGEU - Branch taken") {
        test.setReg(1, -24);        //< unsigned value considered
        test.setReg(2, 42);
        test.loadInstruction(0x0020f463);   // BGEU x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 8);
    }

    SECTION("BGEU - Branch not taken") {
        test.setReg(1, 42);
        test.setReg(2, -24);        //< unsigned value considered
        test.loadInstruction(0x0020f463);   // BGEU x1, x2, 8
        test.execute();

        REQUIRE(test.deltaPC() == 4);
    }
}

TEST_CASE("Jump Instructions", "[cpu][instructions][control-flow]") {
    InstructionTest test;

    SECTION("JAL") {
        u32 expected = MMAP_CODE + 4;
        test.loadInstruction(0x010000ef);       // JAL x1, 16
        test.execute();

        REQUIRE(test.getReg(1) == expected);
        REQUIRE(test.deltaPC() == 16);
    }

    SECTION("JALR") {
        u32 expected = MMAP_CODE + 4;
        test.setReg(2, MMAP_CODE + 16);
        test.loadInstruction(0x000100e7);       // JALR x1, 0(x2)
        test.execute();

        REQUIRE(test.getReg(1) == expected);
        REQUIRE(test.deltaPC() == 16);
    }
}
