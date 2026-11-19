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

// program-specific headers
#include "vc/types.h"
#include "core/constants.h"


namespace vc {

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

}
