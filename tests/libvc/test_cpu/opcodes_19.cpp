#include <catch2/catch_all.hpp>

#include "instruction_test.h"

using namespace vc;

TEST_CASE("ADDI instruction", "[cpu][instruction]") {
    InstructionTest test;

    SECTION("MV pseudo-instruction") {
        test.loadInstruction(0x02a00093);       // addi x1, x0, 42
        test.execute();
        REQUIRE(test.getReg(1) == 42);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("ADDI with positive immediate") {
        test.setReg(1, 42);
        test.loadInstruction(0x01808113);       // addi x2, x1, 24
        test.execute();
        REQUIRE(test.getReg(2) == 66);
    }

    SECTION("ADDI with negative immediate") {
        test.setReg(1, 42);
        test.loadInstruction(0xfe808113);       // addi x2, x1, -24
        test.execute();
        REQUIRE(test.getReg(2) == 18);
    }
}

TEST_CASE("Shift instructions Immediate", "[cpu][instruction][shift]") {
    InstructionTest test;

    SECTION("SLLI") {
        test.setReg(2, 0x1234);
        test.loadInstruction(0x00411093);       // slli x1, x2, 4
        test.execute();
        REQUIRE(test.getReg(1) == 0x12340);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SLTI - True") {
        test.setReg(2, 3);
        test.loadInstruction(0x00412093);       // slti x1, x2, 4
        test.execute();
        REQUIRE(test.getReg(1) == 1);               // 3 < 4
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SLTI - False") {
        test.setReg(2, 5);
        test.loadInstruction(0x00412093);       // slti x1, x2, 4
        test.execute();
        REQUIRE(test.getReg(1) == 0);               // 5 > 4
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SLTIU - True") {
        test.setReg(2, 3);
        test.loadInstruction(0x00413093);       // sltiu x1, x2, 4
        test.execute();
        REQUIRE(test.getReg(1) == 1);               // 3 < 4
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SLTIU - False") {
        test.setReg(2, 5);
        test.loadInstruction(0x00413093);       // sltiu x1, x2, 4
        test.execute();
        REQUIRE(test.getReg(1) == 0);               // 5 > 4
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SRLI") {
        test.setReg(2, 0xA0012340);
        test.loadInstruction(0x00415093);       // srli x1, x2, 4
        test.execute();
        REQUIRE(test.getReg(1) == 0x0A001234);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SRAI") {
        test.setReg(2, 0xA0012340);
        test.loadInstruction(0x40415093);       // srai x1, x2, 4
        test.execute();
        REQUIRE(test.getReg(1) == 0xFA001234);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("RORI") {
        test.setReg(2, 0x12345678);
        test.loadInstruction(0x60415093);       // rori x1, x2, 4
        test.execute();
        REQUIRE(test.getReg(1) == 0x81234567);
        REQUIRE(test.deltaPC() == 4);
    }

}

TEST_CASE("Bit manipulations Immediate", "[cpu][instruction][bit]") {
    InstructionTest test;

    SECTION("BSETI") {
        test.setReg(4, 0x00000010);
        test.loadInstruction(0x28521193);       // bseti x3, x4, 5
        test.execute();
        REQUIRE(test.getReg(3) == 0x00000030);      // bit 4 and 5 set
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BCLRI") {
        test.setReg(6, 0x000000FF);
        test.loadInstruction(0x48331293);       // bclri x5, x6, 3
        test.execute();
        REQUIRE(test.getReg(5) == 0x000000F7);      // bit 3 cleared
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CLZ") {
        test.setReg(12, 0x00000FFF);          // 20 leading 0
        test.loadInstruction(0x60061593);       // clz x11, x12
        test.execute();
        REQUIRE(test.getReg(11) == 20);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CTZ") {
        test.setReg(14, 0x00000FF0);        // 4 trailing 0
        test.loadInstruction(0x60171693);       // ctz x13, x14
        test.execute();
        REQUIRE(test.getReg(13) == 4);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CPOP") {
        test.setReg(16, 0x000000FF);  // x16 = 0b11111111 (8 set bits)
        test.loadInstruction(0x60281793);       // cpop x15, x16
        test.execute();
        REQUIRE(test.getReg(15) == 8);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BINVI") {
        test.setReg(8, 0x00000020);           // bit 5 is set
        test.loadInstruction(0x68441393);       // binvi x7, x8, 4
        test.execute();
        REQUIRE(test.getReg(7) == 0x00000030);      // bit 4 and 5 set
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BEXTI") {
        test.setReg(10, 0x00000040);          // bit 6 is set
        test.loadInstruction(0x48655493);       // bexti x9, x10, 6
        test.execute();
        REQUIRE(test.getReg(9) == 0x00000001);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("REV8") {
        test.setReg(18, 0x12345678);  // x18 = 0x12345678
        test.loadInstruction(0x69895893);       // rev8 x17, x18
        test.execute();
        REQUIRE(test.getReg(17) == 0x78563412);
        REQUIRE(test.deltaPC() == 4);
    }
}

TEST_CASE("Sign extension", "[cpu][instruction][sign-ext]") {
    InstructionTest test;

    SECTION("SEXT.H") {
        test.setReg(22, 0x0000FFFF);
        test.loadInstruction(0x605B1A93);       // sext.h x21, x22
        test.execute();
        REQUIRE(test.getReg(21) == 0xFFFFFFFF);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SEXT.B") {
        test.setReg(20, 0x000000FF);
        test.loadInstruction(0x604A1993);       // sext.b x19, x20
        test.execute();
        REQUIRE(test.getReg(19) == 0xFFFFFFFF);
        REQUIRE(test.deltaPC() == 4);
    }
}

TEST_CASE("Logical operators Immediate", "[cpu][instruction][logical]") {
    InstructionTest test;

    SECTION("XORI") {
        test.setReg(24, 0x000000FF);
        test.loadInstruction(0x0F0C4B93);       // xori x23, x24, 0x0F0
        test.execute();
        REQUIRE(test.getReg(23) == 0x0000000F);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("ORI") {
        test.setReg(26, 0x000000AA);
        test.loadInstruction(0x055D6C93);       // ori x25, x26, 0x055
        test.execute();
        REQUIRE(test.getReg(25) == 0x000000FF);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("ANDI") {
        test.setReg(28, 0x000000FF);
        test.loadInstruction(0x0F0E7D93);       // andi x27, x28, 0x0F0
        test.execute();
        REQUIRE(test.getReg(27) == 0x000000F0);
        REQUIRE(test.deltaPC() == 4);
    }
}
