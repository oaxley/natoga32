#include <catch2/catch_all.hpp>

#include "core/decoders.h"

using namespace vc;

TEST_CASE("Instruction decoder", "[cpu][decode]") {

    SECTION("Decode opcode") {
        // ADDI x1, x0, 42
        u32 instruction = 0x02A00093;
        REQUIRE(decoder::opcode(instruction) == 0x13);
    }

    SECTION("Decode rd") {
        // ADDI x2, x1, 10
        u32 instruction = 0x00A08113;
        REQUIRE(decoder::rd(instruction) == 2);
    }

    SECTION("Decode rs1") {
        // ADDI x2, x1, 10
        u32 instruction = 0x00A08113;
        REQUIRE(decoder::rs1(instruction) == 1);
    }

    SECTION("Decode rs2") {
        // ADDI x3, x1, x2
        u32 instruction = 0x002081B3;
        REQUIRE(decoder::rs2(instruction) == 2);
    }

    SECTION("Decode funct3") {
        // AND x6, x4, x5
        u32 instruction = 0x00527333;
        REQUIRE(decoder::funct3(instruction) == 7);
    }

    SECTION("Decode funct7") {
        // SUB x1, x2, x3
        u32 instruction = 0x402081B3;
        REQUIRE(decoder::funct7(instruction) == 0b00100000);
    }
}

TEST_CASE("S-Type immedidate decoding", "[cpu][decode]") {

    SECTION("Positive immediate: 20") {
        // SW x5, 20(x10)
        u32 instruction = 0x00552A23;
        REQUIRE(decoder::immTypeS(instruction) == 20);
    }

    SECTION("Negative immediate: -8") {
        // SW x12, -8(x8)
        u32 instruction = 0xFEC42C23;
        REQUIRE(decoder::immTypeS(instruction) == -8);
    }

    SECTION("Maximum positive immediate: 2047") {
        u32 instruction = 0x7E000FA3;
        REQUIRE(decoder::immTypeS(instruction) == 2047);
    }

    SECTION("Maximum negative immediate: -2048") {
        u32 instruction = 0x80000023;
        REQUIRE(decoder::immTypeS(instruction) == -2048);
    }
}

TEST_CASE("I-Type immedidate decoding", "[cpu][decode]") {

    SECTION("Positive immediate: 42") {
        // ADDI x5, x10, 42
        u32 instruction = 0x02A50293;
        REQUIRE(decoder::immTypeI(instruction) == 42);
    }

    SECTION("Negative immediate: -5") {
        // ADDI x1, x2, -5
        u32 instruction = 0xFFB10093;
        REQUIRE(decoder::immTypeI(instruction) == -5);
    }

    SECTION("Maximum positive immediate: 2047") {
        u32 instruction = 0x7FF00093;
        REQUIRE(decoder::immTypeI(instruction) == 2047);
    }

    SECTION("Maximum negative immediate: -2048") {
        u32 instruction = 0x80000093;
        REQUIRE(decoder::immTypeI(instruction) == -2048);
    }
}

TEST_CASE("B-Type immedidate decoding", "[cpu][decode]") {

    SECTION("Positive immediate: 100") {
        // BEQ x1, x2, 100
        u32 instruction = 0x06208263;
        REQUIRE(decoder::immTypeB(instruction) == 100);
    }

    SECTION("Negative immediate: -8") {
        // BNE x5, x6, -8
        u32 instruction = 0xFE629CE3;
        REQUIRE(decoder::immTypeB(instruction) == -8);
    }

    SECTION("Maximum positive immediate: 4094") {
        // BEQ x0, x0, 4094
        u32 instruction = 0x7E000FE3;
        REQUIRE(decoder::immTypeB(instruction) == 4094);
    }

    SECTION("Maximum negative immediate: -4096") {
        // BEQ x0, x0, -4096
        u32 instruction = 0x80000063;
        REQUIRE(decoder::immTypeB(instruction) == -4096);
    }
}

TEST_CASE("U-Type immedidate decoding", "[cpu][decode]") {

    SECTION("Positive immediate: 0x12345000") {
        // LUI x5, 0x12345
        u32 instruction = 0x123452B7;
        REQUIRE(decoder::immTypeU(instruction) == (i32)0x12345000);
    }

    SECTION("Negative immediate: 0xFFFFF000") {
        // LUI x6, 0xFFFFF
        u32 instruction = 0xFFFFF337;
        REQUIRE(decoder::immTypeU(instruction) == (i32)0xFFFFF000);
    }

    SECTION("Maximum positive immediate: 0x7FFFF000") {
        // LUI x7, 0x7FFFF
        u32 instruction = 0x7FFFF3B7;
        REQUIRE(decoder::immTypeU(instruction) == (i32)0x7FFFF000);
    }

    SECTION("Maximum negative immediate: 0x80000000") {
        // LUI x8, 0x80000
        u32 instruction = 0x80000437;
        REQUIRE(decoder::immTypeU(instruction) == (i32)0x80000000);
    }
}

TEST_CASE("J-Type immedidate decoding", "[cpu][decode]") {

    SECTION("Positive immediate: 1024") {
        // JAL x1, 1024
        u32 instruction = 0x400000EF;
        REQUIRE(decoder::immTypeJ(instruction) == 1024);
    }

    SECTION("Negative immediate: -512") {
        // JAL x1, -512
        u32 instruction = 0xE01FF0EF;
        REQUIRE(decoder::immTypeJ(instruction) == -512);
    }

    SECTION("Maximum positive immediate: +1048574") {
        // JAL x1, 1048574
        u32 instruction = 0x7FFFF0EF;
        REQUIRE(decoder::immTypeJ(instruction) == 1048574);
    }

    SECTION("Maximum negative immediate: -1048576") {
        // JAL x1, -1048576
        u32 instruction = 0x800000EF;
        REQUIRE(decoder::immTypeJ(instruction) == -1048576);
    }
}
