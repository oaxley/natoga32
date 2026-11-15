#include <catch2/catch_all.hpp>

#include "instruction_test.h"
#include "core/csr.h"
#include "vc/exceptions.h"

using namespace vc;

// trap handler address used by the tests
constexpr u32 TRAP_HANDLER = 0x0000'0200;

static u32 csr(InstructionTest& test, u16 id)
{
    return test.mem.read32(MMAP_CSR_REGISTERS + id * sizeof(u32));
}

// run one load/store on addr (in x1) and return the address of the instruction
static u32 runAccess(InstructionTest& test, u32 instr, u32 addr, u32 x2 = 0)
{
    test.mem.write32(MMAP_CSR_REGISTERS + CSR_MTVEC * sizeof(u32), TRAP_HANDLER);
    test.setReg(1, addr);
    test.setReg(2, x2);
    test.loadInstruction(instr);
    u32 instr_addr = test.getPC();
    test.execute();
    return instr_addr;
}

TEST_CASE("Private thread zone is protected from programs", "[cpu][memory][protection]") {
    InstructionTest test;

    SECTION("LW at the start of the zone raises a load access fault") {
        test.setReg(3, 0x1234);
        u32 instr_addr = runAccess(test, 0x0000A183, 0x0020'5000);      // LW x3, 0(x1)
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_LOAD_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0020'5000);
        REQUIRE(csr(test, CSR_MEPC) == instr_addr);
        REQUIRE(test.getPC() == TRAP_HANDLER);
        REQUIRE(test.getReg(3) == 0x1234);
    }

    SECTION("LB on the last byte of the zone raises a load access fault") {
        runAccess(test, 0x00008183, 0x0020'5FFF);                       // LB x3, 0(x1)
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_LOAD_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0020'5FFF);
    }

    SECTION("SW in the Thread Control Blocks raises a store access fault") {
        u32 before = test.mem.read32(0x0020'5400);
        u32 instr_addr = runAccess(test, 0x0020A023, 0x0020'5400, 0xDEAD'BEEF);  // SW x2, 0(x1)
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_STORE_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0020'5400);
        REQUIRE(csr(test, CSR_MEPC) == instr_addr);
        REQUIRE(test.getPC() == TRAP_HANDLER);
        REQUIRE(test.mem.read32(0x0020'5400) == before);
    }

    SECTION("SB in the reserved part raises a store access fault") {
        runAccess(test, 0x00208023, 0x0020'5800, 0xAB);                 // SB x2, 0(x1)
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_STORE_ACCESS_FAULT);
        REQUIRE(test.mem.read8(0x0020'5800) == 0);
    }

    SECTION("an access straddling the lower edge faults") {
        runAccess(test, 0x0020A023, 0x0020'4FFE, 0xDEAD'BEEF);          // SW x2, 0(x1)
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_STORE_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0020'4FFE);
    }

    SECTION("a halfword straddling the lower edge faults") {
        runAccess(test, 0x00009183, 0x0020'4FFF);                       // LH x3, 0(x1)
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_LOAD_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0020'4FFF);
    }

    SECTION("a faulting load with rd == rs1 keeps rs1") {
        runAccess(test, 0x0000A083, 0x0020'5000);                       // LW x1, 0(x1)
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_LOAD_ACCESS_FAULT);
        REQUIRE(test.getReg(1) == 0x0020'5000);
    }

    SECTION("a load just below the zone works") {
        test.mem.write32(0x0020'4FFC, 0xCAFE'F00D);
        runAccess(test, 0x0000A183, 0x0020'4FFC);                       // LW x3, 0(x1)
        REQUIRE(test.deltaPC() == 4);
        REQUIRE(test.getReg(3) == 0xCAFE'F00D);
    }

    SECTION("a store just above the zone (TLS) works") {
        runAccess(test, 0x0020A023, 0x0020'6000, 0x1234'5678);          // SW x2, 0(x1)
        REQUIRE(test.deltaPC() == 4);
        REQUIRE(test.mem.read32(0x0020'6000) == 0x1234'5678);
    }
}

TEST_CASE("Invalid load/store encodings are illegal instructions", "[cpu][memory][illegal]") {
    InstructionTest test;

    // unused funct3 values, both far from and near the private zone
    const u32 addr = GENERATE(as<u32>{}, MMAP_DATA, 0x0020'4FFC);

    SECTION("load with funct3 = 011") {
        runAccess(test, 0x0000B183, addr);                              // funct3 011, x3, 0(x1)
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_MAIN_ILLEGAL_INSTRUCTION);
        REQUIRE(test.getPC() == TRAP_HANDLER);
    }

    SECTION("load with funct3 = 110") {
        runAccess(test, 0x0000E183, addr);                              // funct3 110
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_MAIN_ILLEGAL_INSTRUCTION);
        REQUIRE(test.getPC() == TRAP_HANDLER);
    }

    SECTION("store with funct3 = 011") {
        runAccess(test, 0x0020B023, addr, 0xDEAD'BEEF);                 // funct3 011
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_MAIN_ILLEGAL_INSTRUCTION);
        REQUIRE(test.getPC() == TRAP_HANDLER);
    }

    SECTION("store with funct3 = 100") {
        runAccess(test, 0x0020C023, addr, 0xDEAD'BEEF);                 // funct3 100
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_MAIN_ILLEGAL_INSTRUCTION);
        REQUIRE(test.mem.read32(addr) == 0);
    }
}
