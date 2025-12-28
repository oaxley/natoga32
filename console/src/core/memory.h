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
 * @brief	Virtual Console - Memory header
 */
#pragma once

// standard library headers
#include <vector>
#include <memory>

// program-specific headers
#include "vc/types.h"


namespace vc {

class Memory
{
public:
    Memory();

    // read operations
    u8  read_u8(u32 addr) const;
    u16 read_u16(u32 addr) const;
    u32 read_u32(u32 addr) const;

    // write operations
    void write_u8(u32 addr, u8 value);
    void write_u16(u32 addr, u16 value);
    void write_u32(u32 addr, u32 value);

    // direct access (for DMA, ...)
    u8* get_ram_ptr() { return ram_.data(); }

    void reset();

private:
    std::vector<u8> ram_;

    // memory map
    static constexpr u32 RAM_BASE = 0x0000'0000;
};

} // namespace vc
