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
#include "constants.h"
#include "memory.h"


namespace vc {

Memory::Memory() :
    ram_(MMAP_RAM_SIZE) {
    reset();
}

void Memory::reset() {
    std::memset(ram_.data(), 0, ram_.size());
}

u8 Memory::read_u8(u32 addr) const {
    if (addr < ram_.size()) {
        return ram_[addr];
    }
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

void Memory::write_u8(u32 addr, u8 value) {
    // ensure ROM remains RO
    if (addr < MMAP_ROM_BOOT_LOADER + MMAP_ROM_SIZE - 1) {
        return;
    }
    ram_[addr] = value;
}

void Memory::write_u16(u32 addr, u16 value) {
    write_u8(addr, value & 0xFF);
    write_u8(addr + 1, (value >> 8) & 0xFF);
}

void Memory::write_u32(u32 addr, u32 value) {
    write_u8(addr, value & 0xFF);
    write_u8(addr+1, (value >> 8) & 0xFF);
    write_u8(addr+2, (value >> 16) & 0xFF);
    write_u8(addr+3, (value >> 24) & 0xFF);
}

} // namespace vc

