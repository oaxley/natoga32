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

#include "instruction_test.h"
#include "core/csr.h"

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
