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
 * @brief	Virtual Console - Cartridge format constants and validation header
 */
#pragma once

// standard library headers
#include <span>
#include <vector>

// program-specific headers
#include "vc/types.h"
#include "core/constants.h"


namespace vc {

class Memory;

//----- file format constants

constexpr u32 CART_HEADER_SIZE = 16;
constexpr u32 CART_IHDR_SIZE = 64;
constexpr u32 CART_INDX_ENTRY_SIZE = 16;
constexpr u32 CART_INDX_MAX_ENTRIES = 15872;
constexpr u8 CART_FORMAT_VERSION = 0x10;
constexpr u64 CART_MAX_SIZE = 256ull << 20;

//----- control zone addresses

constexpr u32 CART_STATUS = MMAP_CART_CONTROL;
constexpr u32 CART_SIZE = MMAP_CART_CONTROL + 4;
constexpr u32 CART_INDX_COUNT = MMAP_CART_CONTROL + 8;
constexpr u32 CART_HEADER_MIRROR = MMAP_CART_CONTROL + 0x1000;
constexpr u32 CART_IHDR_MIRROR = CART_HEADER_MIRROR + 0x10;
constexpr u32 CART_INDX_MIRROR = MMAP_CART_CONTROL + 0x2000;

//----- status register bits

constexpr u32 CART_STATUS_PRESENT = 1u << 0;
constexpr u32 CART_STATUS_VALID = 1u << 1;
// the error code is stored in bits 8-15

//----- validation

enum class CartError : u8 {
    None = 0,
    BadMagic = 1,
    BadVersion = 2,
    HeaderCrc = 3,
    BadSize = 4,
    BadIhdr = 5,
    IhdrCrc = 6,
    BadIndx = 7,
    IndxCrc = 8,
    IndxTooBig = 9,
    BadIend = 10,
    BadChunk = 11,
    IndxMismatch = 12,
    BadFlags = 13
};

/// Validate a cartridge file; returns the first error found, in check order
CartError checkCartridge(std::span<const u8> file);

//----- cartridge device

class Cartridge
{
public:
    /// Store a copy of the file; it is checked on the next map()
    void insert(std::span<const u8> file);
    /// Take over the file without copying it; it is checked on the next map()
    void insert(std::vector<u8>&& file);
    void eject();

    bool isPresent() const;
    size_t size() const;

    /// Check the file and fill the control zone of `mem` (expected zeroed, i.e. right after Memory::reset())
    void map(Memory& mem);

    /// Result of the check done by the last map()
    bool isValid() const;

    /// Raw file byte; 0 beyond the end or when absent
    u8 read8(u32 offset) const;

    /// Byte seen through a cartridge window; 0 unless the cartridge is valid
    u8 readWindow(u16 bank, u32 window_offset) const;

private:
    std::vector<u8> file_;
    bool present_ = false;
    bool valid_ = false;
};

}
