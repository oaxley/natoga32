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
 * @brief	Tests - CRC-32 helper
 */

#include <catch2/catch_all.hpp>

#include <span>
#include <string_view>

#include "core/crc32.h"

using namespace vc;

static std::span<const u8> bytes_of(std::string_view s) {
    return {reinterpret_cast<const u8*>(s.data()), s.size()};
}

TEST_CASE("CRC-32", "[crc32]") {
    REQUIRE( crc32(bytes_of("123456789")) == 0xCBF43926u );
    REQUIRE( crc32({}) == 0u );

    SECTION("chained over two spans") {
        u32 partial = crc32(bytes_of("1234"));
        REQUIRE( crc32(bytes_of("56789"), partial) == 0xCBF43926u );
    }
}
