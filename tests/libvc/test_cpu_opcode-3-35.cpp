#include <catch2/catch_all.hpp>

#include "core/memory.h"
#include "core/cpu.h"

using namespace vc;

TEST_CASE("CPU Load/Store Instructions", "[cpu][instructions][memory]") {
    Memory mem;
    CPU cpu(mem);

    SECTION("SB and LB/LBU") {
        auto& thread = cpu.getThreadContext(0);

        u8 value = 0xAB;
        u32 addr = MMAP_DATA;
        thread.registers[1] = addr;
        thread.registers[2] = value;

        // SB x2, 0(x1)
        u32 instr = 0x00208023;
        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(mem.read8(addr) == value);
        REQUIRE((thread.pc - MMAP_CODE) == 4);


        // LB x3, 0(x1) - positive value
        mem.write32(addr, 0x12);
        instr = 0x00008183;
        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(thread.registers[3] == 0x12);
        REQUIRE((thread.registers[3] & 0x8000'0000) == 0);
        REQUIRE((thread.pc - MMAP_CODE) == 4);

        // LB x3, 0(x1) - negative value
        mem.write32(addr, static_cast<u32>(-34));
        instr = 0x00008183;
        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(static_cast<i8>(thread.registers[3]) == -34);
        REQUIRE((thread.registers[3] & 0x8000'0000) == 0x8000'0000);
        REQUIRE((thread.pc - MMAP_CODE) == 4);


        // LBU x3, 0(x1) - u8 value > 128
        mem.write32(addr, 171);
        instr = 0x0000c183;
        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(thread.registers[3] == 171);
        REQUIRE((thread.registers[3] & 0x8000'0000) == 0);
        REQUIRE((thread.pc - MMAP_CODE) == 4);
    }

    SECTION("SH and LH") {
        auto& thread = cpu.getThreadContext(0);

        u16 value = 0x1234;
        u32 addr = MMAP_DATA;
        thread.registers[1] = addr;
        thread.registers[2] = value;

        // SH x2, 0(x1)
        u32 instr = 0x00209023;
        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(mem.read16(addr) == value);
        REQUIRE((thread.pc - MMAP_CODE) == 4);


        // LH x3, 0(x1) - positive value
        mem.write32(addr, 16384);

        instr = 0x00009183;
        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(thread.registers[3] == 16384);
        REQUIRE((thread.registers[3] & 0x8000'0000) == 0);
        REQUIRE((thread.pc - MMAP_CODE) == 4);

        // LH x3, 0(x1) - negative value
        mem.write32(addr, static_cast<u32>(-16384));

        instr = 0x00009183;
        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(static_cast<i16>(thread.registers[3]) == -16384);
        REQUIRE((thread.registers[3] & 0x8000'0000) == 0x8000'0000);
        REQUIRE((thread.pc - MMAP_CODE) == 4);

        // LHU x3, 0(x1) - unsigned positive value
        mem.write32(addr, 48000);

        instr = 0x0000d183;
        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(thread.registers[3] == 48000);
        REQUIRE((thread.registers[3] & 0x8000'0000) == 0);
        REQUIRE((thread.pc - MMAP_CODE) == 4);
    }

    SECTION("SW and LW") {
        auto& thread = cpu.getThreadContext(0);

        u32 value = 0xCAFEBABE;
        u32 addr = MMAP_DATA;
        thread.registers[1] = addr;
        thread.registers[2] = value;

        // SW x2, 0(x1)
        u32 instr = 0x0020a023;
        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(mem.read32(addr) == value);
        REQUIRE((thread.pc - MMAP_CODE) == 4);

        // LW x3, 0(x1)
        instr = 0x0000a183;

        mem.write32(MMAP_CODE, instr);

        thread.pc = MMAP_CODE;
        cpu.tick();

        REQUIRE(thread.registers[3] == value);
        REQUIRE((thread.pc - MMAP_CODE) == 4);
    }
}
