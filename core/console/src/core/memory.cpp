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

Memory::Memory(size_t ramsize) :
    ram_(ramsize)
{
    reset();
}

void Memory::reset()
{
    // Reset all memory EXCEPT ROM area
    // ROM area: 0x0000'0000 to 0x0000'0000 + MMAP_ROM_SIZE

    // Clear memory from ROM end to the end of RAM
    u32 rom_end = MMAP_ROM_BOOT_LOADER + MMAP_ROM_SIZE;
    if (rom_end < ram_.size()) {
        std::memset(ram_.data() + rom_end, 0, ram_.size() - rom_end);
    }
}

u8 Memory::read8(u32 addr) const
{
    if (!isValidAddress(addr)) {
        return 0;
    }
    return ram_[addr];
}

u16 Memory::read16(u32 addr) const
{
    if (!isValidAddress(addr) || !isValidAddress(addr + 1)) {
        return 0;
    }

    return static_cast<u16>(ram_[addr]) |
           (static_cast<u16>(ram_[addr + 1]) << 8);
}

u32 Memory::read32(u32 addr) const
{
    if (!isValidAddress(addr) || !isValidAddress(addr + 3)) {
        return 0;
    }

    return static_cast<u32>(ram_[addr]) |
           (static_cast<u32>(ram_[addr + 1]) <<  8) |
           (static_cast<u32>(ram_[addr + 2]) << 16) |
           (static_cast<u32>(ram_[addr + 3]) << 24);
}

void Memory::write8(u32 addr, u8 value)
{
    if (!isValidAddress(addr)) {
        return;
    }

    // ensure ROM remains RO
    if (addr < MMAP_ROM_BOOT_LOADER + MMAP_ROM_SIZE - 1) {
        return;
    }
    ram_[addr] = value;
}

void Memory::write16(u32 addr, u16 value)
{
    if (!isValidAddress(addr) || !isValidAddress(addr + 1)) {
        return;
    }
    if (isReadOnly(addr)) {
        return;
    }

    ram_[addr] = static_cast<u8>(value & 0xFF);
    ram_[addr + 1] = static_cast<u8>((value >> 8) & 0xFF);
}

void Memory::write32(u32 addr, u32 value)
{
    if (!isValidAddress(addr) || !isValidAddress(addr + 3)) {
        return;
    }

    ram_[addr] = static_cast<u8>(value & 0xFF);
    ram_[addr + 1] = static_cast<u8>((value >>  8) & 0xFF);
    ram_[addr + 2] = static_cast<u8>((value >> 16) & 0xFF);
    ram_[addr + 3] = static_cast<u8>((value >> 24) & 0xFF);
}

//----- private methods
bool Memory::isValidAddress(u32 addr) const
{
    if (addr < ram_.size()) {
        return true;
    }

    return false;
}

bool Memory::isReadOnly(u32 addr) const
{
    // ROM
    if (isInside(addr, MMAP_ROM_BOOT_LOADER, MMAP_ROM_SIZE)) {
        return true;
    }

    // CartRidge space
    if (isInside(addr, MMAP_CARTRIDGE_IO, MMAP_CARTIO_SIZE)) {
        return true;
    }

    return false;
}

bool Memory::isInside(u32 value, u32 addr, size_t size) const
{
    if ((value >= addr) && (value < (addr + size))) {
        return true;
    }

    return false;
}

} // namespace vc

