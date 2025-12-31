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
 * @brief	RISC-V decoder functions
 */
#pragma once

// program-specific headers
#include "vc/types.h"

namespace vc::decoder {

u32 opcode(u32 instr) {
    return instr & 0x7f;
}

u32 rd(u32 instr) {
    return (instr >> 7) & 0x1F;
}

u32 rs1(u32 instr) {
    return (instr >> 15) & 0x1F;
}

u32 rs2(u32 instr) {
    return (instr >> 20) & 0x1F;
}

u32 funct3(u32 instr) {
    return (instr >> 12) & 0x7;
}

u32 funct7(u32 instr) {
    return (instr >> 25) & 0x7F;
}

i32 immTypeI(u32 instr) {
    return static_cast<i32>(instr & 0xFFF0'0000) >> 20;
}

i32 immTypeS(u32 instr) {
    u32 imm = ((instr >> 7) & 0x1F) | ((instr >> 25) << 5);
    return static_cast<i32>(imm << 20) >> 20;     // sign extended
}

i32 immTypeB(u32 instr) {
    u32 imm = ((instr >>  8) & 0x0F) <<  1 |
              ((instr >> 25) & 0x3F) <<  5 |
              ((instr >>  7) & 0x01) << 11 |
              ((instr >> 31) & 0x01) << 12;
    return static_cast<i32>(imm << 19) >> 19;     // sign extended
}

i32 immTypeU(u32 instr) {
    return static_cast<i32>(instr & 0xFFFF'F000);    // no shift, value is imm[32:12]
}

i32 immTypeJ(u32 instr) {
    u32 imm = ((instr >> 21) & 0x3FF) <<  1 |
              ((instr >> 20) &   0x1) << 11 |
              ((instr >> 12) &  0xFF) << 12 |
              ((instr >> 31) &   0x1) << 20;
    return static_cast<i32>(imm << 11) >> 11;     // sign extendee
}

} // namespace vc::decoder
