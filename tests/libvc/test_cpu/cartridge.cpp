/* -*- coding: utf-8 -*-
 * vim: filetype=cpp
 *
 * This source file is subject to the MIT License
 * that is bundled with this package in the file LICENSE.txt.
 * It is also available through the Internet at this address:
 * https://opensource.org/license/mit
 *
 * @author	Sebastien LEGRAND
 * @license	MIT License
 *
 * @brief	Tests - CPU cartridge bank CSR
 */

#include <catch2/catch_all.hpp>

#include "cartridge_builder.h"
#include "instruction_test.h"
#include "core/cartridge.h"
#include "core/csr.h"
#include "vc/exceptions.h"

using namespace vc;

constexpr u32 TCB0_CARTBANK = 0x0020'5400 + 0x0C;

TEST_CASE("cartbank CSR", "[cpu][cartridge]") {
    InstructionTest test;

    SECTION("is 0 after reset") {
        REQUIRE(test.getCtx().getCartBank() == 0);
        REQUIRE(test.mem.read32(TCB0_CARTBANK) == 0);
    }

    SECTION("a write is stored in the TCB, not in the CSR area") {
        test.setReg(1, 0x1234);
        test.loadInstruction(0x7C009073);       // CSRRW x0, 0x7C0, x1
        test.execute();

        REQUIRE(test.getCtx().getCartBank() == 0x1234);
        REQUIRE(test.mem.read32(TCB0_CARTBANK) == 0x1234);
        REQUIRE(test.mem.read32(MMAP_CSR_REGISTERS + CSR_CARTBANK * sizeof(u32)) == 0);
    }

    SECTION("a write is masked to 16 bits") {
        test.setReg(1, 0x0001'ABCD);
        test.loadInstruction(0x7C009073);       // CSRRW x0, 0x7C0, x1
        test.execute();

        REQUIRE(test.getCtx().getCartBank() == 0xABCD);
        REQUIRE(test.mem.read32(TCB0_CARTBANK) == 0xABCD);
    }

    SECTION("a read returns the bank of the TCB") {
        test.getCtx().setCartBank(0x42);
        test.loadInstruction(0x7C0021F3);       // CSRRS x3, 0x7C0, x0
        test.execute();

        REQUIRE(test.getReg(3) == 0x42);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("is per thread") {
        test.getCtx(1).setCartBank(7);

        REQUIRE(test.getCtx(0).getCartBank() == 0);
        REQUIRE(test.getCtx(1).getCartBank() == 7);
    }

    SECTION("new.t resets the bank of a recycled TCB") {
        u32 tcb = MMAP_THREADS_TCB_BASE + 1 * MMAP_THREAD_TCB_SIZE;
        test.getCtx(1).setCartBank(5);
        REQUIRE(test.mem.read32(tcb + 0x0C) == 5);

        test.setReg(3, MMAP_CODE + 0x100);
        test.loadInstruction(0x0001810b);       // NEW.T x2, x3
        test.execute();

        REQUIRE(test.getReg(2) == 1);
        REQUIRE(test.getCtx(1).getCartBank() == 0);
        REQUIRE(test.mem.read32(tcb + 0x0C) == 0);
    }
}

//-----

namespace
{

constexpr u32 TRAP_HANDLER = 0x0000'0200;

constexpr u32 LW_X3_X1 = 0x0000A183;        // LW x3, 0(x1)
constexpr u32 LB_X3_X1 = 0x00008183;        // LB x3, 0(x1)
constexpr u32 LBU_X3_X1 = 0x0000C183;       // LBU x3, 0(x1)
constexpr u32 SW_X2_X1 = 0x0020A023;        // SW x2, 0(x1)

u32 csr(InstructionTest& test, u16 id)
{
    return test.mem.read32(MMAP_CSR_REGISTERS + id * sizeof(u32));
}

/// Cartridge with one 8 KiB TILS chunk holding a counting pattern
std::vector<u8> cartridgeFile()
{
    std::vector<u8> pattern(8192);
    for (size_t i = 0; i < pattern.size(); ++i) {
        pattern[i] = static_cast<u8>(i);
    }

    test::CartBuilder builder;
    builder.chunks.push_back({ "TILS", pattern, 0 });
    return builder.build();
}

/// InstructionTest with a mapped cartridge and a trap handler
struct CartridgeCpuTest : InstructionTest
{
    std::vector<u8> file = cartridgeFile();
    Cartridge cart;

    CartridgeCpuTest()
    {
        cart.insert(file);
        cart.map(mem);
        cpu.setCartridge(&cart);
        mem.write32(MMAP_CSR_REGISTERS + CSR_MTVEC * sizeof(u32), TRAP_HANDLER);
    }

    /// run one instruction with x1 = addr and x2 = x2_value, return the address of the instruction
    u32 run(u32 instr, u32 addr, u32 x2_value = 0)
    {
        setReg(1, addr);
        setReg(2, x2_value);
        loadInstruction(instr);
        u32 instr_addr = getPC();
        execute();
        return instr_addr;
    }
};

}

TEST_CASE("Cartridge zones from the CPU", "[cpu][cartridge]") {
    CartridgeCpuTest test;

    SECTION("LW in the window reads the start of the file") {
        test.run(LW_X3_X1, MMAP_CART_WINDOW);

        REQUIRE(test.getReg(3) == 0x32334E7F);
        REQUIRE(test.deltaPC() == 4);
    }

    SECTION("LB sign-extends and LBU zero-extends a window byte") {
        u32 tils_data = test::chunkOffset(test.file, 3) + 8;
        REQUIRE(test.file[tils_data + 0x90] == 0x90);
        u32 addr = MMAP_CART_WINDOW + tils_data + 0x90;

        test.run(LB_X3_X1, addr);
        REQUIRE(test.getReg(3) == 0xFFFF'FF90);

        test.run(LBU_X3_X1, addr);
        REQUIRE(test.getReg(3) == 0x90);
    }

    SECTION("LH reads two bytes in little-endian order") {
        test.run(0x00009183, MMAP_CART_WINDOW);                 // LH x3, 0(x1)
        REQUIRE(test.getReg(3) == 0x4E7F);
    }

    SECTION("the window follows the bank of the thread") {
        test.getCtx().setCartBank(1);
        test.run(LW_X3_X1, MMAP_CART_WINDOW);

        REQUIRE(test.getReg(3) == test::get32(test.file, 0x1000));
    }

    SECTION("each thread reads through its own bank") {
        // thread 0 (bank 0) reads, then creates thread 1 and gives it up
        test.setReg(1, MMAP_CART_WINDOW);
        test.setReg(3, MMAP_CODE + 0x100);
        test.loadInstruction(0x0000A283);       // LW x5, 0(x1)
        test.loadInstruction(0x0001810b);       // NEW.T x2, x3
        test.loadInstruction(0x0000100b);       // YIELD.T
        test.execute(3);
        REQUIRE(test.cpu.getCurrentThreadID() == 1);

        // thread 1 selects bank 1 by itself, then reads
        test.setReg(1, MMAP_CART_WINDOW, 1);
        test.setReg(4, 1, 1);
        test.loadInstruction(0x7C021073, 0x100, 1);       // CSRRW x0, 0x7C0, x4
        test.loadInstruction(LW_X3_X1, 0x100, 1);
        test.execute(2);

        REQUIRE(test.getReg(5, 0) == 0x32334E7F);
        REQUIRE(test.getCtx(0).getCartBank() == 0);
        REQUIRE(test.getCtx(1).getCartBank() == 1);
        REQUIRE(test.getReg(3, 1) == test::get32(test.file, 0x1000));
    }

    SECTION("LW at CART_STATUS reads the control zone") {
        test.run(LW_X3_X1, CART_STATUS);
        REQUIRE(test.getReg(3) == (CART_STATUS_PRESENT | CART_STATUS_VALID));
    }

    SECTION("the window reads 0 without a cartridge") {
        test.cpu.setCartridge(nullptr);
        test.setReg(3, 0x1234);
        test.run(LW_X3_X1, MMAP_CART_WINDOW);

        REQUIRE(test.getReg(3) == 0);
        REQUIRE(test.deltaPC() == 4);
        REQUIRE(csr(test, CSR_MCAUSE) == 0);
    }

    SECTION("SW in the window raises a store access fault") {
        u32 instr_addr = test.run(SW_X2_X1, MMAP_CART_WINDOW, 0xDEAD'BEEF);

        REQUIRE(csr(test, CSR_MCAUSE) == CPU_STORE_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == MMAP_CART_WINDOW);
        REQUIRE(csr(test, CSR_MEPC) == instr_addr);
        REQUIRE(test.getPC() == TRAP_HANDLER);
    }

    SECTION("SW on CART_STATUS raises a store access fault and keeps the status") {
        test.run(SW_X2_X1, CART_STATUS, 0xDEAD'BEEF);

        REQUIRE(csr(test, CSR_MCAUSE) == CPU_STORE_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == CART_STATUS);
        REQUIRE(test.mem.read32(CART_STATUS) == 3);
    }

    SECTION("SW straddling the start of the control zone faults") {
        test.run(SW_X2_X1, MMAP_CART_CONTROL - 2, 0xDEAD'BEEF);
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_STORE_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == MMAP_CART_CONTROL - 2);
    }

    SECTION("SW straddling the end of the control zone faults") {
        test.run(SW_X2_X1, 0x0021'FFFE, 0xDEAD'BEEF);
        REQUIRE(csr(test, CSR_MCAUSE) == CPU_STORE_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0021'FFFE);
    }

    SECTION("LW straddling the end of the window raises a load access fault") {
        test.setReg(3, 0x1234);
        test.run(LW_X3_X1, 0x0037'FFFE);

        REQUIRE(csr(test, CSR_MCAUSE) == CPU_LOAD_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0037'FFFE);
        REQUIRE(test.getReg(3) == 0x1234);
        REQUIRE(test.getPC() == TRAP_HANDLER);
    }

    SECTION("LW straddling the start of the window raises a load access fault") {
        test.run(LW_X3_X1, 0x0035'FFFE);

        REQUIRE(csr(test, CSR_MCAUSE) == CPU_LOAD_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0035'FFFE);
    }

    SECTION("LW straddling the end of the control zone raises a load access fault") {
        test.run(LW_X3_X1, 0x0021'FFFE);

        REQUIRE(csr(test, CSR_MCAUSE) == CPU_LOAD_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0021'FFFE);
    }

    SECTION("LW on the last word of the window is allowed") {
        test.run(LW_X3_X1, 0x0037'FFFC);

        REQUIRE(test.getReg(3) == 0);
        REQUIRE(csr(test, CSR_MCAUSE) == 0);
    }

    SECTION("a fetch from the window raises an instruction access fault") {
        test.getCtx().setPC(MMAP_CART_WINDOW);
        test.execute();

        REQUIRE(csr(test, CSR_MCAUSE) == CPU_INSTRUCTION_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == MMAP_CART_WINDOW);
        REQUIRE(csr(test, CSR_MEPC) == MMAP_CART_WINDOW);
        REQUIRE(test.getPC() == TRAP_HANDLER);
    }

    SECTION("a fetch from the control zone raises an instruction access fault") {
        test.getCtx().setPC(0x0022'1000);
        test.execute();

        REQUIRE(csr(test, CSR_MCAUSE) == CPU_INSTRUCTION_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == 0x0022'1000);
        REQUIRE(csr(test, CSR_MEPC) == 0x0022'1000);
        REQUIRE(test.getPC() == TRAP_HANDLER);
    }

    SECTION("a fetch overlapping the zone raises an instruction access fault") {
        test.getCtx().setPC(MMAP_CART_CONTROL - 2);
        test.execute();

        REQUIRE(csr(test, CSR_MCAUSE) == CPU_INSTRUCTION_ACCESS_FAULT);
        REQUIRE(csr(test, CSR_MTVAL) == MMAP_CART_CONTROL - 2);
    }
}
