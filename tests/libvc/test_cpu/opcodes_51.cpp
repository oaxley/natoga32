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

TEST_CASE("Arithmetic unsigned operations", "[cpu][instruction][arithmetic]") {
    InstructionTest test;

    SECTION("MULHU") {
        test.setReg(24, 0xFFFFFFFF);  // x24 = 4,294,967,295 (max u32)
        test.setReg(25, 0x00000002);  // x25 = 2
        test.loadInstruction(0x039C3BB3);   // MULHU x23, x24, x25
        test.execute();
        REQUIRE(test.getReg(23) == 0x00000001);  // Upper 32 bits of 0xFFFFFFFF * 2
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("MULHSU") {
        test.setReg(27, 0xFFFFFFFE);  // x27 = -2 (as signed)
        test.setReg(28, 0x00000002);  // x28 = 2 (as unsigned)
        test.loadInstruction(0x03CDAD33);   // MULHSU x26, x27, x28
        test.execute();
        REQUIRE(test.getReg(26) == 0xFFFFFFFF);  // Upper 32 bits of (-2) * 2 = -4
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("DIVU") {
        test.setReg(30, 0x00000064);  // x30 = 100
        test.setReg(31, 0x0000000A);  // x31 = 10
        test.loadInstruction(0x03FF5EB3);   // DIVU x29, x30, x31
        test.execute();
        REQUIRE(test.getReg(29) == 0x0000000A);  // Result = 100 / 10 = 10
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("DIVU - Divide by zero") {
        test.setReg(30, 0x00000064);  // x30 = 100
        test.setReg(31, 0x00000000);  // x31 = 0
        test.loadInstruction(0x03FF5EB3);   // DIVU x29, x30, x31
        test.execute();
        REQUIRE(test.getReg(29) == 0xFFFFFFFF);  // Result = max unsigned (divide by zero)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("DIVU - Large Unsigned value") {
        test.setReg(30, 0xFFFFFFFF);  // x30 = 4,294,967,295 (max u32)
        test.setReg(31, 0x00000002);  // x31 = 2
        test.loadInstruction(0x03FF5EB3);   // DIVU x29, x30, x31
        test.execute();
        REQUIRE(test.getReg(29) == 0x7FFFFFFF);  // Result = 4,294,967,295 / 2 = 2,147,483,647
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("REMU") {
        test.setReg(2, 0x00000064);  // x2 = 100
        test.setReg(3, 0x0000000B);  // x3 = 11
        test.loadInstruction(0x023170B3);   // REMU x1, x2, x3
        test.execute();
        REQUIRE(test.getReg(1) == 0x00000001);  // Result = 100 % 11 = 1
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("REMU - Divide by zero") {
        test.setReg(2, 0x00000064);  // x2 = 100
        test.setReg(3, 0x00000000);  // x3 = 0
        test.loadInstruction(0x023170B3);   // REMU x1, x2, x3
        test.execute();
        REQUIRE(test.getReg(1) == 0x00000064);  // Result = dividend (100) unchanged
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("REMU - Large unsigned value") {
        test.setReg(2, 0xFFFFFFFF);  // x2 = 4,294,967,295 (max u32)
        test.setReg(3, 0x00000003);  // x3 = 3
        test.loadInstruction(0x023170B3);   // REMU x1, x2, x3
        test.execute();
        REQUIRE(test.getReg(1) == 0x00000000);  // Result = 4,294,967,295 % 3 = 0
        REQUIRE(test.deltaPC() == 4);
    }
}

TEST_CASE("Shift Instructions", "[cpu][instruction][shift]") {
    InstructionTest test;

    SECTION("SLL") {
        test.setReg(5, 0x00000003);  // x5 = 0b11
        test.setReg(6, 0x00000004);  // x6 = 4 (shift amount)
        test.loadInstruction(0x00629233);   // SLL x4, x5, x6
        test.execute();
        REQUIRE(test.getReg(4) == 0x00000030);  // Result = 0b11 << 4 = 0b110000 = 0x30
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SLT") {
        test.setReg(8, 0xFFFFFFFE);  // x8 = -2 (signed)
        test.setReg(9, 0x00000001);  // x9 = 1
        test.loadInstruction(0x009423B3);   // SLT x7, x8, x9
        test.execute();
        REQUIRE(test.getReg(7) == 0x00000001);  // Result = 1 (-2 < 1 is true)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SLTU") {
        test.setReg(11, 0xFFFFFFFE);  // x11 = 4,294,967,294 (unsigned)
        test.setReg(12, 0x00000001);  // x12 = 1
        test.loadInstruction(0x00C5B533);   // SLTU x10, x11, x12
        test.execute();
        REQUIRE(test.getReg(10) == 0x00000000);  // Result = 0 (4,294,967,294 > 1)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SRL") {
        test.setReg(14, 0x80000000);  // x14 = 0b10000000...
        test.setReg(15, 0x00000004);  // x15 = 4 (shift amount)
        test.loadInstruction(0x00F756B3);   // SRL x13, x14, x15
        test.execute();
        REQUIRE(test.getReg(13) == 0x08000000);  // Result = 0x80000000 >> 4 = 0x08000000 (logical)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("SRA") {
        test.setReg(17, 0x80000000);  // x17 = -2147483648 (signed)
        test.setReg(18, 0x00000004);  // x18 = 4 (shift amount)
        test.loadInstruction(0x4128D833);   // SRA x16, x17, x18
        test.execute();
        REQUIRE(test.getReg(16) == 0xF8000000);  // Result = 0x80000000 >> 4 = 0xF8000000 (arithmetic, sign-extended)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("ROR") {
        test.setReg(20, 0x12345678);  // x20 = 0x12345678
        test.setReg(21, 0x00000004);  // x21 = 4 (rotate amount)
        test.loadInstruction(0x615A59B3);   // ROR x19, x20, x21
        test.execute();
        REQUIRE(test.getReg(19) == 0x81234567);  // Result = rotate right by 4
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("ROL") {
        test.setReg(23, 0x12345678);  // x23 = 0x12345678
        test.setReg(24, 0x00000004);  // x24 = 4 (rotate amount)
        test.loadInstruction(0x618B9B33);   // ROL x22, x23, x24
        test.execute();
        REQUIRE(test.getReg(22) == 0x23456781);  // Result = rotate left by 4
        REQUIRE(test.deltaPC() == 4);
    }
}

TEST_CASE("Bit Manipulations", "[cpu][instruction][bit]") {
    InstructionTest test;

    SECTION("BSET") {
        test.setReg(6, 0x00000010);  // x6 = 0b...00010000 (bit 4 set)
        test.setReg(7, 0x00000005);  // x7 = 5 (bit position)
        test.loadInstruction(0x287312B3);   // BSET x5, x6, x7
        test.execute();
        REQUIRE(test.getReg(5) == 0x00000030);  // Result = 0b...00110000 (bits 4 and 5 set)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BCLR") {
        test.setReg(9, 0x000000FF);   // x9 = 0b11111111 (all low bits set)
        test.setReg(10, 0x00000003);  // x10 = 3 (bit position)
        test.loadInstruction(0x48A49433);   // BCLR x8, x9, x10
        test.execute();
        REQUIRE(test.getReg(8) == 0x000000F7);  // Result = 0b11110111 (bit 3 cleared)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BINV") {
        test.setReg(12, 0x00000020);  // x12 = 0b00100000 (bit 5 set)
        test.setReg(13, 0x00000004);  // x13 = 4 (bit position)
        test.loadInstruction(0x68D615B3);   // BINV x11, x12, x13
        test.execute();
        REQUIRE(test.getReg(11) == 0x00000030);  // Result = 0b00110000 (bits 4 and 5 set)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("BEXT") {
        test.setReg(15, 0x00000040);  // x15 = 0b01000000 (bit 6 set)
        test.setReg(16, 0x00000006);  // x16 = 6 (bit position)
        test.loadInstruction(0x4907D733);   // BEXT x14, x15, x16
        test.execute();
        REQUIRE(test.getReg(14) == 0x00000001);  // Result = 1 (bit 6 was set)
        REQUIRE(test.deltaPC() == 4);
    }
}

TEST_CASE("Logical operators", "[cpu][instruction][logical]") {
    InstructionTest test;

    SECTION("XOR") {
        test.setReg(18, 0x000000FF);  // x18 = 0b11111111
        test.setReg(19, 0x000000F0);  // x19 = 0b11110000
        test.loadInstruction(0x013948B3);   // XOR x17, x18, x19
        test.execute();
        REQUIRE(test.getReg(17) == 0x0000000F);  // Result = 0xFF ^ 0xF0 = 0x0F
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("OR") {
        test.setReg(21, 0x000000AA);  // x21 = 0b10101010
        test.setReg(22, 0x00000055);  // x22 = 0b01010101
        test.loadInstruction(0x016AEA33);   // OR x20, x21, x22
        test.execute();
        REQUIRE(test.getReg(20) == 0x000000FF);  // Result = 0xAA | 0x55 = 0xFF
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("AND") {
        test.setReg(24, 0x000000FF);  // x24 = 0b11111111
        test.setReg(25, 0x000000F0);  // x25 = 0b11110000
        test.loadInstruction(0x019C7BB3);   // AND x23, x24, x25
        test.execute();
        REQUIRE(test.getReg(23) == 0x000000F0);  // Result = 0xFF & 0xF0 = 0xF0
        REQUIRE(test.deltaPC() == 4);
    }
}

TEST_CASE("Other operators", "[cpu][instruction][other]") {
    InstructionTest test;

    SECTION("ZEXT.H") {
        test.setReg(27, 0xFFFF8765);  // x27 = 0xFFFF8765 (upper bits set)
        test.loadInstruction(0x080DCD33);   // ZEXT.H x26, x27
        test.execute();
        REQUIRE(test.getReg(26) == 0x00008765);  // Result = lower 16 bits only, upper zeroed
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("PACK") {
        test.setReg(29, 0xAAAA1234);  // x29 = lower 16 bits = 0x1234
        test.setReg(30, 0xBBBB5678);  // x30 = lower 16 bits = 0x5678
        test.loadInstruction(0x09EECE33);   // PACK x28, x29, x30
        test.execute();
        REQUIRE(test.getReg(28) == 0x56781234);  // Result = (0x5678 << 16) | 0x1234
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("PACK.H") {
        test.setReg(2, 0xAAAA00AB);  // x2 = lower 8 bits = 0xAB
        test.setReg(3, 0xBBBB00CD);  // x3 = lower 8 bits = 0xCD
        test.loadInstruction(0x083170B3);   // PACKH x1, x2, x3
        test.execute();
        REQUIRE(test.getReg(1) == 0x0000CDAB);  // Result = (0xCD << 8) | 0xAB
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("MIN") {
        test.setReg(5, 0xFFFFFFFE);  // x5 = -2 (signed)
        test.setReg(6, 0x00000005);  // x6 = 5
        test.loadInstruction(0x0A62C233);   // MIN x4, x5, x6
        test.execute();
        REQUIRE(test.getReg(4) == 0xFFFFFFFE);  // Result = -2 (minimum signed value)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("MINU") {
        test.setReg(8, 0xFFFFFFFE);  // x8 = 4,294,967,294 (unsigned)
        test.setReg(9, 0x00000005);  // x9 = 5
        test.loadInstruction(0x0A9453B3);   // MINU x7, x8, x9
        test.execute();
        REQUIRE(test.getReg(7) == 0x00000005);  // Result = 5 (minimum unsigned value)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("MAX") {
        test.setReg(11, 0xFFFFFFFE);  // x11 = -2 (signed)
        test.setReg(12, 0x00000005);  // x12 = 5
        test.loadInstruction(0x0AC5E533);   // MAX x10, x11, x12
        test.execute();
        REQUIRE(test.getReg(10) == 0x00000005);  // Result = 5 (maximum signed value)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("MAXU") {
        test.setReg(14, 0xFFFFFFFE);  // x14 = 4,294,967,294 (unsigned)
        test.setReg(15, 0x00000005);  // x15 = 5
        test.loadInstruction(0x0AF776B3);   // MAXU x13, x14, x15
        test.execute();
        REQUIRE(test.getReg(13) == 0xFFFFFFFE);  // Result = 4,294,967,294 (maximum unsigned value)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CZERO.EQZ - True") {
        test.setReg(17, 0x12345678);  // x17 = 0x12345678
        test.setReg(18, 0x00000000);  // x18 = 0 (condition true)
        test.loadInstruction(0x0F28D833);   // CZERO.EQZ x16, x17, x18
        test.execute();
        REQUIRE(test.getReg(16) == 0x00000000);  // Result = 0 (because x18 == 0)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CZERO.EQZ - False") {
        test.setReg(17, 0x12345678);  // x17 = 0x12345678
        test.setReg(18, 0x00000001);  // x18 = 1 (condition false)
        test.loadInstruction(0x0F28D833);   // CZERO.EQZ x16, x17, x18
        test.execute();
        REQUIRE(test.getReg(16) == 0x12345678);  // Result = x17 (because x18 != 0)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CZERO.NEZ - True") {
        test.setReg(20, 0x12345678);  // x20 = 0x12345678
        test.setReg(21, 0x00000001);  // x21 = 1 (condition true)
        test.loadInstruction(0x0F5A79B3);   // CZERO.NEZ x19, x20, x21
        test.execute();
        REQUIRE(test.getReg(19) == 0x00000000);  // Result = 0 (because x21 != 0)
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("CZERO.NEZ - False") {
        test.setReg(20, 0x12345678);  // x20 = 0x12345678
        test.setReg(21, 0x00000000);  // x21 = 0 (condition false)
        test.loadInstruction(0x0F5A79B3);   // CZERO.NEZ x19, x20, x21
        test.execute();
        REQUIRE(test.getReg(19) == 0x12345678);  // Result = x20 (because x21 == 0)
        REQUIRE(test.deltaPC() == 4);
    }
}
