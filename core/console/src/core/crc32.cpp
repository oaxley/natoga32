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
 * @brief	Virtual Console - CRC-32 helper implementation
 */
// standard library headers
#include <array>

// program-specific headers
#include "crc32.h"


namespace vc {

//----- constants and lookup table

namespace {

constexpr u32 CRC32_POLY = 0xEDB88320u;
constexpr u32 CRC32_XOROUT = 0xFFFFFFFFu;

constexpr std::array<u32, 256> make_table() {
    std::array<u32, 256> table{};
    for (u32 n = 0; n < 256; ++n) {
        u32 c = n;
        for (int k = 0; k < 8; ++k) {
            c = (c & 1) ? (CRC32_POLY ^ (c >> 1)) : (c >> 1);
        }
        table[n] = c;
    }
    return table;
}

constexpr auto CRC32_TABLE = make_table();

}

//----- public functions

// crc = 0 starts from the init state 0xFFFFFFFF, since the state is inverted on entry
u32 crc32(std::span<const u8> data, u32 crc) {
    u32 c = crc ^ CRC32_XOROUT;
    for (u8 b : data) {
        c = CRC32_TABLE[(c ^ b) & 0xFFu] ^ (c >> 8);
    }
    return c ^ CRC32_XOROUT;
}

}
