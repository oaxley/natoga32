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
 * @brief	Tests - Cartridge validation
 */

#include <catch2/catch_all.hpp>
#include <vector>

#include "cartridge_builder.h"
#include "core/cartridge.h"

using namespace vc;

namespace
{

CartError checkWithFlags(u16 flags) {
    test::CartBuilder b;
    b.headerFlags = flags;
    return checkCartridge(b.build());
}

}

TEST_CASE("Cartridge header checks", "[cartridge]")
{
    SECTION("valid default cartridge") {
        CHECK(checkCartridge(test::CartBuilder{}.build()) == CartError::None);
    }

    SECTION("magic") {
        auto file = test::CartBuilder{}.build();
        CHECK(checkCartridge(std::span<const u8>(file.data(), 15)) == CartError::BadMagic);
        CHECK(checkCartridge(std::span<const u8>()) == CartError::BadMagic);
        for (u32 i = 0; i < 4; ++i) {
            auto bad = file;
            bad[i] ^= 0x01;
            CHECK(checkCartridge(bad) == CartError::BadMagic);
        }
    }

    SECTION("format version") {
        auto file = test::CartBuilder{}.build();
        file[4] = 0x11;
        test::fixHeaderCrc(file);
        CHECK(checkCartridge(file) == CartError::BadVersion);
    }

    SECTION("header CRC") {
        auto file = test::CartBuilder{}.build();
        file[5] = 0x01;
        CHECK(checkCartridge(file) == CartError::HeaderCrc);
    }

    SECTION("flags") {
        CHECK(checkWithFlags(0x0000) == CartError::None);
        CHECK(checkWithFlags(0x0010) == CartError::BadFlags);
        CHECK(checkWithFlags(0x8000) == CartError::BadFlags);
        CHECK(checkWithFlags(0x0002) == CartError::BadFlags);
        CHECK(checkWithFlags(0x0008 | 4) == CartError::BadFlags);
        CHECK(checkWithFlags(0x0008 | 7) == CartError::BadFlags);
        CHECK(checkWithFlags(0x0008 | 3) == CartError::None);
        CHECK(checkWithFlags(0x0008) == CartError::None);
    }

    SECTION("total size") {
        auto file = test::CartBuilder{}.build();

        auto bigger = file;
        test::put32(bigger, 8, static_cast<u32>(file.size()) + 4);
        test::fixHeaderCrc(bigger);
        CHECK(checkCartridge(bigger) == CartError::BadSize);

        auto huge = file;
        test::put32(huge, 8, 0xFFFF'FFFF);
        test::fixHeaderCrc(huge);
        CHECK(checkCartridge(huge) == CartError::BadSize);

        // the size matches, so the header is fine (the trailing bytes fail a later check)
        auto extra = file;
        extra.insert(extra.end(), 4, 0);
        test::put32(extra, 8, static_cast<u32>(extra.size()));
        test::fixHeaderCrc(extra);
        CHECK(checkCartridge(extra) != CartError::BadSize);
    }
}
