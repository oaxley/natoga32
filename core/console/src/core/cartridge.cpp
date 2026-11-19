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
 * @brief	Virtual Console - Cartridge validation implementation
 */
// standard library headers
#include <cstring>

// program-specific headers
#include "cartridge.h"
#include "crc32.h"


namespace vc {

//----- helpers

namespace {

constexpr u8 CART_MAGIC[4] = { 0x7F, 'N', '3', '2' };

constexpr u32 HEADER_OFF_VERSION = 4;
constexpr u32 HEADER_OFF_FLAGS = 6;
constexpr u32 HEADER_OFF_SIZE = 8;
constexpr u32 HEADER_OFF_CRC = 12;

constexpr u16 FLAGS_MODE_MASK = 0x0007;
constexpr u16 FLAGS_MODE_REQUIRED = 0x0008;
constexpr u16 FLAGS_RESERVED_MASK = 0xFFF0;
constexpr u16 FLAGS_MODE_MAX = 3;

u16 read16(std::span<const u8> d, u32 off) {
    return static_cast<u16>(d[off] | (d[off + 1] << 8));
}

u32 read32(std::span<const u8> d, u32 off) {
    return static_cast<u32>(d[off]) | (static_cast<u32>(d[off + 1]) << 8) |
           (static_cast<u32>(d[off + 2]) << 16) | (static_cast<u32>(d[off + 3]) << 24);
}

}

//----- public functions

CartError checkCartridge(std::span<const u8> file) {
    if (file.size() < CART_HEADER_SIZE || std::memcmp(file.data(), CART_MAGIC, sizeof(CART_MAGIC)) != 0) {
        return CartError::BadMagic;
    }

    if (file[HEADER_OFF_VERSION] != CART_FORMAT_VERSION) {
        return CartError::BadVersion;
    }

    if (crc32(file.first(HEADER_OFF_CRC)) != read32(file, HEADER_OFF_CRC)) {
        return CartError::HeaderCrc;
    }

    const u16 flags = read16(file, HEADER_OFF_FLAGS);
    const u16 mode = flags & FLAGS_MODE_MASK;
    if ((flags & FLAGS_RESERVED_MASK) != 0 || mode > FLAGS_MODE_MAX ||
        (mode != 0 && (flags & FLAGS_MODE_REQUIRED) == 0)) {
        return CartError::BadFlags;
    }

    const u32 size = read32(file, HEADER_OFF_SIZE);
    if (size != file.size() || size > CART_MAX_SIZE) {
        return CartError::BadSize;
    }

    return CartError::None;
}

}
