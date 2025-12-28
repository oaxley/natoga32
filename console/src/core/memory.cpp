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
 * @brief	Virtual Console - Memory implementation
 */
// standard library headers
#include <cstring>
#include <stdexcept>

// program-specific headers
#include "memory.h"


namespace vc {

Memory::Memory() :
    ram_(8*1024*1024) {
    reset();
}

void Memory::reset() {
    std::memset(ram_.data(), 0, ram_.size());
}

u8 Memory::read_u8(u32 addr) const {
    if (addr < RAM_BASE + ram_.size()) {
        return ram_[addr - RAM_BASE];
    }
    // TODO: Memory Map I/O

    return 0;
}

u16 Memory::read_u16(u32 addr) const {
    return read_u8(addr) | (read_u8(addr + 1) << 8);
}

u32 Memory::read_u32(u32 addr) const {
    return read_u8(addr) |
           (read_u8(addr + 1) << 8) |
           (read_u8(addr + 2) << 16) |
           (read_u8(addr + 3) << 24);
}

} // namespace vc

