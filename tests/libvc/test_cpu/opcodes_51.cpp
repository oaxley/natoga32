#include <catch2/catch_all.hpp>

#include "instruction_test.h"

using namespace vc;

TEST_CASE("Arithmetic signed instructions", "[cpu][instruction][arithmetic]") {
    InstructionTest test;

    SECTION("ADD") {
        test.setReg(6, 0x00000064);  // x6 = 100
        test.setReg(7, 0x00000032);  // x7 = 50
        test.loadInstruction(0x007302B3);   // ADD x5, x6, x7
        test.execute();
        REQUIRE(test.getReg(5) == 0x00000096);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("MUL") {
        test.setReg(9, 0x0000000A);   // x9 = 10
        test.setReg(10, 0x00000014);  // x10 = 20
        test.loadInstruction(0x02A48433);   // MUL x8, x9, x10
        test.execute();
        REQUIRE(test.getReg(8) == 0x000000C8);  // Result = 10 * 20 = 200
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("MUL - Overflow") {
        test.setReg(9, 0x00010000);   // x9 = 65536
        test.setReg(10, 0x00010000);  // x10 = 65536
        test.loadInstruction(0x02A48433);   // MUL x8, x9, x10
        test.execute();
        REQUIRE(test.getReg(8) == 0x00000000);  // Result = 0 (overflow)
        REQUIRE(test.deltaPC() == 4);
    }


    SECTION("MULH") {
        test.setReg(21, 0x00010000);    // x21 = 65536
        test.setReg(22, 0x00010000);    // x22 = 65536
        test.loadInstruction(0x036A9A33);   // MULH x20, x21, x22
        test.execute();
        REQUIRE(test.getReg(20) == 0x00000001);  // Result = 0 (overflow)
        REQUIRE(test.deltaPC() == 4);
    }


    SECTION("SUB") {
        test.setReg(12, 0x00000064);  // x12 = 100
        test.setReg(13, 0x00000032);  // x13 = 50
        test.loadInstruction(0x40D605B3);   // SUB x11, x12, x13
        test.execute();
        REQUIRE(test.getReg(11) == 0x00000032);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("DIV") {
        test.setReg(15, 0x00000064);  // x15 = 100
        test.setReg(16, 0x0000000A);  // x16 = 10
        test.loadInstruction(0x0307C733);   // DIV x14, x15, x16
        test.execute();
        REQUIRE(test.getReg(14) == 0x0000000A);  // Result = 100 / 10 = 10
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("DIV - Divide by zero") {
        test.setReg(15, 0x00000064);  // x15 = 100
        test.setReg(16, 0x00000000);  // x16 = 10
        test.loadInstruction(0x0307C733);   // DIV x14, x15, x16
        test.execute();
        REQUIRE(test.getReg(14) == 0xFFFFFFFF);  // Result = 100 / 0 = -1
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("DIV - Overflow") {
        test.setReg(15, 0x80000000);  // x15 = INT32_MIN = -2147483648
        test.setReg(16, 0xFFFFFFFF);  // x16 = -1
        test.loadInstruction(0x0307C733);   // DIV x14, x15, x16
        test.execute();
        REQUIRE(test.getReg(14) == 0x80000000);  // Result = INT32_MIN (overflow)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("REM") {
        test.setReg(18, 0x00000064);  // x18 = 100
        test.setReg(19, 0x0000000B);  // x19 = 11
        test.loadInstruction(0x033968B3);   // REM x17, x18, x19
        test.execute();
        REQUIRE(test.getReg(17) == 0x00000001);  // Result = 100 % 11 = 1
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("REM - Divide by zero") {
        test.setReg(18, 0x00000064);  // x18 = 100
        test.setReg(19, 0x00000000);  // x19 = 11
        test.loadInstruction(0x033968B3);   // REM x17, x18, x19
        test.execute();
        REQUIRE(test.getReg(17) == 0x00000064);  // Result = dividend (100) unchanged
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("REM - Overflow") {
        test.setReg(18, 0x80000000);  // x15 = INT32_MIN = -2147483648
        test.setReg(19, 0xFFFFFFFF);  // x16 = -1
        test.loadInstruction(0x033968B3);   // REM x17, x18, x19
        test.execute();
        REQUIRE(test.getReg(17) == 0x00000000);  // Result = 0 (overflow)
        REQUIRE(test.deltaPC() == 4);
    }

}
