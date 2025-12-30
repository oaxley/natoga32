#include <catch2/catch_all.hpp>

#include "core/constants.h"
#include "core/memory.h"

using namespace vc;

constexpr u32 MEM_1 = MMAP_INTERRUPT_VECTOR;
constexpr u32 MEM_2 = MMAP_VIDEO_RAM;

TEST_CASE("Memory Initialization", "[memory]") {
    Memory mem;

    SECTION("Memory starts zeroed") {
        // check only the first Kilo
        for (u32 i = 0; i < 1024; i++) {
            REQUIRE(mem.read_u8(i) == 0);
        }
    }
}

TEST_CASE("Memory read/write u8", "[memory]") {
    Memory mem;

    SECTION("Write and read byte") {
        mem.write_u8(MEM_1, 0x42);
        REQUIRE(mem.read_u8(MEM_1) == 0x42);
    }

    SECTION("Multiple writes") {
        mem.write_u8(MEM_2 + 0, 0xAB);
        mem.write_u8(MEM_2 + 1, 0xCD);
        mem.write_u8(MEM_2 + 2, 0xEF);

        REQUIRE(mem.read_u8(MEM_2 + 0) == 0xAB);
        REQUIRE(mem.read_u8(MEM_2 + 1) == 0xCD);
        REQUIRE(mem.read_u8(MEM_2 + 2) == 0xEF);
    }
}

TEST_CASE("Memory read/write u16", "[memory]") {
    Memory mem;

    SECTION("Write and read halfword") {
        mem.write_u16(MEM_1, 0x1234);
        REQUIRE(mem.read_u16(MEM_1) == 0x1234);
    }

    SECTION("Little-endian byte order") {
        mem.write_u16(MEM_2, 0xABCD);
        REQUIRE(mem.read_u8(MEM_2 + 0) == 0xCD);
        REQUIRE(mem.read_u8(MEM_2 + 1) == 0xAB);
    }
}

TEST_CASE("Memory read/write u32", "[memory]") {
    Memory mem;

    SECTION("Write and read word") {
        mem.write_u32(MEM_1, 0xBAADC0DE);
        REQUIRE(mem.read_u32(MEM_1) == 0xBAADC0DE);
    }

    SECTION("Little-endian byte order") {
        mem.write_u32(MEM_2, 0x12345678);
        REQUIRE(mem.read_u8(MEM_2 + 0) == 0x78);
        REQUIRE(mem.read_u8(MEM_2 + 1) == 0x56);
        REQUIRE(mem.read_u8(MEM_2 + 2) == 0x34);
        REQUIRE(mem.read_u8(MEM_2 + 3) == 0x12);
    }
}

TEST_CASE("Memory reset", "[memory]") {
    Memory mem;

    mem.write_u32(MEM_1, 0xBAADC0DE);
    mem.write_u32(MEM_2, 0xCAFEBABE);

    mem.reset();

    REQUIRE(mem.read_u32(MEM_1) == 0);
    REQUIRE(mem.read_u32(MEM_2) == 0);
}

TEST_CASE("Memory ROM is RO", "[memory]") {
    Memory mem;

    constexpr u32 addr = 0x100;

    // hack to get a view on the ROM
    auto rom = mem.memview<u32>(addr, 1);
    rom[0] = 0xDEADC0DE;

    // try to overwrite the value
    mem.write_u32(addr, 0xBAADC0DE);
    REQUIRE(mem.read_u32(addr) == rom[0]);
}
