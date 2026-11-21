#include <catch2/catch_all.hpp>

#include "instruction_test.h"

using namespace vc;

// Zba: address generation. shNadd rd, rs1, rs2 = (rs1 << N) + rs2
TEST_CASE("Zba Instructions", "[cpu][instruction][zba]") {
    InstructionTest test;
    test.setReg(11, 0x00000003);    // a1 = index
    test.setReg(12, 0x00001000);    // a2 = base

    SECTION("SH1ADD") {
        test.loadInstruction(0x20C5A533);   // sh1add a0, a1, a2
        test.execute();
        REQUIRE(test.getReg(10) == 0x00001006);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SH2ADD") {
        test.loadInstruction(0x20C5C533);   // sh2add a0, a1, a2
        test.execute();
        REQUIRE(test.getReg(10) == 0x0000100C);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SH3ADD") {
        test.loadInstruction(0x20C5E533);   // sh3add a0, a1, a2
        test.execute();
        REQUIRE(test.getReg(10) == 0x00001018);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SH3ADD wraps around 32 bits") {
        test.setReg(11, 0xF0000001);
        test.setReg(12, 0x00000010);
        test.loadInstruction(0x20C5E533);   // sh3add a0, a1, a2
        test.execute();
        REQUIRE(test.getReg(10) == 0x80000018);
    }
}

// Zbb: logic with negate and orc.b
TEST_CASE("Zbb Logic Instructions", "[cpu][instruction][zbb]") {
    InstructionTest test;
    test.setReg(11, 0xFF00FF00);    // a1
    test.setReg(12, 0x0FF00FF0);    // a2

    SECTION("ANDN") {
        test.loadInstruction(0x40C5F533);   // andn a0, a1, a2
        test.execute();
        REQUIRE(test.getReg(10) == 0xF000F000);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("ORN") {
        test.loadInstruction(0x40C5E533);   // orn a0, a1, a2
        test.execute();
        REQUIRE(test.getReg(10) == 0xFF0FFF0F);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("XNOR") {
        test.loadInstruction(0x40C5C533);   // xnor a0, a1, a2
        test.execute();
        REQUIRE(test.getReg(10) == 0x0F0F0F0F);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("ORC.B sets each non-zero byte to 0xFF") {
        test.setReg(11, 0x00120080);
        test.loadInstruction(0x2875D513);   // orc.b a0, a1
        test.execute();
        REQUIRE(test.getReg(10) == 0x00FF00FF);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("ORC.B of 0 is 0") {
        test.setReg(11, 0);
        test.loadInstruction(0x2875D513);   // orc.b a0, a1
        test.execute();
        REQUIRE(test.getReg(10) == 0);
    }

    SECTION("BSETI still works (same funct7 as ORC.B)") {
        test.setReg(11, 0);
        test.loadInstruction(0x28759513);   // bseti a0, a1, 7
        test.execute();
        REQUIRE(test.getReg(10) == 0x80);
    }
}
